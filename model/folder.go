package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidParentFolderId = errors.New("invalid parent folder id")
	ErrEmptyFolderName = errors.New("empty folder name")
	ErrEmptyFolderID = errors.New("empty folder id")
	ErrEmptyItemID = errors.New("empty item id")
	ErrInvalidFolderOwnerType = errors.New("invalid folder owner type")
	ErrInvalidFolderItemType = errors.New("invalid folder item type")
	ErrDuplicateFolderName = errors.New("duplicate folder name in path")
	ErrDuplicateItem = errors.New("duplicate item in path")
	ErrFolderIsNotEmpty = errors.New("folder is not empty")
	ErrInvalidItemLink = errors.New("invalid item link")
	ErrUpdateFolderFailed  = errors.New("update folder into data access failed")
)

type IFolderModel interface{
	//创建Folder
	CreateFolder(ctx context.Context, req entity.CreateFolderRequest, operator *entity.Operator)(string ,error)
	//添加一个item
	AddItem(ctx context.Context, req entity.CreateFolderItemRequest, operator *entity.Operator) (string, error)
	//删除一批item
	RemoveItem(ctx context.Context, fid string, operator *entity.Operator) error

	//修改Folder
	UpdateFolder(ctx context.Context, folderID string, d entity.UpdateFolderRequest, operator *entity.Operator) error


	//移动item
	MoveItem(ctx context.Context, fid string, distFolder string, operator *entity.Operator) error

	//列出Folder下的所有item
	ListItems(ctx context.Context, folderID string, itemType entity.ItemType, operator *entity.Operator) ([]*entity.FolderItem, error)
	//查询Folder
	SearchFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator)(int, []*entity.FolderItem, error)
	SearchPrivateFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator)(int, []*entity.FolderItem, error)
	SearchOrgFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator)(int, []*entity.FolderItem, error)

	//获取Folder
	GetFolderByID(ctx context.Context, folderID string, operator *entity.Operator) (*entity.FolderItemInfo, error)

	//内部API，修改Folder的Visibility Settings
	AddOrUpdateOrgFolderItem(ctx context.Context, tx *dbo.DBContext, link string, visibilitySettings string, operator *entity.Operator) error
	RemoveItemByLink(ctx context.Context, tx *dbo.DBContext, ownerType entity.OwnerType, owner string, link string) error
}

type FolderModel struct{}

func (f *FolderModel) CreateFolder(ctx context.Context, req entity.CreateFolderRequest, operator *entity.Operator) (string, error) {
	return f.createFolder(ctx, dbo.MustGetDB(ctx), req, operator)
}

func (f *FolderModel) createFolder(ctx context.Context, tx *dbo.DBContext, req entity.CreateFolderRequest, operator *entity.Operator) (string, error) {
	parentFolder, err := f.prepareCreateRequestEntity(ctx, req)
	if err != nil{
		return "", err
	}
	path := "/"
	ownerType := req.OwnerType
	owner := req.OwnerType.Owner(operator)
	if parentFolder != nil {
		path = parentFolder.Path.ParentPath() + "/" + parentFolder.ID
		ownerType = parentFolder.OwnerType
		owner = parentFolder.Owner
	}
	now := time.Now().Unix()
	id := utils.NewID()
	folder := entity.FolderItem{
		ID:          id,
		ItemType: 	 entity.FolderItemTypeFolder,
		OwnerType:   ownerType,
		Owner:       owner,
		ParentId:    req.ParentId,
		Name:        req.Name,
		Path:		 entity.NewPath(path),
		Thumbnail:   req.Thumbnail,
		VisibilitySetting: constant.NoVisibilitySetting,
		Creator:	 operator.UserID,
		CreateAt:    now,
		UpdateAt:    now,
		DeleteAt:    0,
	}
	//创建folder
	_, err = da.GetFolderDA().CreateFolder(ctx, tx, folder)
	if err != nil{
		log.Error(ctx, "create folder failed", log.Err(err), log.Any("content", folder))
		return "", err
	}
	return id, nil
}

