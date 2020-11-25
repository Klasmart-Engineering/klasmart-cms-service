package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidParentFolderId  = errors.New("invalid parent folder id")
	ErrEmptyFolderName        = errors.New("empty folder name")
	ErrEmptyFolderID          = errors.New("empty folder id")
	ErrEmptyItemID            = errors.New("empty item id")
	ErrMoveToNotFolder            = errors.New("move to an item not folder")
	ErrInvalidFolderOwnerType = errors.New("invalid folder owner type")
	ErrInvalidPartition = errors.New("invalid partition")
	ErrInvalidFolderItemType  = errors.New("invalid folder item type")
	ErrDuplicateFolderName    = errors.New("duplicate folder name in path")
	ErrDuplicateItem          = errors.New("duplicate item in path")
	ErrFolderIsNotEmpty       = errors.New("folder is not empty")
	ErrInvalidItemLink        = errors.New("invalid item link")
	ErrUpdateFolderFailed     = errors.New("update folder into data access failed")
	ErrFolderItemPathError    = errors.New("folder item path error")
	ErrNoFolder = errors.New("no folder")
	ErrMoveRootFolder            = errors.New("move root folder")
	ErrMoveToChild            = errors.New("move to child folder")
)

type IFolderModel interface {
	//创建Folder
	CreateFolder(ctx context.Context, req entity.CreateFolderRequest, operator *entity.Operator) (string, error)
	//添加一个item
	AddItem(ctx context.Context, req entity.CreateFolderItemRequest, operator *entity.Operator) (string, error)
	//删除一批item
	RemoveItem(ctx context.Context, fid string, operator *entity.Operator) error

	//修改Folder
	UpdateFolder(ctx context.Context, folderID string, d entity.UpdateFolderRequest, operator *entity.Operator) error

	//移动item
	MoveItem(ctx context.Context, fid string, distFolder string, operator *entity.Operator) error

	//移动item
	MoveItemBulk(ctx context.Context, fid []string, distFolder string, operator *entity.Operator) error

	//列出Folder下的所有item
	ListItems(ctx context.Context, folderID string, itemType entity.ItemType, operator *entity.Operator) ([]*entity.FolderItem, error)
	//查询Folder
	SearchFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItem, error)
	SearchPrivateFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItem, error)
	SearchOrgFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItem, error)

	//获取Folder
	GetFolderByID(ctx context.Context, folderID string, operator *entity.Operator) (*entity.FolderItemInfo, error)

	GetRootFolder(ctx context.Context, partition entity.FolderPartition, ownerType entity.OwnerType, operator *entity.Operator) (*entity.FolderItem, error)

	//内部API，修改Folder的Visibility Settings
	AddOrUpdateOrgFolderItem(ctx context.Context, tx *dbo.DBContext, partition entity.FolderPartition, link string, operator *entity.Operator) error
	RemoveItemByLink(ctx context.Context, tx *dbo.DBContext, ownerType entity.OwnerType, owner string, link string) error
}

type FolderModel struct{}

func (f *FolderModel) CreateFolder(ctx context.Context, req entity.CreateFolderRequest, operator *entity.Operator) (string, error) {
	return f.createFolder(ctx, dbo.MustGetDB(ctx), req, operator)
}

func (f *FolderModel) AddItem(ctx context.Context, req entity.CreateFolderItemRequest, operator *entity.Operator) (string, error) {
	return f.addItemInternal(ctx, dbo.MustGetDB(ctx), req, operator)
}

func (f *FolderModel) UpdateFolder(ctx context.Context, folderID string, d entity.UpdateFolderRequest, operator *entity.Operator) error {
	//get folder object
	folder, err := f.getFolder(ctx, dbo.MustGetDB(ctx), folderID)
	if err != nil {
		return err
	}
	folder.Name = d.Name
	folder.Thumbnail = d.Thumbnail
	folder.Editor = operator.UserID
	err = da.GetFolderDA().UpdateFolder(ctx, dbo.MustGetDB(ctx), folderID, folder)
	if err != nil {
		log.Error(ctx, "update folder item failed", log.Err(err), log.Any("folder", folder))
		return err
	}
	return nil
}