func (f *FolderModel) addItemInternal(ctx context.Context, tx *dbo.DBContext, req entity.CreateFolderItemRequest, operator *entity.Operator) (string, error) {
	//check item
	item, err := CreateFolderItemById(ctx, req.Link, operator)
	if err != nil{
		log.Warn(ctx, "check item id failed", log.Err(err), log.Any("req", req))
		return "", err
	}

	parentFolder, err := f.checkAddItemRequest(ctx, req, item)
	if err != nil{
		return "", err
	}
	//组织下,不能有重复file,检查是否重复
	if parentFolder.OwnerType == entity.OwnerTypeOrganization {
		hasItem, err := f.hasFolderFileItem(ctx, dbo.MustGetDB(ctx), parentFolder.OwnerType, parentFolder.Owner, req.Link)
		if err != nil{
			return "", err
		}
		if hasItem {
			log.Warn(ctx, "duplicate item in org folder", log.Err(err), log.Any("req", req), log.Any("parentFolder", parentFolder))
			return "", ErrDuplicateItem
		}
	}

	path := parentFolder.Path.ParentPath() + "/" + parentFolder.ID
	ownerType := parentFolder.OwnerType
	owner := parentFolder.Owner

	now := time.Now().Unix()
	id := utils.NewID()
	folderItem := entity.FolderItem{
		ID:          id,
		Link:		 req.Link,
		ItemType: 	 entity.FolderItemTypeFile,
		OwnerType:   ownerType,
		Owner:       owner,
		ParentId:    req.FolderID,
		Name:        item.Name,
		Path:		 entity.NewPath(path),
		VisibilitySetting: item.VisibilitySetting,
		Thumbnail:   item.Thumbnail,
		Creator:	 operator.UserID,
		CreateAt:    now,
		UpdateAt:    now,
		DeleteAt:    0,
	}
	_, err = da.GetFolderDA().CreateFolder(ctx, tx, folderItem)
	if err != nil{
		log.Warn(ctx, "create folder item failed", log.Err(err), log.Any("item", item))
		return "", err
	}
	return id, nil
}

func (f *FolderModel) AddItem(ctx context.Context, req entity.CreateFolderItemRequest, operator *entity.Operator) (string, error) {
	return f.addItemInternal(ctx, dbo.MustGetDB(ctx), req, operator)
}

func (f *FolderModel) UpdateFolder(ctx context.Context, folderID string, d entity.UpdateFolderRequest, operator *entity.Operator) error {
	folder, err := f.mustGetFolder(ctx, folderID)
	if err != nil{
		return err
	}
	folder.Name = d.Name
	folder.Thumbnail = d.Thumbnail
	err = da.GetFolderDA().UpdateFolder(ctx, dbo.MustGetDB(ctx), folderID, *folder)
	if err != nil{
		log.Error(ctx, "update folder item failed", log.Err(err), log.Any("folder", folder))
		return err
	}
	return nil
}


func (f *FolderModel) updateFolderVisibilitySetting(ctx context.Context, tx *dbo.DBContext, links []string, visibilitySetting string) error {
	err := da.GetFolderDA().BatchUpdateFolderVisibilitySettings(ctx, tx, links, visibilitySetting)
	if err != nil{
		log.Error(ctx, "update folder item visibility settings failed", log.Err(err), log.Any("links", links))
		return err
	}
	return nil
}

func (f *FolderModel) AddOrUpdateOrgFolderItem(ctx context.Context, tx *dbo.DBContext, link string, visibilitySettings string, operator *entity.Operator) error {
	//Update folder item visibility settings
	hasLink, err := f.hasFolderFileItem(ctx, tx, entity.OwnerTypeOrganization, operator.OrgID, link)
	if err != nil {
		log.Error(ctx, "check has folder item failed", log.Err(err), log.String("link", link))
		return ErrUpdateFolderFailed
	}
	//若存在content，则更新visibility setting
	//已发布的情况下修改visibility setting
	if hasLink {
		err = f.updateFolderVisibilitySetting(ctx, tx, []string{link}, visibilitySettings)
		if err != nil {
			log.Error(ctx, "update folder visibility setting failed", log.Err(err))
			return ErrUpdateFolderFailed
		}
	}

	//若不存在，则创建
	//新发布的content
	folderID, err := f.getRootFolder(ctx, tx, entity.OwnerTypeOrganization, operator)
	if err != nil {
		log.Error(ctx, "get root folder failed", log.Err(err), log.Any("operator", operator))
		return ErrUpdateFolderFailed
	}
	_, err = f.addItemInternal(ctx, tx, entity.CreateFolderItemRequest{
		FolderID: folderID,
		Link:     link,
	}, operator)
	if err != nil {
		log.Error(ctx, "add folder item failed", log.Err(err),
			log.String("folderId", folderID),
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
		ItemType: int(entity.FolderItemTypeFile),
		OwnerType: int(ownerType),
		Owner: owner,
		Link: link,
	}
	_, folderItems, err := da.GetFolderDA().SearchFolder(ctx, tx, condition)
	if err != nil{
		log.Warn(ctx, "search folder failed", log.Any("condition", condition), log.Int("ownerType", int(ownerType)), log.String("owner", owner), log.String("link", link))
		return err
	}
	for i := range folderItems {
		err = da.GetFolderDA().DeleteFolder(ctx, tx, folderItems[i].ID)
		if err != nil{
			log.Error(ctx, "delete folder item failed", log.Err(err), log.Any("folderItem", folderItems[i]))
			return err
		}
	}
	return nil
}

func (f *FolderModel) RemoveItem(ctx context.Context, fid string, operator *entity.Operator) error {
	//检查文件夹是否为空
	err := f.checkFolderEmpty(ctx, fid)
	if err != nil{
		return err
	}

	//若为空Folder，删除

	err = da.GetFolderDA().DeleteFolder(ctx, dbo.MustGetDB(ctx), fid)
	if err != nil{
		log.Error(ctx, "delete folder item failed", log.Err(err), log.Any("fid", fid))
		return err
	}
	return nil
}

func (f *FolderModel) MoveItem(ctx context.Context, fid string, dfID string, operator *entity.Operator) error {
	distFolder, err := f.mustGetFolder(ctx, dfID)
	if err != nil{
		return err
	}
	folder, err := f.mustGetFolder(ctx, fid)
	if err != nil{
		return err
	}

	subItems, err := f.getDescendantItems(ctx, folder)
	if err != nil{
		return err
	}
	ids := make([]string, 0)
	for i := range subItems {
		if subItems[i].ID != folder.ID {
			ids = append(ids, subItems[i].ID)
		}
	}

	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		path := distFolder.Path.ParentPath() + "/" + distFolder.ID
		folder.Path = entity.NewPath(path)
		folder.ParentId = distFolder.ID
		err := da.GetFolderDA().UpdateFolder(ctx, tx, fid, *folder)
		if err != nil{
			log.Warn(ctx, "update folder failed", log.Err(err), log.Any("folder", folder))
			return err
		}
		err = da.GetFolderDA().BatchUpdateFolderPath(ctx, tx, ids, path + "/" + folder.ID)
		if err != nil{
			log.Warn(ctx, "update folder path failed", log.Err(err), log.Strings("ids", ids), log.String("path", path))
			return err
		}
		return nil
	})
	if err != nil{
		return err
	}
	return nil
}

func (f *FolderModel) ListItems(ctx context.Context, folderID string, itemType entity.ItemType, operator *entity.Operator) ([]*entity.FolderItem, error) {
	//check owner type
	_, folderItems, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		ParentId:  folderID,
		ItemType: int(itemType),
	})
	if err != nil{
		log.Error(ctx, "list items failed", log.Err(err), log.String("folderID", folderID))
		return nil, err
	}
	return folderItems, nil
}