func (f *FolderModel) updateFolderPathByLinks(ctx context.Context, tx *dbo.DBContext, links []string, path entity.Path) error {
	err := da.GetFolderDA().BatchUpdateFolderPathByLink(ctx, tx, links, path)
	if err != nil {
		log.Error(ctx, "update folder item visibility settings failed", log.Err(err), log.Any("links", links))
		return err
	}
	return nil
}

func (f *FolderModel) AddOrUpdateOrgFolderItem(ctx context.Context, tx *dbo.DBContext, partition entity.FolderPartition, link string, operator *entity.Operator) error {
	//Update folder item visibility settings
	hasLink, err := f.hasFolderFileItem(ctx, tx, entity.OwnerTypeOrganization, operator.OrgID, link)
	if err != nil {
		log.Error(ctx, "check has folder item failed", log.Err(err), log.String("link", link))
		return ErrUpdateFolderFailed
	}
	//若存在link，则无需修改
	if hasLink {
		log.Info(ctx, "already has item", log.Err(err), log.String("link", link))
		return nil
	}

	//若不存在，则创建
	//新发布的content
	folder, err := f.getRootFolder(ctx, tx, partition, entity.OwnerTypeOrganization, operator)
	if err != nil {
		log.Error(ctx, "get root folder failed", log.Err(err), log.Any("operator", operator))
		return ErrUpdateFolderFailed
	}
	_, err = f.addItemInternal(ctx, tx, entity.CreateFolderItemRequest{
		FolderID: folder.ID,
		Link:     link,
	}, operator)
	if err != nil {
		log.Error(ctx, "add folder item failed", log.Err(err),
			log.Any("folder", folder),
			log.String("link", link))
		return err
	}
	return nil
}

func (f *FolderModel) RemoveItemByLink(ctx context.Context, tx *dbo.DBContext, ownerType entity.OwnerType, owner string, link string) error {
	if !ownerType.Valid() {
		log.Warn(ctx, "invalid folder owner type", log.Int("ownerType", int(ownerType)), log.String("owner", owner), log.String("link", link))
		return ErrInvalidFolderOwnerType
	}
	condition := da.FolderCondition{
		ItemType:  int(entity.FolderItemTypeFile),
		OwnerType: int(ownerType),
		Owner:     owner,
		Link:      link,
	}
	folderItems, err := da.GetFolderDA().SearchFolder(ctx, tx, condition)
	if err != nil {
		log.Warn(ctx, "search folder failed", log.Any("condition", condition), log.Int("ownerType", int(ownerType)), log.String("owner", owner), log.String("link", link))
		return err
	}
	for i := range folderItems {
		err = f.removeItemInternal(ctx, tx, folderItems[i].ID)
		if err != nil {
			log.Error(ctx, "delete folder item failed", log.Err(err), log.Any("folderItem", folderItems[i]))
			return err
		}
	}
	return nil
}

func (f *FolderModel) RemoveItem(ctx context.Context, fid string, operator *entity.Operator) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := f.removeItemInternal(ctx, tx, fid)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
func (f *FolderModel) MoveItemBulk(ctx context.Context, fids []string, distFolder string, operator *entity.Operator) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		distFolder, err := f.getFolder(ctx, tx, distFolder)
		if err != nil {
			return err
		}
		for i := range fids {
			err := f.moveItem(ctx, tx, fids[i], distFolder, operator)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil

}

func (f *FolderModel) MoveItem(ctx context.Context, fid string, dfID string, operator *entity.Operator) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		distFolder, err := f.getFolder(ctx, tx, dfID)
		if err != nil {
			return err
		}
		err = f.moveItem(ctx, tx, fid, distFolder, operator)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (f *FolderModel) ListItems(ctx context.Context, folderID string, itemType entity.ItemType, operator *entity.Operator) ([]*entity.FolderItem, error) {
	//check owner type
	folderItems, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		ParentID: folderID,
		ItemType: int(itemType),
	})
	if err != nil {
		log.Error(ctx, "list items failed", log.Err(err), log.String("folderID", folderID))
		return nil, err
	}
	return folderItems, nil
}

func (f *FolderModel) SearchFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItem, error) {
	total, folderItems, err := da.GetFolderDA().SearchFolderPage(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		ParentID:  condition.ParentID,
		Name:      condition.Name,
		ItemType:  int(condition.ItemType),
		OwnerType: int(condition.OwnerType),
		Owner:     condition.Owner,
		Link:      condition.Link,
		//VisibilitySetting: condition.VisibilitySetting,
		ExactDirPath:      condition.Path,
		Pager:             condition.Pager,
		OrderBy:           da.NewFolderOrderBy(condition.OrderBy),
	})
	if err != nil {
		log.Warn(ctx, "list items failed", log.Err(err), log.Any("condition", condition))
		return 0, nil, err
	}
	return total, folderItems, nil
}

func (f *FolderModel) GetRootFolder(ctx context.Context, partition entity.FolderPartition, ownerType entity.OwnerType, operator *entity.Operator) (*entity.FolderItem, error) {
	return f.getRootFolder(ctx, dbo.MustGetDB(ctx), partition, ownerType, operator)
}

func (f *FolderModel) SearchPrivateFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItem, error) {
	condition.Owner = operator.UserID
	condition.OwnerType = entity.OwnerTypeUser
	return f.SearchFolder(ctx, condition, operator)
}

func (f *FolderModel) SearchOrgFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItem, error) {
	condition.Owner = operator.OrgID
	condition.OwnerType = entity.OwnerTypeOrganization

	//if condition.ItemType != entity.FolderItemTypeFolder {
	//	log.Info(ctx, "search org folder before filter visibility settings condition",
	//		log.Any("condition", condition), log.Any("operator", operator))
	//	err := f.addContentConditionFilter(ctx, &condition, operator)
	//	if err != nil {
	//		log.Warn(ctx, "addContentConditionFilter failed", log.Err(err),
	//			log.Any("condition", condition), log.Any("operator", operator))
	//		return 0, nil, err
	//	}
	//	log.Info(ctx, "search org folder after filter visibility settings condition",
	//		log.Any("condition", condition), log.Any("operator", operator))
	//}

	return f.SearchFolder(ctx, condition, operator)
}

func (f *FolderModel) GetFolderByID(ctx context.Context, folderID string, operator *entity.Operator) (*entity.FolderItemInfo, error) {
	folderItem, err := da.GetFolderDA().GetFolderByID(ctx, dbo.MustGetDB(ctx), folderID)
	if err != nil {
		log.Error(ctx, "get folder by id failed", log.Err(err), log.String("folderID", folderID))
		return nil, ErrNoFolder
	}
	result := &entity.FolderItemInfo{
		FolderItem: *folderItem,
		Items:      nil,
	}
	if folderItem.ItemType.IsFolder() {
		_, folderItems, err := f.SearchFolder(ctx, entity.SearchFolderCondition{
			ParentID: folderItem.ID,
		}, operator)
		if err != nil {
			log.Warn(ctx, "search folder by id failed", log.Err(err), log.String("folderID", folderID))
			return nil, err
		}

		result.Items = folderItems
	}
	return result, nil
}

func (f *FolderModel) checkMoveItem(ctx context.Context, folder *entity.FolderItem, distFolder *entity.FolderItem) error {
	//check if parentFolder is a folder
	if !distFolder.ItemType.IsFolder() {
		log.Error(ctx, "move to an item not folder", log.Any("parentFolder", distFolder))
		return ErrMoveToNotFolder
	}
	if folder.DirPath == "" || folder.DirPath == "/" {
		return ErrMoveRootFolder
	}

	//check if dist is folder children
	if folder.ItemType.IsFolder() && distFolder.DirPath.IsChild(folder.ID) {
		//distFolder.DirPath
		return ErrMoveToChild
	}

	//check items duplicate
	items, err := f.getItemsFromFolders(ctx, distFolder.ID)
	if err != nil {
		return err
	}
	for i := range items {
		if items[i].ItemType == entity.FolderItemTypeFile &&
			items[i].Name == folder.Name &&
			items[i].Link == folder.Link {
			log.Error(ctx, "duplicate item in path", log.Any("items", items), log.Any("folder", folder))
			return ErrDuplicateItem
		}
	}

	//组织下,不能有重复file,检查是否重复
	//if !folder.ItemType.IsFolder() && distFolder.OwnerType == entity.OwnerTypeOrganization {
	//	hasItem, err := f.hasFolderFileItem(ctx, dbo.MustGetDB(ctx), distFolder.OwnerType, distFolder.Owner, folder.Link)
	//	if err != nil {
	//		return err
	//	}
	//	if hasItem {
	//		log.Warn(ctx, "duplicate item in org folder", log.Err(err), log.Any("folder", folder), log.Any("distFolder", distFolder))
	//		return ErrDuplicateItem
	//	}
	//}

	return nil
}