func (f *FolderModel) SearchFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItem, error) {
	total, folderItems, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		ParentId:  condition.ParentId,
		ItemType: int(condition.ItemType),
		OwnerType: int(condition.OwnerType),
		Owner: condition.Owner,
		Link: condition.Link,
		VisibilitySetting: condition.VisibilitySetting,
		ExactPath: condition.Path,
		Pager: condition.Pager,
		OrderBy: da.NewFolderOrderBy(condition.OrderBy),
	})
	if err != nil{
		log.Warn(ctx, "list items failed", log.Err(err), log.Any("condition", condition))
		return 0, nil, err
	}
	return total, folderItems, nil
}

func (f *FolderModel) SearchPrivateFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItem, error) {
	condition.Owner = operator.UserID
	condition.OwnerType = entity.OwnerTypeUser
	return f.SearchFolder(ctx, condition, operator)
}

func (f *FolderModel) SearchOrgFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItem, error) {
	condition.Owner = operator.OrgID
	condition.OwnerType = entity.OwnerTypeOrganization
	//add visibility settings
	visibilitySettings, err := GetContentModel().ListVisibleScopes(ctx, visiblePermissionPublished, operator)
	if err != nil{
		log.Warn(ctx, "get visibility settings failed", log.Err(err),
			log.String("status", string(visiblePermissionPublished)), log.Any("operator", operator))
		return 0, nil, err
	}
	condition.VisibilitySetting = visibilitySettings
	return f.SearchFolder(ctx, condition, operator)
}

func (f *FolderModel) getRootFolder(ctx context.Context, tx *dbo.DBContext, ownerType entity.OwnerType, operator *entity.Operator) (string, error){
	condition := entity.SearchFolderCondition{
		IDs:               nil,
		OwnerType:         ownerType,
		Owner:             ownerType.Owner(operator),
		Path:              "/",
	}
	total, folderList, err := f.SearchFolder(ctx, condition, operator)
	if err != nil{
		return "", err
	}
	if total < 1 {
		//若没有则创建一个
		id, err := f.createFolder(ctx, tx, entity.CreateFolderRequest{
			OwnerType: ownerType,
			Name:      "root",
		}, operator)
		if err != nil{
			return "", err
		}
		return id, nil
	}
	return folderList[0].ID, nil
}

func (f *FolderModel) hasFolderFileItem(ctx context.Context, tx *dbo.DBContext, ownerType entity.OwnerType, owner string, link string) (bool, error){
	if !ownerType.Valid() {
		log.Warn(ctx, "invalid folder owner type", log.Int("ownerType", int(ownerType)), log.String("owner", owner), log.String("link", link))
		return false, ErrInvalidFolderOwnerType
	}
	condition := da.FolderCondition{
		ItemType: int(entity.FolderItemTypeFile),
		OwnerType: int(ownerType),
		Owner: owner,
		Link: link,
	}
	total, _, err := da.GetFolderDA().SearchFolder(ctx, tx, condition)
	if err != nil{
		log.Warn(ctx, "search folder failed", log.Any("condition", condition), log.Int("ownerType", int(ownerType)), log.String("owner", owner), log.String("link", link))
		return false, err
	}
	return total > 0, nil
}

func (f *FolderModel) GetFolderByID(ctx context.Context, folderID string, operator *entity.Operator) (*entity.FolderItemInfo, error) {
	folderItem, err := da.GetFolderDA().GetFolderByID(ctx, dbo.MustGetDB(ctx), folderID)
	if err != nil{
		log.Warn(ctx, "get folder by id failed", log.Err(err), log.String("folderID", folderID))
		return nil, err
	}
	result := &entity.FolderItemInfo{
		FolderItem: *folderItem,
		Items:      nil,
	}
	if folderItem.ItemType.IsFolder() {
		_, folderItems, err := f.SearchFolder(ctx, entity.SearchFolderCondition{
			ParentId:  folderItem.ID,
		}, operator)
		if err != nil{
			log.Warn(ctx, "search folder by id failed", log.Err(err), log.String("folderID", folderID))
			return nil, err
		}

		result.Items = folderItems
	}
	return result, nil
}