func (f *FolderModel) moveItem(ctx context.Context, tx *dbo.DBContext, fid string, distFolder *entity.FolderItem, operator *entity.Operator) error {
	folder, err := f.getFolder(ctx, tx, fid)
	if err != nil {
		return err
	}

	//检查参数是否有问题
	err = f.checkMoveItem(ctx, folder, distFolder)
	if err != nil {
		return err
	}

	//获取目录下的所有文件（所有子文件一起移动）
	subItems, err := f.getDescendantItems(ctx, folder)
	if err != nil {
		return err
	}
	ids := make([]string, 0)
	links := make([]string, 0)
	for i := range subItems {
		//更新子文件时排除自己
		if subItems[i].ID != folder.ID {
			ids = append(ids, subItems[i].ID)
		}
		//更新关联文件时，只更新带link的文件
		if !subItems[i].ItemType.IsFolder() {
			links = append(links, subItems[i].Link)
		}
	}
	path := distFolder.ChildrenPath()
	folder.DirPath = path
	folder.ParentID = distFolder.ID
	err = da.GetFolderDA().UpdateFolder(ctx, tx, fid, folder)
	if err != nil {
		log.Warn(ctx, "update folder failed", log.Err(err), log.Any("folder", folder))
		return err
	}
	newPath := folder.ChildrenPath()
	err = da.GetFolderDA().BatchUpdateFolderPath(ctx, tx, ids, newPath)
	if err != nil {
		log.Warn(ctx, "update folder path failed", log.Err(err), log.Strings("ids", ids), log.String("path", string(path)))
		return err
	}

	if !folder.ItemType.IsFolder() {
		newPath = distFolder.ChildrenPath()
	}
	err = f.updateLinkedItemPath(ctx, tx, links, string(newPath))
	if err != nil {
		log.Warn(ctx, "update notify move item path failed", log.Err(err), log.Strings("ids", ids), log.Strings("links", links), log.String("path", string(path)))
		return err
	}

	//更新文件数量
	moveCount := len(subItems)
	err = da.GetFolderDA().AddFolderItemsCount(ctx, tx, folder.ParentID, -moveCount)
	if err != nil {
		log.Warn(ctx, "update originParentFolder items count failed", log.Err(err),
			log.String("folder.ParentID", folder.ParentID),
			log.Int("count", moveCount))
		return err
	}

	err = da.GetFolderDA().AddFolderItemsCount(ctx, tx, distFolder.ID, moveCount)
	if err != nil {
		log.Warn(ctx, "update distFolder items count failed", log.Err(err),
			log.Any("parentFolder", distFolder),
			log.Int("count", moveCount))
		return err
	}

	return nil
}

func (f *FolderModel) getRootFolder(ctx context.Context, tx *dbo.DBContext, partition entity.FolderPartition, ownerType entity.OwnerType, operator *entity.Operator) (*entity.FolderItem, error) {
	if !partition.Valid() {
		log.Error(ctx, "partition invalid", log.Any("partition", partition),
			log.Int("ownerType", int(ownerType)), log.Any("operator", operator))
		return nil, ErrInvalidPartition
	}
	condition := entity.SearchFolderCondition{
		Name:      string(partition),
		OwnerType: ownerType,
		Owner:     ownerType.Owner(operator),
		ParentID:  "/",
	}
	total, folderList, err := f.SearchFolder(ctx, condition, operator)
	if err != nil {
		return nil, err
	}
	if total < 1 {
		//若没有则创建一个
		id, err := f.createFolder(ctx, tx, entity.CreateFolderRequest{
			OwnerType: ownerType,
			Name:      string(partition),
		}, operator)
		if err != nil {
			return nil, err
		}
		folder, err := f.getFolder(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		return folder, nil
	}
	for i := range folderList {
		if folderList[i].DirPath == "/" {
			return folderList[i], nil
		}
	}
	log.Error(ctx, "folder item path error", log.Any("list", folderList),
		log.Int("ownerType", int(ownerType)), log.Any("operator", operator))
	return nil, ErrFolderItemPathError
}

func (f *FolderModel) hasFolderFileItem(ctx context.Context, tx *dbo.DBContext, ownerType entity.OwnerType, owner string, link string) (bool, error) {
	if !ownerType.Valid() {
		log.Warn(ctx, "invalid folder owner type", log.Int("ownerType", int(ownerType)), log.String("owner", owner), log.String("link", link))
		return false, ErrInvalidFolderOwnerType
	}
	condition := da.FolderCondition{
		ItemType:  int(entity.FolderItemTypeFile),
		OwnerType: int(ownerType),
		Owner:     owner,
		Link:      link,
	}
	total, err := da.GetFolderDA().SearchFolderCount(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "search folder failed", log.Any("condition", condition), log.Int("ownerType", int(ownerType)), log.String("owner", owner), log.String("link", link))
		return false, err
	}
	return total > 0, nil
}

func (f *FolderModel) createFolder(ctx context.Context, tx *dbo.DBContext, req entity.CreateFolderRequest, operator *entity.Operator) (string, error) {
	var parentFolder *entity.FolderItem
	var err error
	//get parent folder if exists
	if req.ParentID != "" {
		parentFolder, err = f.getFolder(ctx, tx, req.ParentID)
		if err != nil {
			return "", err
		}
	}
	//check create request
	err = f.checkCreateRequestEntity(ctx, req, parentFolder)
	if err != nil {
		return "", err
	}
	folder := f.prepareCreateFolderParams(ctx, req, parentFolder, operator)

	//创建folder
	_, err = da.GetFolderDA().CreateFolder(ctx, tx, folder)
	if err != nil {
		log.Error(ctx, "create folder failed", log.Err(err), log.Any("content", folder))
		return "", err
	}
	return folder.ID, nil
}

func (f *FolderModel) removeItemInternal(ctx context.Context, tx *dbo.DBContext, fid string) error {
	folderItem, err := f.getFolder(ctx, dbo.MustGetDB(ctx), fid)
	if err != nil {
		return err
	}

	//检查文件夹是否为空
	err = f.checkFolderEmpty(ctx, folderItem)
	if err != nil {
		return err
	}

	//若为空Folder，删除
	err = da.GetFolderDA().DeleteFolder(ctx, dbo.MustGetDB(ctx), fid)
	if err != nil {
		log.Error(ctx, "delete folder item failed", log.Err(err), log.Any("fid", fid))
		return err
	}
	//获取parent folder
	//更新parent folder文件数量
	err = da.GetFolderDA().AddFolderItemsCount(ctx, tx, folderItem.ParentID, -1)
	if err != nil {
		log.Warn(ctx, "update parent folder items count failed", log.Err(err), log.String("folderItem.ParentID", folderItem.ParentID))
		return err
	}

	return nil
}