func (f *FolderModel) checkAddItemRequest(ctx context.Context, req entity.CreateFolderItemRequest, item *FolderItem) (*entity.FolderItem,error) {
	//1.check folder owner
	//2.check item
	//if !req.ItemType.ValidExcludeFolder() {
	//	log.Warn(ctx, "invalid folder item type", log.Any("req", req))
	//	return nil, ErrInvalidFolderItemType
	//}
	//check owner type
	//if !req.OwnerType.Valid() {
	//	log.Warn(ctx, "invalid folder owner type", log.Any("req", req))
	//	return nil, ErrInvalidFolderOwnerType
	//}

	if req.FolderID == "" {
		log.Warn(ctx, "invalid folder id", log.Any("req", req))
		return nil, ErrEmptyFolderID
	}
	if req.Link == "" {
		log.Warn(ctx, "invalid item id", log.Any("req", req))
		return nil, ErrEmptyFolderID
	}

	//check fold id
	folder, err := f.mustGetFolder(ctx, req.FolderID)
	if err != nil{
		return nil, err
	}

	//check items duplicate
	items, err := f.getItemsFromFolders(ctx, req.FolderID)
	if err != nil{
		return nil, err
	}
	for i := range items {
		if items[i].ItemType == entity.FolderItemTypeFile &&
			items[i].Name == item.Name &&
			items[i].Link == req.Link{
			log.Warn(ctx, "duplicate item in path", log.Any("items", items), log.Any("req", req))
			return nil, ErrDuplicateItem
		}
	}

	return folder, nil
}

func (f *FolderModel) prepareCreateRequestEntity(ctx context.Context, req entity.CreateFolderRequest) (*entity.FolderItem, error) {
	//check name
	if req.Name == "" {
		log.Warn(ctx, "empty folder name", log.Any("req", req))
		return nil, ErrEmptyFolderName
	}
	//check owner type
	if !req.OwnerType.Valid() {
		log.Warn(ctx, "invalid folder owner type", log.Any("req", req))
		return nil, ErrInvalidFolderOwnerType
	}

	//check parentId
	var parentFolder *entity.FolderItem
	var err error
	if req.ParentId != "" {
		parentFolder, err = f.mustGetFolder(ctx, req.ParentId)
		if err != nil{
			return nil, err
		}
	}
	//check duplicate name
	err = f.checkDuplicateFolderName(ctx, req.ParentId, req.Name)
	if err != nil{
		return nil, err
	}

	return parentFolder, nil
}

func (f *FolderModel) checkFolderEmpty(ctx context.Context, fid string) error {
	folderItem, err := f.mustGetFolder(ctx, fid)
	if err != nil{
		return err
	}
	//若不是folder，返回
	if !folderItem.ItemType.IsFolder() {
		log.Info(ctx, "item is not folder", log.String("fid", fid), log.String("path", string(folderItem.Path)))
		return nil
	}

	//若是folder，检查是否为空folder
	total, subItems, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		Path:      folderItem.Path.ParentPath() + "/" + folderItem.ID,
	})
	if err != nil{
		log.Error(ctx, "search sub folder item failed", log.Err(err), log.String("fid", fid), log.String("path", string(folderItem.Path)))
		return err
	}
	//若total > 1,则表示一定有子文件，不能删除
	if total > 1 {
		log.Error(ctx, "folder is not empty", log.Err(err),
			log.Any("total", total),
			log.String("path", string(folderItem.Path)),
			log.String("fid", fid),
			log.Any("items", subItems))
		return ErrFolderIsNotEmpty
	}
	return nil
}