func (f *FolderModel) addItemInternal(ctx context.Context, tx *dbo.DBContext, req entity.CreateFolderItemRequest, operator *entity.Operator) (string, error) {
	//get parent folder
	parentFolder, err := f.getFolder(ctx, tx, req.FolderID)
	if err != nil {
		return "", err
	}
	//check item
	item, err := createFolderItemByID(ctx, req.Link, operator)
	if err != nil {
		log.Warn(ctx, "check item id failed", log.Err(err), log.Any("req", req))
		return "", err
	}

	//check add item request
	err = f.checkAddItemRequest(ctx, req, parentFolder, item)
	if err != nil {
		return "", err
	}
	//build add item params
	folderItem := f.prepareAddItemParams(ctx, req, parentFolder, item, operator)

	//do create folder item
	_, err = da.GetFolderDA().CreateFolder(ctx, tx, folderItem)
	if err != nil {
		log.Warn(ctx, "create folder item failed", log.Err(err), log.Any("item", item))
		return "", err
	}

	//更新parent folder文件数量
	//parentFolder.ItemsCount = parentFolder.ItemsCount + 1
	//err = da.GetFolderDA().UpdateFolder(ctx, tx, parentFolder.ID, parentFolder)
	err = da.GetFolderDA().AddFolderItemsCount(ctx, tx, parentFolder.ID, 1)
	if err != nil {
		log.Warn(ctx, "update parent folder items count failed", log.Err(err), log.Any("parentFolder", parentFolder))
		return "", err
	}

	return folderItem.ID, nil
}
//
//func (f *FolderModel) addContentConditionFilter(ctx context.Context, condition *entity.SearchFolderCondition, operator *entity.Operator) error {
//	//获取所有查看资源的权限
//	visibilitySettings, err := GetContentModel().ListVisibleScopes(ctx, visiblePermissionPublished, operator)
//	if err != nil {
//		log.Warn(ctx, "get visibility settings failed", log.Err(err),
//			log.String("status", string(visiblePermissionPublished)), log.Any("operator", operator))
//		return err
//	}
//	condition.VisibilitySetting = append(visibilitySettings, constant.NoVisibilitySetting)
//	log.Info(ctx, "all visible scopes",
//		log.Strings("visibility settings", visibilitySettings))
//	//检查是否有查看Assets权限
//	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.CreateAssetPage301)
//	if err != nil {
//		log.Warn(ctx, "can't get schools from org", log.Err(err))
//		return err
//	} else if hasPermission {
//		condition.VisibilitySetting = append(visibilitySettings, constant.AssetsVisibilitySetting)
//	}
//	log.Info(ctx, "has assets Permission",
//		log.Bool("permission", hasPermission))
//	return nil
//}

func (f *FolderModel) prepareAddItemParams(ctx context.Context, req entity.CreateFolderItemRequest, parentFolder *entity.FolderItem, item *FolderItem, operator *entity.Operator) *entity.FolderItem {
	path := parentFolder.ChildrenPath()
	ownerType := parentFolder.OwnerType
	owner := parentFolder.Owner

	now := time.Now().Unix()
	id := utils.NewID()
	return &entity.FolderItem{
		ID:        id,
		Link:      req.Link,
		ItemType:  entity.FolderItemTypeFile,
		OwnerType: ownerType,
		Owner:     owner,
		Editor:		operator.UserID,
		ParentID:  req.FolderID,
		Name:      item.Name,
		DirPath:   path,
		//VisibilitySetting: item.VisibilitySetting,
		Thumbnail:         item.Thumbnail,
		Creator:           operator.UserID,
		CreateAt:          now,
		UpdateAt:          now,
		DeleteAt:          0,
	}
}
func (f *FolderModel) checkAddItemRequest(ctx context.Context, req entity.CreateFolderItemRequest, parentFolder *entity.FolderItem, item *FolderItem) error {
	if req.FolderID == "" {
		log.Warn(ctx, "invalid folder id", log.Any("req", req))
		return ErrEmptyFolderID
	}
	if req.Link == "" {
		log.Warn(ctx, "invalid item id", log.Any("req", req))
		return ErrEmptyFolderID
	}
	//check if parentFolder is a folder
	if !parentFolder.ItemType.IsFolder() {
		log.Warn(ctx, "move to an item not folder", log.Any("parentFolder", parentFolder))
		return ErrMoveToNotFolder
	}

	//check items duplicate
	items, err := f.getItemsFromFolders(ctx, req.FolderID)
	if err != nil {
		return err
	}
	for i := range items {
		if items[i].ItemType == entity.FolderItemTypeFile &&
			items[i].Name == item.Name &&
			items[i].Link == req.Link {
			log.Warn(ctx, "duplicate item in path", log.Any("items", items), log.Any("req", req))
			return ErrDuplicateItem
		}
	}

	//组织下,不能有重复file,检查是否重复
	if parentFolder.OwnerType == entity.OwnerTypeOrganization {
		hasItem, err := f.hasFolderFileItem(ctx, dbo.MustGetDB(ctx), parentFolder.OwnerType, parentFolder.Owner, req.Link)
		if err != nil {
			return err
		}
		if hasItem {
			log.Warn(ctx, "duplicate item in org folder", log.Err(err), log.Any("req", req), log.Any("parentFolder", parentFolder))
			return ErrDuplicateItem
		}
	}

	return nil
}

func (f *FolderModel) prepareCreateFolderParams(ctx context.Context, req entity.CreateFolderRequest, parentFolder *entity.FolderItem, operator *entity.Operator) *entity.FolderItem {
	path := entity.NewPath("/")
	ownerType := req.OwnerType
	owner := req.OwnerType.Owner(operator)
	if parentFolder != nil {
		path = parentFolder.ChildrenPath()
		ownerType = parentFolder.OwnerType
		owner = parentFolder.Owner
	}
	if req.ParentID == "" {
		req.ParentID = "/"
	}
	now := time.Now().Unix()
	id := utils.NewID()
	return &entity.FolderItem{
		ID:        id,
		ItemType:  entity.FolderItemTypeFolder,
		OwnerType: ownerType,
		Owner:     owner,
		ParentID:  req.ParentID,
		Editor:		operator.UserID,
		Name:      req.Name,
		DirPath:   path,
		Thumbnail: req.Thumbnail,
		//VisibilitySetting: constant.NoVisibilitySetting,
		Creator:           operator.UserID,
		CreateAt:          now,
		UpdateAt:          now,
		DeleteAt:          0,
	}
}

func (f *FolderModel) checkCreateRequestEntity(ctx context.Context, req entity.CreateFolderRequest, parentFolder *entity.FolderItem) error {
	//check name
	if req.Name == "" {
		log.Warn(ctx, "empty folder name", log.Any("req", req))
		return ErrEmptyFolderName
	}
	//check owner type
	if !req.OwnerType.Valid() {
		log.Warn(ctx, "invalid folder owner type", log.Any("req", req))
		return ErrInvalidFolderOwnerType
	}
	if parentFolder != nil && !parentFolder.ItemType.IsFolder() {
		log.Warn(ctx, "move to an item not folder", log.Any("req", req))
		return ErrMoveToNotFolder
	}

	//check duplicate name
	err := f.checkDuplicateFolderName(ctx, req.ParentID, req.Name)
	if err != nil {
		return err
	}

	return nil
}

func (f *FolderModel) checkFolderEmpty(ctx context.Context, folderItem *entity.FolderItem) error {
	//若不是folder，返回
	if !folderItem.ItemType.IsFolder() {
		log.Info(ctx, "item is not folder", log.Any("folderItem", folderItem), log.String("path", string(folderItem.DirPath)))
		return nil
	}
	if folderItem.ItemsCount > 0 {
		return ErrFolderIsNotEmpty
	}

	////若是folder，检查是否为空folder
	//total, subItems, err := da.GetFolderDA().SearchFolderPage(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
	//	DirDescendant: folderItem.DirDescendant.ParentPath() + "/" + folderItem.ID,
	//})
	//if err != nil {
	//	log.Error(ctx, "search sub folder item failed", log.Err(err), log.Any("folderItem", folderItem), log.String("path", string(folderItem.DirDescendant)))
	//	return err
	//}
	////若total > 1,则表示一定有子文件，不能删除
	//if total > 1 {
	//	log.Error(ctx, "folder is not empty", log.Err(err),
	//		log.Int("total", total),
	//		log.String("path", string(folderItem.DirDescendant)),
	//		log.Any("folderItem", folderItem),
	//		log.Any("items", subItems))
	//	return ErrFolderIsNotEmpty
	//}
	return nil
}

func (f *FolderModel) checkDuplicateFolderName(ctx context.Context, parentId string, name string) error {
	//check get all sub folders from parent folder
	condition := da.FolderCondition{
		IDs:           nil,
		ItemType:     	int(entity.FolderItemTypeFolder),
		//ParentID:      parentId,
		Name:          name,
	}
	total, err := da.GetFolderDA().SearchFolderCount(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		log.Error(ctx, "count folder for check duplicate folder failed",
			log.Err(err), log.Any("condition", condition))
		return err
	}
	//check duplicate folder name
	if total > 0 {
		return ErrDuplicateFolderName
	}

	return nil
}

func (f *FolderModel) getDescendantItems(ctx context.Context, folder *entity.FolderItem) ([]*entity.FolderItem, error) {
	//若为文件，则直接返回
	if !folder.ItemType.IsFolder() {
		return []*entity.FolderItem{folder}, nil
	}
	items, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		DirDescendant: string(folder.ChildrenPath()),
	})
	if err != nil {
		log.Warn(ctx, "search folder failed", log.Err(err), log.String("path", string(folder.DirPath)))
		return nil, err
	}
	items = append(items, folder)
	return items, nil
}