func (f *FolderModel) checkDuplicateFolderName(ctx context.Context, parentId string, name string) error {
	//check get all sub folders from parent folder
	subFolders, err := f.getSubFolders(ctx, parentId)
	if err != nil{
		return err
	}
	//check duplicate folder name
	for i := range subFolders{
		if name == subFolders[i].Name {
			return ErrDuplicateFolderName
		}
	}
	return nil
}

func (f *FolderModel) getDescendantItems(ctx context.Context, folder *entity.FolderItem) ([]*entity.FolderItem, error){
	//若为文件，则直接返回
	if !folder.ItemType.IsFolder() {
		return []*entity.FolderItem{folder}, nil
	}
	_, items, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		Path:  folder.Path.ParentPath() + "/" + folder.ID,
	})
	if err != nil{
		log.Warn(ctx, "search folder failed", log.Err(err), log.String("path", string(folder.Path)))
		return nil, err
	}
	return items, nil
}

func (f *FolderModel) getSubFolders(ctx context.Context, fid string) ([]*entity.FolderItem, error) {
	relations, err := f.getItemsFromFolders(ctx, fid)
	if err != nil{
		return nil, err
	}
	folderItems := make([]string, 0)
	for i := range relations {
		if relations[i].ItemType == entity.FolderItemTypeFolder{
			folderItems = append(folderItems, relations[i].ID)
		}
	}
	if len(folderItems) > 0 {
		folders, err := da.GetFolderDA().GetFolderByIDList(ctx, dbo.MustGetDB(ctx), folderItems)
		if err != nil{
			log.Warn(ctx, "get folder info failed", log.Err(err), log.Strings("ids", folderItems))
			return nil, err
		}
		return folders, nil
	}

	return nil, nil
}

func (f *FolderModel) getItemsFromFolders(ctx context.Context, parentId string) ([]*entity.FolderItem, error) {
	_, items, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		ParentId:  parentId,
	})
	if err != nil{
		log.Warn(ctx, "get folder items failed", log.Err(err), log.String("id", parentId))
		return nil, err
	}
	return items, nil
}
func (f *FolderModel) mustGetFolder(ctx context.Context, fid string) (*entity.FolderItem, error) {
	parentFolder, err := da.GetFolderDA().GetFolderByID(ctx, dbo.MustGetDB(ctx), fid)
	if err != nil || parentFolder == nil {
		log.Warn(ctx, "no such folder id", log.Err(err), log.String("fid", fid))
		return nil, ErrInvalidParentFolderId
	}
	return parentFolder, nil
}

type FolderItem struct {
	Name string
	Thumbnail string
	VisibilitySetting string
}

func CreateFolderItemById(ctx context.Context, link string, user *entity.Operator) (*FolderItem, error) {
	linkPairs := strings.Split(link, "-")
	if len(linkPairs) != 2 {
		log.Warn(ctx, "link is invalid", log.Err(ErrInvalidItemLink),
			log.String("link", link))
		return nil, ErrInvalidItemLink
	}
	fileType := linkPairs[0]
	id := linkPairs[1]
	switch fileType{
	case entity.FileTypeContent:
		content, err := GetContentModel().GetContentByID(ctx, dbo.MustGetDB(ctx), id, user)
		if err != nil{
			log.Warn(ctx, "can't find content by id", log.Err(err),
				log.String("itemType", fileType), log.String("id", id))
			return nil, err
		}
		return &FolderItem{
			Name:      content.Name,
			Thumbnail: content.Thumbnail,
			VisibilitySetting: content.PublishScope,
		}, nil
	}
	log.Warn(ctx, "unsupported file type",
		log.String("itemType", fileType), log.String("id", id))
	return nil, ErrInvalidFolderItemType
}


var (
	folderModel    IFolderModel
	_folderModelOnce sync.Once
)

func GetFolderModel() IFolderModel {
	_folderModelOnce.Do(func() {
		folderModel = new(FolderModel)
	})

	return folderModel
}