func (f *FolderModel) getSubFolders(ctx context.Context, fid string) ([]*entity.FolderItem, error) {
	relations, err := f.getItemsFromFolders(ctx, fid)
	if err != nil {
		return nil, err
	}
	folderItems := make([]string, 0)
	for i := range relations {
		if relations[i].ItemType == entity.FolderItemTypeFolder {
			folderItems = append(folderItems, relations[i].ID)
		}
	}
	if len(folderItems) > 0 {
		folders, err := da.GetFolderDA().GetFolderByIDList(ctx, dbo.MustGetDB(ctx), folderItems)
		if err != nil {
			log.Warn(ctx, "get folder info failed", log.Err(err), log.Strings("ids", folderItems))
			return nil, err
		}
		return folders, nil
	}

	return nil, nil
}

func (f *FolderModel) getItemsFromFolders(ctx context.Context, parentID string) ([]*entity.FolderItem, error) {
	items, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		ParentID: parentID,
	})
	if err != nil {
		log.Warn(ctx, "get folder items failed", log.Err(err), log.String("id", parentID))
		return nil, err
	}
	return items, nil
}
func (f *FolderModel) getFolder(ctx context.Context, tx *dbo.DBContext, fid string) (*entity.FolderItem, error) {
	parentFolder, err := da.GetFolderDA().GetFolderByID(ctx, tx, fid)
	if err != nil || parentFolder == nil {
		log.Warn(ctx, "no such folder id", log.Err(err), log.String("fid", fid))
		return nil, ErrInvalidParentFolderId
	}
	return parentFolder, nil
}

type FolderItem struct {
	Name              string
	Thumbnail         string
	VisibilitySetting string
}

func (f *FolderModel) updateLinkedItemPath(ctx context.Context, tx *dbo.DBContext, links []string, path string) error {
	for _, link := range links {
		fileType, id, err := parseLink(ctx, link)
		if err != nil {
			return err
		}
		switch fileType {
		case entity.FileTypeContent:
			err = GetContentModel().UpdateContentPath(ctx, tx, id, path)
			if err != nil {
				log.Warn(ctx, "can't update content path by id", log.Err(err),
					log.String("itemType", fileType), log.String("id", id), log.String("path", path))
				return err
			}
		default:
			log.Warn(ctx, "unsupported file type",
				log.String("itemType", fileType), log.String("id", id))
			return ErrInvalidFolderItemType
		}
	}
	return nil
}

func createFolderItemByID(ctx context.Context, link string, user *entity.Operator) (*FolderItem, error) {
	fileType, id, err := parseLink(ctx, link)
	if err != nil {
		return nil, err
	}
	switch fileType {
	case entity.FileTypeContent:
		content, err := GetContentModel().GetContentByID(ctx, dbo.MustGetDB(ctx), id, user)
		if err != nil {
			log.Warn(ctx, "can't find content by id", log.Err(err),
				log.String("itemType", fileType), log.String("id", id))
			return nil, err
		}
		return &FolderItem{
			Name:              content.Name,
			Thumbnail:         content.Thumbnail,
			VisibilitySetting: content.PublishScope,
		}, nil
	}
	log.Warn(ctx, "unsupported file type",
		log.String("itemType", fileType), log.String("id", id))
	return nil, ErrInvalidFolderItemType
}

func parseLink(ctx context.Context, link string) (string, string, error) {
	linkPairs := strings.Split(link, "-")
	if len(linkPairs) != 2 {
		log.Warn(ctx, "link is invalid", log.Err(ErrInvalidItemLink),
			log.String("link", link))
		return "", "", ErrInvalidItemLink
	}
	fileType := linkPairs[0]
	id := linkPairs[1]
	return fileType, id, nil
}

var (
	folderModel      IFolderModel
	_folderModelOnce sync.Once
)

func GetFolderModel() IFolderModel {
	_folderModelOnce.Do(func() {
		folderModel = new(FolderModel)
	})

	return folderModel
}
