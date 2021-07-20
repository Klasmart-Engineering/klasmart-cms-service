package model

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrInvalidParentFolderId  = errors.New("invalid parent folder id")
	ErrEmptyFolderName        = errors.New("empty folder name")
	ErrEmptyLinkID            = errors.New("empty link id")
	ErrMoveToNotFolder        = errors.New("move to an item not folder")
	ErrInvalidFolderOwnerType = errors.New("invalid folder owner type")
	ErrInvalidPartition       = errors.New("invalid partition")
	ErrInvalidFolderItemType  = errors.New("invalid folder item type")
	ErrDuplicateFolderName    = errors.New("duplicate folder name in path")
	ErrDuplicateItem          = errors.New("duplicate item in path")
	ErrFolderIsNotEmpty       = errors.New("folder is not empty")
	ErrInvalidItemLink        = errors.New("invalid item link")
	ErrNoFolder               = errors.New("no folder")
	ErrMoveToChild            = errors.New("move to child folder")
	ErrMoveToSameFolder       = errors.New("move to same folder")

	ErrShareToUnsupportedRegion = errors.New("share to unsupported region")

	ErrSearchSharedFolderFailed = errors.New("search shared folder failed")

	ErrNotHeadquartersShare     = errors.New("not headquarters share folder")
	ErrUnknownHeadquarterRegion = errors.New("unknown headquarter region")
)

type IFolderModel interface {
	//创建Folder
	//Create a folder
	CreateFolder(ctx context.Context, req entity.CreateFolderRequest, operator *entity.Operator) (string, error)

	//删除一批item
	//Delete a folder item
	RemoveItem(ctx context.Context, fid string, operator *entity.Operator) error

	//批量删除item
	//Batch delete folder items
	RemoveItemBulk(ctx context.Context, fids []string, operator *entity.Operator) error

	//修改Folder
	//Update folder
	UpdateFolder(ctx context.Context, folderID string, d entity.UpdateFolderRequest, operator *entity.Operator) error

	//移动item
	//Move folder items
	MoveItem(ctx context.Context, req entity.MoveFolderRequest, operator *entity.Operator) error

	//移动item
	//Move folder items
	MoveItemBulk(ctx context.Context, req entity.MoveFolderIDBulkRequest, operator *entity.Operator) error

	//列出Folder下的所有item
	//List all folder items belongs to the folder
	ListItems(ctx context.Context, folderID string, itemType entity.ItemType, operator *entity.Operator) ([]*entity.FolderItemInfo, error)
	//查询Folder
	//Query folders
	SearchFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItemInfo, error)
	SearchPrivateFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItemInfo, error)
	SearchOrgFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItemInfo, error)

	//获取Folder
	//Get Folder
	GetFolderByID(ctx context.Context, folderID string, operator *entity.Operator) (*entity.FolderItemInfo, error)

	//Share folder
	ShareFolders(ctx context.Context, req entity.ShareFoldersRequest, operator *entity.Operator) error
	//Get share records
	GetFoldersSharedRecords(ctx context.Context, fids []string, operator *entity.Operator) (*entity.FolderShareRecords, error)

	//内部API，修改Folder的Visibility Settings
	//internal API, for updating folder's visibility settings
	//查看路径是否存在
	//check path existing
	MustGetPath(ctx context.Context, tx *dbo.DBContext, ownerType entity.OwnerType, itemType entity.ItemType, path string, partition entity.FolderPartition, operator *entity.Operator) (string, error)
}

type FolderModel struct{}

func (f *FolderModel) CreateFolder(ctx context.Context, req entity.CreateFolderRequest, operator *entity.Operator) (string, error) {
	return f.createFolder(ctx, dbo.MustGetDB(ctx), req, operator)
}

func (f *FolderModel) GetFoldersSharedRecords(ctx context.Context, fids []string, operator *entity.Operator) (*entity.FolderShareRecords, error) {
	records, err := da.GetSharedFolderDA().Search(ctx, dbo.MustGetDB(ctx), da.SharedFolderCondition{
		FolderIDs: fids,
	})
	if err != nil {
		log.Error(ctx, "Get folders failed",
			log.Err(err),
			log.Strings("fids", fids),
			log.Any("operator", operator))
		return nil, err
	}
	orgIDs := make([]string, 0)
	folderOrgsMap := make(map[string][]string)
	for i := range records {
		folderOrgsMap[records[i].FolderID] = append(folderOrgsMap[records[i].FolderID], records[i].OrgID)
		//TODO:fix share_all to all orgInfo
		//orgIDs = append(orgIDs, records[i].OrgID)
	}
	orgIDs = utils.SliceDeduplication(orgIDs)
	orgs, err := external.GetOrganizationServiceProvider().BatchGet(ctx, operator, orgIDs)
	if err != nil {
		log.Error(ctx, "Get org failed",
			log.Err(err),
			log.Any("ids", orgIDs),
			log.Any("operator", operator))
		return nil, err
	}
	orgMap := make(map[string]*external.NullableOrganization)
	for i := range orgs {
		if orgs[i].Valid {
			orgMap[orgs[i].ID] = orgs[i]
		}
	}
	result := new(entity.FolderShareRecords)
	for folderId, ids := range folderOrgsMap {
		ids = utils.SliceDeduplication(ids)
		orgs := make([]*entity.OrganizationInfo, len(ids))
		for i := range ids {
			org := orgMap[ids[i]]
			name := ""
			if org != nil {
				name = org.Name
			}
			orgs[i] = &entity.OrganizationInfo{
				ID:   ids[i],
				Name: name,
			}
		}
		folderInfo := &entity.FolderShareRecord{
			FolderID: folderId,
			Orgs:     orgs,
		}
		result.Data = append(result.Data, folderInfo)
	}
	return result, nil
}

func (f *FolderModel) ShareFolders(ctx context.Context, req entity.ShareFoldersRequest, operator *entity.Operator) error {
	//0.check headquarter region
	err := f.checkHeadquarterRegionOrganizations(ctx, req.OrgIDs, operator)
	if err != nil {
		return err
	}

	if len(req.FolderIDs) < 1 {
		log.Warn(ctx, "No folder ids", log.Any("req", req))
		return nil
	}

	//1.Get folder & check folder exists
	folderIDs := utils.SliceDeduplication(req.FolderIDs)
	orgIDs := utils.SliceDeduplication(req.OrgIDs)
	for i := range folderIDs {
		//lock all folders for share folders
		locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixFolderShare, folderIDs[i])
		if err != nil {
			return err
		}
		locker.Lock()
		defer locker.Unlock()
	}

	//get folder objects by folder ids
	folders, err := da.GetFolderDA().GetFolderByIDList(ctx, dbo.MustGetDB(ctx), folderIDs)
	if err != nil {
		log.Error(ctx, "Get folders failed",
			log.Err(err),
			log.Any("req", req),
			log.Strings("folderIDs", folderIDs),
			log.Any("operator", operator))
		return err
	}

	//check share folders valid
	orgsMap, err := f.checkShareFolders(ctx, orgIDs, folderIDs, folders, operator)
	if err != nil {
		log.Error(ctx, "checkShareFolders failed",
			log.Err(err),
			log.Any("req", req),
			log.Strings("folderIDs", folderIDs),
			log.Any("operator", operator))
		return err
	}

	//fetch pending share orgs, which orgs should be added share & which should be deleted share
	sharedFolderPendingOrgsMap, err := f.fetchPendingShareOrgs(ctx, folderIDs, orgIDs, orgsMap, operator)
	if err != nil {
		log.Error(ctx, "fetchPendingShareOrgs failed",
			log.Err(err),
			log.Strings("orgIDs", orgIDs),
			log.Strings("folderIDs", folderIDs),
			log.Any("orgsMap", orgsMap))
		return err
	}
	//fetch all related contents & folder content relation
	contentFolderMap, err := f.fetchSharedContentIDs(ctx, folders, operator)
	if err != nil {
		log.Error(ctx, "fetchSharedContentIDs failed",
			log.Err(err),
			log.Any("folders", folders),
			log.Any("operator", operator))
		return err
	}

	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		//5.Remove folder share records & Remove content auth records & Get contents from folders
		err := f.removeSharedFolderAndAuthedContent(ctx, tx,
			&entity.HandleSharedFolderAndAuthedContentRequest{
				SharedFolderPendingOrgsMap: sharedFolderPendingOrgsMap,
				ShareFolders:               folders,
				ContentFolderMap:           contentFolderMap,
			}, operator)
		if err != nil {
			return err
		}
		log.Info(ctx, "allContentIDs",
			log.Any("allContentIDs", contentFolderMap))
		//6.Add folder share records & authed contents
		err = f.addSharedFolderAndAuthedContent(ctx, tx, &entity.HandleSharedFolderAndAuthedContentRequest{
			SharedFolderPendingOrgsMap: sharedFolderPendingOrgsMap,
			ShareFolders:               folders,
			ContentFolderMap:           contentFolderMap,
		}, operator)
		if err != nil {
			return err
		}
		return nil
	})
}

func (f *FolderModel) addSharedFolderAndAuthedContent(ctx context.Context, tx *dbo.DBContext,
	req *entity.HandleSharedFolderAndAuthedContentRequest,
	operator *entity.Operator) error {
	log.Info(ctx, "building recordsData",
		log.Any("folders", req.ShareFolders),
		log.Any("sharedFolderPendingOrgsMap", req.SharedFolderPendingOrgsMap))

	//Batch add share folders
	recordsData := make([]*entity.SharedFolderRecord, 0)
	for i := range req.ShareFolders {
		for j := range req.SharedFolderPendingOrgsMap[req.ShareFolders[i].ID].AddOrgs {
			recordsData = append(recordsData, &entity.SharedFolderRecord{
				FolderID: req.ShareFolders[i].ID,
				OrgID:    req.SharedFolderPendingOrgsMap[req.ShareFolders[i].ID].AddOrgs[j],
				Creator:  operator.UserID,
			})
		}
	}
	log.Info(ctx, "recordsData",
		log.Any("recordsData", recordsData))
	if len(recordsData) > 0 {
		err := da.GetSharedFolderDA().BatchAdd(ctx, tx, recordsData)
		if err != nil {
			log.Error(ctx, "Batch add shared folder failed",
				log.Err(err),
				log.Any("recordsData", recordsData),
				log.Any("operator", operator))
			return err
		}
	}

	//Batch add content auth records
	log.Info(ctx, "records authed Data",
		log.Any("folders", req.ShareFolders),
		log.Any("allContentIDs", req.ContentFolderMap.ContentFolderMap),
		log.Any("sharedFolderPendingOrgsMap", req.SharedFolderPendingOrgsMap))
	authData := make([]*entity.AddAuthedContentRequest, 0)
	for i := range req.ShareFolders {
		for j := range req.ContentFolderMap.ContentFolderMap[req.ShareFolders[i].ID] {
			for k := range req.SharedFolderPendingOrgsMap[req.ShareFolders[i].ID].AddOrgs {
				authData = append(authData, &entity.AddAuthedContentRequest{
					FromFolderID: req.ShareFolders[i].ID,
					ContentID:    req.ContentFolderMap.ContentFolderMap[req.ShareFolders[i].ID][j],
					OrgID:        req.SharedFolderPendingOrgsMap[req.ShareFolders[i].ID].AddOrgs[k],
				})
			}
		}
	}

	log.Info(ctx, "authData",
		log.Any("authData", authData))
	if len(authData) > 0 {
		err := GetAuthedContentRecordsModel().BatchAddByOrgIDs(ctx, tx, authData, operator)
		if err != nil {
			log.Error(ctx, "Batch add auth contents failed",
				log.Err(err),
				log.Any("authData", authData),
				log.Any("operator", operator))
			return err
		}
	}
	return nil
}

func (f *FolderModel) fetchSharedContentIDs(ctx context.Context, folders []*entity.FolderItem, operator *entity.Operator) (*entity.ContentFolderMap, error) {
	allContentFolderMap := make(map[string][]string, 0)
	folderPathList := make([]string, len(folders))
	for i := range folders {
		folderPathList[i] = folders[i].ChildrenPath().String()
	}

	//4.Get contents from folders
	condition := &da.ContentConditionInternal{
		ContentType: []int{entity.ContentTypePlan, entity.ContentTypeMaterial},
		DirPath:     entity.NullStrings{Strings: folderPathList, Valid: true},
	}
	_, contents, err := da.GetContentDA().SearchContentInternal(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		log.Error(ctx, "Search folder items failed",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("operator", operator))
		return nil, err
	}

	contentIDs := make([]string, len(contents))
	for i := range contents {
		allContentFolderMap[contents[i].DirPath.Parent()] = append(allContentFolderMap[contents[i].DirPath.Parent()], contents[i].ID)
		contentIDs[i] = contents[i].ID
	}
	return &entity.ContentFolderMap{
		ContentFolderMap: allContentFolderMap,
		AllContentIDs:    contentIDs,
	}, nil
}

func (f *FolderModel) removeSharedFolderAndAuthedContent(ctx context.Context, tx *dbo.DBContext,
	req *entity.HandleSharedFolderAndAuthedContentRequest, operator *entity.Operator) error {
	allContentFolderMap := make(map[string][]string, 0)

	//Remove folder share records
	for folderID, pendingOrgList := range req.SharedFolderPendingOrgsMap {
		if len(pendingOrgList.DeleteOrgs) > 0 {
			err := da.GetSharedFolderDA().BatchDeleteByOrgIDs(ctx, tx, folderID, pendingOrgList.DeleteOrgs)
			if err != nil {
				log.Error(ctx, "Batch delete folder failed",
					log.Err(err),
					log.String("folderID", folderID),
					log.Any("item", pendingOrgList),
					log.Any("operator", operator))
				return err
			}
		}
	}

	//Remove content share records

	//fetch all folders has content
	for i := range req.ShareFolders {
		//Judge if folder should delete share org
		if len(req.SharedFolderPendingOrgsMap[req.ShareFolders[i].ID].DeleteOrgs) < 1 {
			continue
		}
		//remove authed content when it should be deleted
		if len(req.ContentFolderMap.ContentFolderMap[req.ShareFolders[i].ID]) > 0 {
			err := GetAuthedContentRecordsModel().BatchDelete(ctx, tx, entity.BatchDeleteAuthedContentByOrgsRequest{
				OrgIDs:     req.SharedFolderPendingOrgsMap[req.ShareFolders[i].ID].DeleteOrgs,
				FolderIDs:  []string{req.ShareFolders[i].ID},
				ContentIDs: allContentFolderMap[req.ShareFolders[i].ID],
			}, operator)
			if err != nil {
				log.Error(ctx, "Batch delete auth content failed",
					log.Err(err),
					log.Strings("orgIDs", req.SharedFolderPendingOrgsMap[req.ShareFolders[i].ID].DeleteOrgs),
					log.Strings("contentIDs", allContentFolderMap[req.ShareFolders[i].ID]),
					log.Any("operator", operator))
				return err
			}
		}
	}

	return nil
}

func (f *FolderModel) getFolderPendingOrgs(ctx context.Context,
	sharedFolderOrgsMap map[string][]string, orgsMap map[string]bool,
	orgIDs []string, folderIDs []string) map[string]*entity.ShareFoldersDeleteAddOrgList {
	sharedFolderPendingOrgsMap := make(map[string]*entity.ShareFoldersDeleteAddOrgList)
	//handle every folder shared orgs
	for fid, folderSharedOrgs := range sharedFolderOrgsMap {
		//get pending delete orgs
		//delete org is orgs not exist in req.OrgIDs
		for i := range folderSharedOrgs {
			_, ok := orgsMap[folderSharedOrgs[i]]
			//if org in current records, but not in req.org
			//remove the org from records
			if !ok {
				//if orgsList is not exist, create a new one
				orgsList, existOrgsList := sharedFolderPendingOrgsMap[fid]
				if !existOrgsList {
					orgsList = new(entity.ShareFoldersDeleteAddOrgList)
				}
				//add pending delete orgs
				orgsList.DeleteOrgs = append(orgsList.DeleteOrgs, folderSharedOrgs[i])
				sharedFolderPendingOrgsMap[fid] = orgsList
			}
		}
		//get pending add orgs
		for i := range orgIDs {
			flag := false
			//check if folder already has the org
			for j := range folderSharedOrgs {
				if folderSharedOrgs[j] == orgIDs[i] {
					flag = true
					break
				}
			}

			//if already has org, don't add, else add to pending add orgs
			if !flag {
				orgsList, existOrgsList := sharedFolderPendingOrgsMap[fid]
				if !existOrgsList {
					orgsList = new(entity.ShareFoldersDeleteAddOrgList)
				}
				//add pending delete orgs
				orgsList.AddOrgs = append(orgsList.AddOrgs, orgIDs[i])
				sharedFolderPendingOrgsMap[fid] = orgsList
			}
		}
	}

	//when sharedFolderPendingOrgsMap doesn't contains folder, add all orgs
	for i := range folderIDs {
		_, exist := sharedFolderPendingOrgsMap[folderIDs[i]]
		if !exist {
			sharedFolderPendingOrgsMap[folderIDs[i]] = &entity.ShareFoldersDeleteAddOrgList{
				AddOrgs: orgIDs,
			}
		}
	}
	return sharedFolderPendingOrgsMap
}

func (f *FolderModel) checkOrgs(ctx context.Context, orgIDs []string, operator *entity.Operator) (map[string]bool, error) {
	//Get orgs by ids
	hasShareAll := false
	validOrgs := make([]string, 0, len(orgIDs))
	for i := range orgIDs {
		if orgIDs[i] != constant.ShareToAll {
			validOrgs = append(validOrgs, orgIDs[i])
		} else {
			hasShareAll = true
		}
	}
	orgs, err := external.GetOrganizationServiceProvider().BatchGet(ctx, operator, validOrgs)
	if err != nil {
		log.Error(ctx, "Get orgs failed",
			log.Err(err),
			log.Strings("orgIDs", orgIDs),
			log.Any("operator", operator))
		return nil, err
	}
	//check if all orgs are exist
	orgsMap := make(map[string]bool)
	if hasShareAll {
		orgsMap[constant.ShareToAll] = true
	}
	for i := range orgs {
		if orgs[i].Valid {
			orgsMap[orgs[i].ID] = true
		}
	}
	return orgsMap, nil
}

func (f *FolderModel) UpdateFolder(ctx context.Context, folderID string, d entity.UpdateFolderRequest, operator *entity.Operator) error {
	//get folder object
	folder, err := f.getFolder(ctx, dbo.MustGetDB(ctx), folderID)
	if err != nil {
		return err
	}
	//if name is updated, check duplicate
	if d.Name != "" && d.Name != folder.Name {
		err = f.checkDuplicateFolderNameForUpdate(ctx, d.Name, folder, operator)
		if err != nil {
			return err
		}
		//Lock folder name
		locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixFolderName, d.Name)
		if err != nil {
			log.Error(ctx, "get locker failed",
				log.Err(err),
				log.String("prefix", da.RedisKeyPrefixFolderName),
				log.String("name", d.Name))
			return err
		}
		locker.Lock()
		defer locker.Unlock()
	}

	folder.Name = d.Name
	folder.Thumbnail = d.Thumbnail
	folder.Editor = operator.UserID
	folder.Description = d.Description
	folder.Keywords = strings.Join(d.Keywords, constant.StringArraySeparator)
	err = da.GetFolderDA().UpdateFolder(ctx, dbo.MustGetDB(ctx), folderID, folder)
	if err != nil {
		log.Error(ctx, "update folder item failed", log.Err(err), log.Any("folder", folder))
		return err
	}
	return nil
}

func (f *FolderModel) RemoveItem(ctx context.Context, fid string, operator *entity.Operator) error {
	folderItem, err := f.getFolder(ctx, dbo.MustGetDB(ctx), fid)
	if err != nil {
		return err
	}

	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := f.removeItemInternal(ctx, tx, folderItem)
		if err != nil {
			return err
		}
		err = f.batchRepairFolderItemsCountByIDs(ctx, tx, []string{folderItem.ParentID})
		if err != nil {
			return err
		}
		return nil
	})
}

func (f *FolderModel) RemoveItemBulk(ctx context.Context, fids []string, operator *entity.Operator) error {
	folders, err := da.GetFolderDA().GetFolderByIDList(ctx, dbo.MustGetDB(ctx), fids)
	if err != nil {
		log.Error(ctx, "GetFolderByIDList failed", log.Err(err), log.Strings("fids", fids))
		return err
	}
	parentIDs := make([]string, len(folders))
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		for i := range folders {
			parentIDs[i] = folders[i].ParentID
		}
		//remove folder
		err = f.batchRemoveItemInternal(ctx, tx, folders)
		if err != nil {
			log.Error(ctx, "batchRemoveItemInternal failed", log.Err(err), log.Any("folder", folders))
			return err
		}

		//repair folder items
		err = f.batchRepairFolderItemsCountByIDs(ctx, tx, parentIDs)
		if err != nil {
			return err
		}
		return nil
	})
}

func (f *FolderModel) MoveItemBulk(ctx context.Context, req entity.MoveFolderIDBulkRequest, operator *entity.Operator) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		distFolder, err := f.getFolderMaybeRoot(ctx, tx, req.Dist, req.OwnerType, req.Partition, operator)
		if err != nil {
			return err
		}
		err = f.batchMoveItem(ctx, tx, req.FolderInfo, req.OwnerType, req.Partition, distFolder, operator)
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

func (f *FolderModel) MoveItem(ctx context.Context, req entity.MoveFolderRequest, operator *entity.Operator) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		distFolder, err := f.getFolderMaybeRoot(ctx, tx, req.Dist, req.OwnerType, req.Partition, operator)
		if err != nil {
			return err
		}
		err = f.moveItem(ctx, tx,
			req.OwnerType,
			req.FolderFileType,
			req.Partition,
			req.ID,
			distFolder,
			operator)
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

func (f *FolderModel) ListItems(ctx context.Context, folderID string, itemType entity.ItemType, operator *entity.Operator) ([]*entity.FolderItemInfo, error) {
	//check owner type
	folderItems, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		ParentID: folderID,
		ItemType: int(itemType),
	})
	if err != nil {
		log.Error(ctx, "list items failed", log.Err(err), log.String("folderID", folderID))
		return nil, err
	}
	return f.folderItemToFolderItemInfoBatch(ctx, folderItems), nil
}

func (f *FolderModel) SearchFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItemInfo, error) {
	total, folderItems, err := da.GetFolderDA().SearchFolderPage(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		ParentID:  condition.ParentID,
		NameLike:  condition.Name,
		ItemType:  int(condition.ItemType),
		OwnerType: int(condition.OwnerType),
		Owner:     condition.Owner,
		Link:      condition.Link,
		Partition: entity.NewFolderPartition(condition.Partition),
		//VisibilitySetting: condition.VisibilitySetting,
		ExactDirPath: condition.Path,
		Pager:        condition.Pager,
		OrderBy:      da.NewFolderOrderBy(condition.OrderBy),
	})
	if err != nil {
		log.Error(ctx, "list items failed", log.Err(err), log.Any("condition", condition))
		return 0, nil, err
	}
	return total, f.folderItemToFolderItemInfoBatch(ctx, folderItems), nil
}

func (f *FolderModel) SearchPrivateFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItemInfo, error) {
	condition.Owner = operator.UserID
	condition.OwnerType = entity.OwnerTypeUser
	return f.SearchFolder(ctx, condition, operator)
}

func (f *FolderModel) SearchOrgFolder(ctx context.Context, condition entity.SearchFolderCondition, operator *entity.Operator) (int, []*entity.FolderItemInfo, error) {
	condition.Owner = operator.OrgID
	condition.OwnerType = entity.OwnerTypeOrganization

	return f.SearchFolder(ctx, condition, operator)
}

func (f *FolderModel) getParentFromPath(ctx context.Context, path string) string {
	if path == "" || path == constant.FolderRootPath {
		return constant.FolderRootPath
	}
	pathDirs := strings.Split(path, constant.FolderPathSeparator)
	if len(pathDirs) < 1 {
		log.Info(ctx, "check folder exists with array 0",
			log.Strings("pathDirs", pathDirs))
		return constant.FolderRootPath
	}
	parentID := pathDirs[len(pathDirs)-1]
	return parentID
}

func (f *FolderModel) MustGetPath(ctx context.Context, tx *dbo.DBContext, ownerType entity.OwnerType, itemType entity.ItemType, path string, partition entity.FolderPartition, operator *entity.Operator) (string, error) {
	if path == "" || path == constant.FolderRootPath {
		log.Info(ctx, "check folder exists with nil",
			log.String("path", path))
		return constant.FolderRootPath, nil
	}
	pathDirs := strings.Split(path, constant.FolderPathSeparator)
	if len(pathDirs) < 1 {
		log.Info(ctx, "check folder exists with array 0",
			log.Strings("pathDirs", pathDirs))
		return constant.FolderRootPath, nil
	}

	parentID := pathDirs[len(pathDirs)-1]
	condition := da.FolderCondition{
		IDs:       entity.NullStrings{Strings: []string{parentID}, Valid: true},
		OwnerType: int(ownerType),
		Owner:     ownerType.Owner(operator),
		Partition: partition,
		ItemType:  int(itemType),
	}

	folders, err := da.GetFolderDA().SearchFolder(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "search folder failed",
			log.Err(err),
			log.Any("condition", condition))
		return constant.FolderRootPath, err
	}
	log.Info(ctx, "search folder count",
		log.Any("condition", condition),
		log.Any("folders", folders))
	if len(folders) < 1 {
		log.Info(ctx, "search folder response no folders",
			log.Err(err),
			log.Any("condition", condition))
		return constant.FolderRootPath, nil
	}
	return folders[0].ChildrenPath().AsParentPath(), nil
}

func (f *FolderModel) GetFolderByID(ctx context.Context, folderID string, operator *entity.Operator) (*entity.FolderItemInfo, error) {
	folderItem, err := da.GetFolderDA().GetFolderByID(ctx, dbo.MustGetDB(ctx), folderID)
	if err != nil {
		log.Error(ctx, "get folder by id failed", log.Err(err), log.String("folderID", folderID))
		return nil, ErrNoFolder
	}

	result := f.folderItemToFolderItemInfo(ctx, folderItem)

	userIDs := []string{result.Creator, result.Editor}
	users, err := external.GetUserServiceProvider().BatchGet(ctx, operator, userIDs)
	if err != nil {
		log.Error(ctx, "get user name failed",
			log.Err(err), log.Any("folderItem", result))
		return nil, ErrNoFolder
	}
	for i := range users {
		if users[i].Valid && users[i].ID == result.Creator {
			result.CreatorName = users[i].Name
		}
		if users[i].Valid && users[i].ID == result.Editor {
			result.EditorName = users[i].Name
		}
	}
	//if folder item is folder, add children items
	if folderItem.ItemType.IsFolder() {
		_, folderItems, err := f.SearchFolder(ctx, entity.SearchFolderCondition{
			ParentID: folderItem.ID,
		}, operator)
		if err != nil {
			log.Warn(ctx, "search folder by id failed", log.Err(err), log.String("folderID", folderID))
			return nil, err
		}

		//calculate avalible
		contentTypeList := []int{entity.ContentTypeAssets, entity.AliasContentTypeFolder}
		if folderItem.Partition == entity.FolderPartitionMaterialAndPlans {
			contentTypeList = []int{entity.ContentTypePlan, entity.ContentTypeMaterial, entity.AliasContentTypeFolder}
		}
		condition := entity.ContentConditionRequest{
			ContentType:   contentTypeList,
			PublishStatus: []string{entity.ContentStatusPublished},
			DirPath:       string(folderItem.ChildrenPath()),
		}
		err = GetContentFilterModel().FilterPublishContent(ctx, &condition, operator)
		if err != nil {
			log.Warn(ctx, "FilterPublishContent failed", log.Err(err), log.Any("condition", condition))
			return nil, err
		}
		total, err := GetContentModel().CountUserFolderContent(ctx, dbo.MustGetDB(ctx), &condition, operator)
		if err != nil {
			log.Warn(ctx, "count folder failed failed", log.Err(err), log.Any("condition", condition))
			return nil, err
		}
		result.Available = total
		result.Items = folderItems
	}
	return result, nil
}

func (f *FolderModel) checkMoveFolder(ctx context.Context, folder *entity.FolderItem, distFolder *entity.FolderItem) error {
	if !folder.ItemType.IsFolder() {
		log.Warn(ctx, "move folder must be a folder", log.Any("folder", folder))
		return ErrInvalidFolderItemType
	}
	if folder.ID == distFolder.ID {
		log.Warn(ctx, "move to the same folder",
			log.Any("dist", distFolder),
			log.Any("folder", folder))
	}
	if folder.Partition != distFolder.Partition {
		log.Warn(ctx, "move folder must be the same partition",
			log.Any("dist", distFolder),
			log.Any("folder", folder))
		return ErrInvalidPartition
	}
	//检查参数是否有问题
	//check params
	err := f.checkMoveItem(ctx, folder, distFolder)
	if err != nil {
		return err
	}
	return nil
}

func (f *FolderModel) checkMoveItem(ctx context.Context, folder *entity.FolderItem, distFolder *entity.FolderItem) error {
	//check if dist is folder children
	if folder.ItemType.IsFolder() && distFolder.DirPath.IsChild(folder.ID) {
		//distFolder.DirPath
		return ErrMoveToChild
	}
	log.Info(ctx, "check dir",
		log.Any("distFolder", distFolder),
		log.Any("folder", folder),
		log.String("children path", string(folder.ChildrenPath())))

	if distFolder.DirPath == folder.ChildrenPath() {
		//distFolder.DirPath
		log.Error(ctx, "check dir",
			log.Err(ErrMoveToSameFolder),
			log.Any("distFolder", distFolder),
			log.Any("folder", folder),
			log.String("children path", string(folder.ChildrenPath())))
		return ErrMoveToSameFolder
	}

	return nil
}

func (f *FolderModel) checkHeadquarterRegionOrganizations(ctx context.Context, orgIDs []string, operator *entity.Operator) error {
	orgProperty, err := GetOrganizationPropertyModel().MustGet(ctx, operator.OrgID)
	if err != nil {
		log.Error(ctx, "Get organization properties failed",
			log.Err(err),
			log.Strings("orgIDs", orgIDs),
			log.Any("operator", operator))
		return err
	}
	if orgProperty.Type != entity.OrganizationTypeHeadquarters {
		log.Warn(ctx, "Org is not headquarters",
			log.Strings("orgIDs", orgIDs),
			log.Any("orgProperty", orgProperty),
			log.Any("operator", operator))
		return ErrNotHeadquartersShare
	}
	if orgProperty.Region == entity.UnknownRegion {
		//unknown org order
		log.Warn(ctx, "unknown region",
			log.Strings("orgIDs", orgIDs),
			log.Any("orgProperty", orgProperty),
			log.Any("operator", operator))
		return ErrUnknownHeadquarterRegion
	} else if orgProperty.Region == entity.Global {
		//global headquarters can share to any org
		return nil
	}

	regionOrgIDs, err := GetOrganizationRegionModel().GetOrganizationByHeadquarter(ctx, dbo.MustGetDB(ctx), operator.OrgID)
	if err != nil {
		return err
	}
	regionOrgMap := make(map[string]bool)
	for i := range regionOrgIDs {
		regionOrgMap[regionOrgIDs[i]] = true
	}
	for i := range orgIDs {
		_, ok := regionOrgMap[orgIDs[i]]
		if !ok {
			log.Error(ctx, "Share to unsupported region",
				log.Err(err),
				log.Strings("orgIDs", orgIDs),
				log.Strings("regionOrgIDs", regionOrgIDs),
				log.Any("operator", operator))
			return ErrShareToUnsupportedRegion
		}
	}

	return nil
}

func (f *FolderModel) fetchPendingShareOrgs(ctx context.Context,
	folderIDs []string,
	orgIDs []string,
	orgsMap map[string]bool,
	operator *entity.Operator) (map[string]*entity.ShareFoldersDeleteAddOrgList, error) {
	//3.search shared folders from database
	//for removing and adding pending modified organizations list
	records, err := da.GetSharedFolderDA().Search(ctx, dbo.MustGetDB(ctx), da.SharedFolderCondition{
		FolderIDs: folderIDs,
	})
	if err != nil {
		log.Error(ctx, "Search records failed",
			log.Err(err),
			log.Strings("orgIDs", orgIDs),
			log.Strings("folderIDs", folderIDs),
			log.Any("operator", operator))
		return nil, ErrSearchSharedFolderFailed
	}
	log.Info(ctx, "check org maps",
		log.Strings("FolderIDs", folderIDs),
		log.Any("records", records))
	//build folders already shared organizations map
	//mapping folder ids and shared organizations
	sharedFolderOrgsMap := make(map[string][]string)
	for i := range records {
		sharedFolderOrgsMap[records[i].FolderID] = append(sharedFolderOrgsMap[records[i].FolderID], records[i].OrgID)
	}
	//4.get pending add folders in orgs & pending delete folders in orgs
	sharedFolderPendingOrgsMap := f.getFolderPendingOrgs(ctx, sharedFolderOrgsMap, orgsMap, orgIDs, folderIDs)

	log.Info(ctx, "getFolderPendingOrgs",
		log.Any("sharedFolderPendingOrgsMap", sharedFolderPendingOrgsMap),
		log.Any("sharedFolderOrgsMap", sharedFolderOrgsMap))
	return sharedFolderPendingOrgsMap, nil
}

func (f *FolderModel) checkShareFolders(ctx context.Context, orgIDs []string, shareFolderIDs []string,
	shareFolders []*entity.FolderItem, operator *entity.Operator) (map[string]bool, error) {
	//check if all folders are exist
	if len(shareFolders) != len(shareFolderIDs) {
		log.Warn(ctx, "Some folders not found",
			log.Strings("folderIDs", shareFolderIDs),
			log.Any("folders", shareFolders),
			log.Any("operator", operator))
		return nil, ErrNoFolder
	}

	//check if all folders partition are right
	for i := range shareFolders {
		if shareFolders[i].Partition != entity.FolderPartitionMaterialAndPlans {
			log.Error(ctx, "invalid partition",
				log.Err(ErrInvalidPartition),
				log.Any("folders[i]", shareFolders[i]),
				log.Strings("folderIDs", shareFolderIDs),
				log.Any("operator", operator))
			return nil, ErrInvalidPartition
		}
	}
	log.Info(ctx, "GetFolderByIDList",
		log.Strings("folderIDs", shareFolderIDs),
		log.Any("folders", shareFolders))

	//2.Check organizations validation
	orgsMap, err := f.checkOrgs(ctx, orgIDs, operator)
	if err != nil {
		log.Error(ctx, "Get orgs failed",
			log.Err(err),
			log.Strings("orgIDs", orgIDs),
			log.Strings("folderIDs", shareFolderIDs),
			log.Any("operator", operator))
		return nil, err
	}
	log.Info(ctx, "check org maps",
		log.Strings("orgIDs", orgIDs),
		log.Any("orgsMap", orgsMap))
	return orgsMap, nil
}
func (f *FolderModel) batchRepairFolderItemsCount(ctx context.Context, tx *dbo.DBContext, items []*entity.FolderItem) error {
	itemsCountMap, err := f.batchGetFolderItemsCountMap(ctx, tx, items)
	if err != nil {
		return err
	}
	req := make([]*entity.UpdateFolderItemsCountRequest, 0)
	for i := range items {
		itemsCount := 0
		if count, exists := itemsCountMap[items[i].ID]; exists {
			itemsCount = count
		}

		if items[i].ItemsCount != itemsCount {
			req = append(req, &entity.UpdateFolderItemsCountRequest{
				ID:    items[i].ID,
				Count: itemsCount,
			})
		}
	}
	//No need to update items count
	if len(req) < 1 {
		return nil
	}
	err = da.GetFolderDA().BatchUpdateFolderItemsCount(ctx, tx, req)
	if err != nil {
		log.Error(ctx, "BatchUpdateFolderItemsCount failed",
			log.Err(err),
			log.Any("req", req),
			log.Any("items", items))
		return err
	}
	return nil
}

func (f *FolderModel) batchRepairFolderItemsCountByIDs(ctx context.Context, tx *dbo.DBContext, fids []string) error {
	items, err := da.GetFolderDA().GetFolderByIDList(ctx, tx, fids)
	if err != nil {
		log.Error(ctx, "GetFolderByIDList failed",
			log.Err(err),
			log.Strings("fids", fids))
		return err
	}
	return f.batchRepairFolderItemsCount(ctx, tx, items)
}

func (f *FolderModel) batchGetFolderItemsCountMap(ctx context.Context, tx *dbo.DBContext, items []*entity.FolderItem) (map[string]int, error) {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	ids = utils.SliceDeduplication(ids)

	itemCountRes, err := da.GetFolderDA().BatchGetFolderItemsCount(ctx, tx, ids)
	if err != nil {
		log.Error(ctx, "BatchGetFolderItemsCount failed",
			log.Err(err),
			log.Any("items", items),
			log.Strings("ids", ids))
		return nil, err
	}
	res := make(map[string]int)
	for i := range itemCountRes {
		_, exists := res[itemCountRes[i].ID]
		count := 0
		if exists {
			count = res[itemCountRes[i].ID]
		}
		res[itemCountRes[i].ID] = count + int(itemCountRes[i].Count)
	}
	return res, nil
}
func (f *FolderModel) folderItemToFolderItemInfo(ctx context.Context, item *entity.FolderItem) *entity.FolderItemInfo {

	return &entity.FolderItemInfo{
		ID:          item.ID,
		OwnerType:   item.OwnerType,
		Owner:       item.Owner,
		ParentID:    item.ParentID,
		Link:        item.Link,
		ItemType:    item.ItemType,
		DirPath:     item.DirPath,
		Partition:   item.Partition,
		Name:        item.Name,
		Description: item.Description,
		Keywords:    strings.Split(item.Keywords, constant.StringArraySeparator),
		Thumbnail:   item.Thumbnail,
		Creator:     item.Creator,
		ItemsCount:  item.ItemsCount,
		Editor:      item.Editor,
		CreateAt:    item.CreateAt,
		UpdateAt:    item.UpdateAt,
		Items:       nil,
	}
}

func (f *FolderModel) folderItemToFolderItemInfoBatch(ctx context.Context, items []*entity.FolderItem) []*entity.FolderItemInfo {
	ret := make([]*entity.FolderItemInfo, len(items))

	for i := range items {
		ret[i] = &entity.FolderItemInfo{
			ID:          items[i].ID,
			OwnerType:   items[i].OwnerType,
			Owner:       items[i].Owner,
			ParentID:    items[i].ParentID,
			Link:        items[i].Link,
			ItemType:    items[i].ItemType,
			DirPath:     items[i].DirPath,
			Partition:   items[i].Partition,
			Name:        items[i].Name,
			Description: items[i].Description,
			Keywords:    strings.Split(items[i].Keywords, constant.StringArraySeparator),
			Thumbnail:   items[i].Thumbnail,
			Creator:     items[i].Creator,
			ItemsCount:  items[i].ItemsCount,
			Editor:      items[i].Editor,
			CreateAt:    items[i].CreateAt,
			UpdateAt:    items[i].UpdateAt,
			Items:       nil,
		}
	}
	return ret
}
func (f *FolderModel) getParentFolderListByContentIDs(ctx context.Context, tx *dbo.DBContext, ids []string) ([]*entity.FolderItem, error) {

	_, contentList, err := da.GetContentDA().SearchContentInternal(ctx, tx,
		&da.ContentConditionInternal{
			IDS: entity.NullStrings{
				Strings: ids,
				Valid:   true,
			}})
	if err != nil {
		log.Warn(ctx, "SearchContentInternal failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}
	folderIDs := make([]string, len(contentList))
	for i := range contentList {
		folderIDs[i] = contentList[i].DirPath.Parent()
	}
	folderIDs = utils.SliceDeduplication(folderIDs)
	folders, err := da.GetFolderDA().GetFolderByIDList(ctx, tx, folderIDs)
	if err != nil {
		log.Warn(ctx, "GetFolderByIDList failed",
			log.Err(err),
			log.Strings("ids", ids),
			log.Strings("folderIDs", folderIDs))
		return nil, err
	}

	return folders, nil
}

func (f *FolderModel) handleMoveContentByLink(ctx context.Context,
	tx *dbo.DBContext,
	ownerType entity.OwnerType,
	ids []string,
	partition entity.FolderPartition,
	distFolder *entity.FolderItem,
	operator *entity.Operator) error {

	//get parent folders
	folders, err := f.getParentFolderListByContentIDs(ctx, tx, ids)
	if err != nil {
		log.Warn(ctx, "getParentFolderListByContentIDs failed",
			log.Err(err),
			log.Strings("ids", ids))
		return err
	}
	//add root folder
	rootFolder := f.rootFolder(ctx, ownerType, partition, operator)
	folders = append(folders, rootFolder)

	//更新子目录link文件
	//Update links in the descendant folders
	err = f.updateLinkedItemPath(ctx, tx,
		ids,
		distFolder.ChildrenPath(),
		folders, distFolder, operator)
	if err != nil {
		log.Warn(ctx, "update notify move item distPathpath failed",
			log.Err(err),
			log.Strings("ids", ids))
		return err
	}

	//更新文件数量
	//Update file count
	folderIDs := make([]string, len(folders)+1)
	for i := range folders {
		folderIDs[i] = folders[i].ID
	}
	folderIDs[len(folderIDs)-1] = distFolder.ID
	err = f.updateMoveFolderItemCount(ctx, tx, folderIDs)
	if err != nil {
		return err
	}

	return nil
}

func (f *FolderModel) handleMoveFolder(ctx context.Context, tx *dbo.DBContext,
	fid string,
	distFolder *entity.FolderItem,
	operator *entity.Operator) error {

	folder, err := f.getFolder(ctx, tx, fid)
	if err != nil {
		return err
	}
	//检查参数是否有问题
	//check params
	err = f.checkMoveFolder(ctx, folder, distFolder)
	if err != nil {
		return err
	}

	//获取目录下的所有文件（所有子文件一起移动）
	//get all items from the folder
	descendantFolderIDs, err := f.getDescendantItemsIDs(ctx, folder)
	if err != nil {
		return err
	}

	//list all content in the folder
	linkOriginPath := folder.ChildrenPath()
	contentIDs, err := f.listContentInFolder(ctx, tx, linkOriginPath.String())
	if err != nil {
		log.Warn(ctx, "listContentInFolder failed",
			log.Err(err), log.Strings("ids", descendantFolderIDs),
			log.String("linkOriginPath", linkOriginPath.String()))
		return err
	}

	//Update all content in the folder
	//replace the path of all content in the folder
	err = f.replaceLinkedContentPath(ctx, tx, &entity.ReplaceLinkedContentPathRequest{
		FromFolder: folder,
		DistFolder: distFolder,
		ContentIDs: contentIDs,
	})
	if err != nil {
		log.Warn(ctx, "replaceLinkedContentPath failed", log.Err(err),
			log.Any("distFolder", distFolder),
			log.String("linkOriginPath", linkOriginPath.String()),
			log.Strings("contentIDs", contentIDs),
			log.Any("FromFolder", folder))
		return err
	}
	//更新当前目录
	//update current folder
	originPath := folder.DirPath
	originParentID := folder.ParentID
	//Move folder
	err = f.doMoveFolder(ctx, tx, folder, distFolder)
	if err != nil {
		log.Error(ctx, "doMoveFolder failed", log.Err(err),
			log.Any("distFolder", distFolder),
			log.Any("folder", folder))
		return err
	}
	//Move sub folders
	err = f.updateSubFolderItems(ctx, tx, folder, descendantFolderIDs, originPath)
	if err != nil {
		log.Error(ctx, "updateSubFolderItems failed", log.Err(err),
			log.Strings("ids", descendantFolderIDs),
			log.String("origin path", originPath.String()),
			log.Any("folder", folder))
		return err
	}

	//更新文件数量
	//update item count
	err = f.updateMoveFolderItemCount(ctx, tx, []string{originParentID, distFolder.ID})
	if err != nil {
		return err
	}

	//Handle shared folder content
	err = f.handleMoveSharedContentFolderRecursion(ctx, tx, contentIDs, folder, distFolder, operator)
	if err != nil {
		log.Error(ctx, "can't handle move shared content",
			log.Err(err),
			log.Any("toFolder", distFolder),
			log.Strings("contentLinkIds", contentIDs),
			log.Any("folder", folder))
		return err
	}
	return nil
}

func (f *FolderModel) handleMoveFolders(ctx context.Context, tx *dbo.DBContext,
	fids []string,
	distFolder *entity.FolderItem,
	operator *entity.Operator) error {

	folders, err := da.GetFolderDA().GetFolderByIDList(ctx, tx, fids)
	if err != nil {
		log.Error(ctx, "replaceLinkedContentPath failed", log.Err(err),
			log.Strings("fids", fids))
		return err
	}
	//检查参数是否有问题
	//check params
	for i := range folders {
		err = f.checkMoveFolder(ctx, folders[i], distFolder)
		if err != nil {
			return err
		}
	}

	//获取目录下的所有文件（所有子文件一起移动）
	//get all items from the folder
	descendantFolderMap, err := f.getDescendantItemsMapByFolders(ctx, folders)
	if err != nil {
		return err
	}

	//list all content in the folder
	folderContentMap, err := f.listContentMapInFolders(ctx, tx, descendantFolderMap.FolderChildrenPathMap)
	if err != nil {
		log.Warn(ctx, "listContentInFolder failed",
			log.Err(err),
			log.Any("descendantFolderMap", descendantFolderMap))
		return err
	}

	//Save origin parent id list for update item counts
	originParentIDList := make([]string, len(folders)+1)

	//Update all content in the folder
	//replace the path of all content in the folder
	//TODO:Execute one statement per folder, Maybe can Accelerate
	for i := range folders {
		originParentIDList[i] = folders[i].ParentID
		err = f.replaceLinkedContentPath(ctx, tx, &entity.ReplaceLinkedContentPathRequest{
			FromFolder: folders[i],
			DistFolder: distFolder,
			ContentIDs: folderContentMap[folders[i].ID],
		})
		if err != nil {
			log.Warn(ctx, "replaceLinkedContentPath failed", log.Err(err),
				log.Any("distFolder", distFolder),
				log.Strings("contentIDs", folderContentMap[folders[i].ID]),
				log.Any("FromFolder", folders[i]))
			return err
		}
	}

	originParentIDList[len(originParentIDList)-1] = distFolder.ID

	//更新当前目录
	//update current folder
	//Move folder
	err = f.batchDoMoveFolder(ctx, tx, fids, distFolder)
	if err != nil {
		log.Error(ctx, "doMoveFolder failed", log.Err(err),
			log.Any("distFolder", distFolder),
			log.Strings("fids", fids))
		return err
	}

	//Move sub folders
	//TODO:Execute one statement per folder, Maybe can Accelerate
	for i := range folders {
		//collect sub folder ids
		subFolders := descendantFolderMap.DescendantFoldersMap[folders[i].ID]
		descendantIDs := make([]string, len(subFolders))
		for i := range subFolders {
			descendantIDs[i] = subFolders[i].ID
		}

		err = f.updateSubFolderItems(ctx, tx, folders[i],
			descendantIDs,
			descendantFolderMap.FolderChildrenPathMap[folders[i].ID])
		if err != nil {
			log.Error(ctx, "updateSubFolderItems failed", log.Err(err),
				log.Strings("ids", descendantIDs),
				log.String("origin path", descendantFolderMap.FolderChildrenPathMap[folders[i].ID].String()),
				log.Any("folder", folders[i]))
			return err
		}

		//Handle shared folder content
		err = f.handleMoveSharedContentFolderRecursion(ctx, tx,
			folderContentMap[folders[i].ID],
			folders[i],
			distFolder,
			operator)
		if err != nil {
			log.Error(ctx, "can't handle move shared content",
				log.Err(err),
				log.Any("toFolder", distFolder),
				log.Strings("contentLinkIds", folderContentMap[folders[i].ID]),
				log.Any("folder", folders[i]))
			return err
		}
	}

	//更新文件数量
	//update item count
	err = f.updateMoveFolderItemCount(ctx, tx, originParentIDList)
	if err != nil {
		return err
	}
	return nil
}

func (f *FolderModel) doMoveFolder(ctx context.Context, tx *dbo.DBContext, folder, distFolder *entity.FolderItem) error {
	path := distFolder.ChildrenPath()
	folder.DirPath = path
	folder.ParentID = distFolder.ID
	err := da.GetFolderDA().UpdateFolder(ctx, tx, folder.ID, folder)
	if err != nil {
		log.Error(ctx, "update folder failed", log.Err(err), log.Any("folder", folder))
		return err
	}
	return nil
}

func (f *FolderModel) batchDoMoveFolder(ctx context.Context, tx *dbo.DBContext, fids []string, distFolder *entity.FolderItem) error {
	path := distFolder.ChildrenPath()
	err := da.GetFolderDA().BatchUpdateFoldersPath(ctx, tx, fids, path)
	if err != nil {
		log.Error(ctx, "update folder failed", log.Err(err),
			log.Any("dist", distFolder),
			log.Strings("fids", fids))
		return err
	}
	return nil
}

func (f *FolderModel) updateSubFolderItems(ctx context.Context, tx *dbo.DBContext, folder *entity.FolderItem, ids []string, originPath entity.Path) error {
	newPath := folder.DirPath
	if originPath == constant.FolderRootPath {
		//to prevent target path build as //xxxx
		if newPath != constant.FolderRootPath {
			//if origin path is root("/")
			//when old path is root path("/"), replace function in mysql will replace all "/" in path
			//we prefix the new path as => new_path + origin_path
			err := da.GetFolderDA().BatchUpdateFolderPathPrefix(ctx, tx, ids, newPath)
			if err != nil {
				log.Error(ctx, "update folder path failed", log.Err(err), log.Strings("ids", ids), log.String("path", originPath.String()))
				return err
			}
		}

	} else {
		//origin: /xxx => /xxx/
		//target: /
		//to prevent target path build as //xxxx
		if newPath == constant.FolderRootPath {
			originPath = originPath + constant.FolderPathSeparator
		}
		//when old path is not root path like "/xxx/xxx"
		//replace origin path "/xxx/xxx" to target path "/yyy/yyy/yyy"
		err := da.GetFolderDA().BatchReplaceFolderPath(ctx, tx, ids, originPath, newPath)
		if err != nil {
			log.Error(ctx, "update folder path failed", log.Err(err), log.Strings("ids", ids), log.String("path", originPath.String()))
			return err
		}
	}
	return nil
}

func (f *FolderModel) moveItem(ctx context.Context, tx *dbo.DBContext, ownerType entity.OwnerType, folderFileType entity.FolderFileType, partition entity.FolderPartition, fid string, distFolder *entity.FolderItem, operator *entity.Operator) error {
	if partition != distFolder.Partition {
		log.Error(ctx, "can't move to different partition", log.Any("from", distFolder),
			log.Any("to", distFolder))
		return ErrMoveToDifferentPartition
	}
	//check if parentFolder is a folder
	if !distFolder.ItemType.IsFolder() {
		log.Warn(ctx, "move to an item not folder", log.Any("parentFolder", distFolder))
		return ErrMoveToNotFolder
	}
	log.Info(ctx, "move item",
		log.Int("ownerType", int(ownerType)),
		log.String("fileType", string(folderFileType)),
		log.String("partition", string(partition)),
		log.Any("parentFolder", distFolder),
		log.Any("user", operator),
		log.String("fid", fid))
	if fid == distFolder.ID {
		log.Warn(ctx, "move to item self", log.Any("parentFolder", distFolder), log.String("fid", fid))
		return ErrMoveToNotFolder
	}
	switch folderFileType {
	case entity.FolderFileTypeFolder:
		//若文件为Folder
		//folder case
		return f.handleMoveFolder(ctx, tx, fid, distFolder, operator)
	case entity.FolderFileTypeContent:
		//若文件为content
		//content case
		return f.handleMoveContentByLink(ctx, tx, ownerType, []string{fid}, partition, distFolder, operator)
	}
	log.Warn(ctx, "invalid folder file type", log.String("file_type", string(folderFileType)), log.Any("operator", operator))
	return ErrInvalidFolderItemType
}

func (f *FolderModel) batchMoveItem(ctx context.Context,
	tx *dbo.DBContext,
	folderTypeWithIDList []entity.FolderIdWithFileType,
	ownerType entity.OwnerType,
	partition entity.FolderPartition,
	distFolder *entity.FolderItem,
	operator *entity.Operator) error {
	if partition != distFolder.Partition {
		log.Error(ctx, "can't move to different partition", log.Any("from", distFolder),
			log.Any("to", distFolder))
		return ErrMoveToDifferentPartition
	}
	//check if parentFolder is a folder
	if !distFolder.ItemType.IsFolder() {
		log.Warn(ctx, "move to an item not folder", log.Any("parentFolder", distFolder))
		return ErrMoveToNotFolder
	}
	log.Info(ctx, "move item",
		log.Int("ownerType", int(ownerType)),
		log.String("partition", string(partition)),
		log.Any("parentFolder", distFolder),
		log.Any("user", operator),
		log.Any("folderTypeWithIDList", folderTypeWithIDList))

	//collect content ids & folder ids
	contentIDs := make([]string, 0)
	folderIDs := make([]string, 0)
	for i := range folderTypeWithIDList {
		if folderTypeWithIDList[i].FolderFileType == entity.FolderFileTypeFolder {
			folderIDs = append(folderIDs, folderTypeWithIDList[i].ID)
		} else if folderTypeWithIDList[i].FolderFileType == entity.FolderFileTypeContent {
			contentIDs = append(contentIDs, folderTypeWithIDList[i].ID)
		}
	}
	err := f.handleMoveFolders(ctx, tx, folderIDs, distFolder, operator)
	if err != nil {
		log.Error(ctx, "handleMoveFolders failed",
			log.Err(err),
			log.Strings("folderIDs", folderIDs),
			log.Any("dist", distFolder),
			log.Any("operator", operator))
		return err
	}
	err = f.handleMoveContentByLink(ctx, tx, ownerType, contentIDs, partition, distFolder, operator)
	if err != nil {
		log.Error(ctx, "handleMoveContentByLink failed",
			log.Err(err),
			log.Strings("contentIDs", contentIDs),
			log.Any("dist", distFolder),
			log.Any("operator", operator))
		return err
	}

	log.Warn(ctx, "invalid folder file type", log.Any("folderTypeWithIDList", folderTypeWithIDList), log.Any("operator", operator))
	return ErrInvalidFolderItemType
}

func (f *FolderModel) hasFolderFileItem(ctx context.Context, tx *dbo.DBContext, ownerType entity.OwnerType, partition entity.FolderPartition, owner string, link string) (bool, error) {
	if !ownerType.Valid() {
		log.Warn(ctx, "invalid folder owner type", log.Int("ownerType", int(ownerType)), log.String("partition", string(partition)), log.String("owner", owner), log.String("link", link))
		return false, ErrInvalidFolderOwnerType
	}
	condition := da.FolderCondition{
		ItemType:  int(entity.FolderItemTypeFile),
		OwnerType: int(ownerType),
		Owner:     owner,
		Partition: partition,
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

	//Lock folder name
	if req.Name != "" {
		locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixFolderName, req.Name)
		if err != nil {
			log.Error(ctx, "get locker failed",
				log.Err(err),
				log.String("prefix", da.RedisKeyPrefixFolderName),
				log.String("name", req.Name))
			return "", err
		}
		locker.Lock()
		defer locker.Unlock()
	}
	//get parent folder if exists
	if req.ParentID != "" && req.ParentID != constant.FolderRootPath {
		parentFolder, err = f.getFolder(ctx, tx, req.ParentID)
		if err != nil {
			return "", err
		}
	}
	//check create request
	err = f.checkCreateRequestEntity(ctx, req, parentFolder, operator)
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
	//更新parent folder文件数量
	//Update parent folder items count
	//parentFolder.ItemsCount = parentFolder.ItemsCount + 1
	//err = da.GetFolderDA().UpdateFolder(ctx, tx, parentFolder.ID, parentFolder)
	if req.ParentID != "" && req.ParentID != constant.FolderRootPath {
		err = f.batchRepairFolderItemsCountByIDs(ctx, tx, []string{req.ParentID})
		if err != nil {
			log.Error(ctx, "update parent folder items count failed", log.Err(err), log.Any("req", req))
			return "", err
		}
	}
	return folder.ID, nil
}

func (f *FolderModel) batchRemoveItemInternal(ctx context.Context, tx *dbo.DBContext, folderItems []*entity.FolderItem) error {
	fids := make([]string, len(folderItems))
	for i := range folderItems {
		//检查文件夹是否为空
		//check folder empty
		err := f.checkFolderEmpty(ctx, folderItems[i])
		if err != nil {
			log.Error(ctx, "checkFolderEmpty failed", log.Err(err), log.Any("folder", folderItems[i]))
			return err
		}
		fids[i] = folderItems[i].ID
	}

	//若为空Folder，删除
	//if the folder is empty, delete it
	err := da.GetFolderDA().BatchDeleteFolders(ctx, tx, fids)
	if err != nil {
		log.Error(ctx, "delete folder item failed", log.Err(err), log.Strings("fids", fids))
		return err
	}
	return nil
}

func (f *FolderModel) removeItemInternal(ctx context.Context, tx *dbo.DBContext, folderItem *entity.FolderItem) error {

	//检查文件夹是否为空
	//check folder empty
	err := f.checkFolderEmpty(ctx, folderItem)
	if err != nil {
		return err
	}

	//若为空Folder，删除
	//if the folder is empty, delete it
	err = da.GetFolderDA().DeleteFolder(ctx, tx, folderItem.ID)
	if err != nil {
		log.Error(ctx, "delete folder item failed", log.Err(err), log.Any("fid", folderItem.ID))
		return err
	}

	return nil
}

func (f *FolderModel) prepareAddItemParams(ctx context.Context, req entity.CreateFolderItemRequest, path entity.Path, ownerType entity.OwnerType, owner string, item *FolderItem, operator *entity.Operator) *entity.FolderItem {
	//path := parentFolder.ChildrenPath()
	//ownerType := parentFolder.OwnerType
	//owner := parentFolder.Owner

	now := time.Now().Unix()
	id := utils.NewID()
	return &entity.FolderItem{
		ID:        id,
		Link:      req.Link,
		ItemType:  entity.FolderItemTypeFile,
		OwnerType: ownerType,
		Owner:     owner,
		Editor:    operator.UserID,
		ParentID:  req.ParentFolderID,
		Name:      item.Name,
		DirPath:   path,
		Partition: req.Partition,
		//VisibilitySetting: item.VisibilitySetting,
		Thumbnail: item.Thumbnail,
		Creator:   operator.UserID,
		CreateAt:  now,
		UpdateAt:  now,
		DeleteAt:  0,
	}
}

func (f *FolderModel) checkAddItemRequest(ctx context.Context, req entity.CreateFolderItemRequest) error {
	//if req.ParentFolderID == "" {
	//	log.Warn(ctx, "invalid folder id", log.Any("req", req))
	//	return ErrEmptyFolderID
	//}
	if !req.Partition.Valid() {
		log.Warn(ctx, "invalid partition", log.Any("req", req))
		return ErrInvalidPartition
	}
	if !req.OwnerType.Valid() {
		log.Warn(ctx, "invalid owner type", log.Any("req", req))
		return ErrInvalidFolderOwnerType
	}
	if req.Link == "" {
		log.Warn(ctx, "invalid item id", log.Any("req", req))
		return ErrEmptyLinkID
	}

	return nil
}

func (f *FolderModel) checkAddItemParentRequest(ctx context.Context, req entity.CreateFolderItemRequest, parentFolder *entity.FolderItem, item *FolderItem) error {
	//check if parentFolder is a folder
	if !parentFolder.ItemType.IsFolder() {
		log.Warn(ctx, "move to an item not folder", log.Any("parentFolder", parentFolder))
		return ErrMoveToNotFolder
	}

	//组织下,不能有重复file,检查是否重复
	//One organization can't exists two file with the same item link, check duplication
	if parentFolder.OwnerType == entity.OwnerTypeOrganization {
		hasItem, err := f.hasFolderFileItem(ctx, dbo.MustGetDB(ctx), parentFolder.OwnerType, req.Partition, parentFolder.Owner, req.Link)
		if err != nil {
			return err
		}
		if hasItem {
			log.Warn(ctx, "duplicate item in org folder", log.Err(err), log.Any("req", req), log.Any("parentFolder", parentFolder))
			return ErrDuplicateItem
		}
	} else {
		//个人下查重名
		//check items duplicate
		//user private folder, can't exists two file with the same name
		items, err := f.getItemsFromFolders(ctx, req.ParentFolderID)
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
	}

	return nil
}

func (f *FolderModel) prepareCreateFolderParams(ctx context.Context, req entity.CreateFolderRequest, parentFolder *entity.FolderItem, operator *entity.Operator) *entity.FolderItem {
	path := entity.NewPath(constant.FolderRootPath)
	owner := req.OwnerType.Owner(operator)
	if parentFolder != nil {
		path = parentFolder.ChildrenPath()
		req.OwnerType = parentFolder.OwnerType
		owner = parentFolder.Owner
	}
	if req.ParentID == "" {
		req.ParentID = constant.FolderRootPath
	}
	now := time.Now().Unix()
	id := utils.NewID()
	return &entity.FolderItem{
		ID:          id,
		ItemType:    entity.FolderItemTypeFolder,
		OwnerType:   req.OwnerType,
		Owner:       owner,
		ParentID:    req.ParentID,
		Editor:      operator.UserID,
		Name:        req.Name,
		Description: req.Description,
		Keywords:    strings.Join(req.Keywords, constant.StringArraySeparator),
		DirPath:     path,
		Thumbnail:   req.Thumbnail,
		Partition:   req.Partition,
		//VisibilitySetting: constant.NoVisibilitySetting,
		Creator:  operator.UserID,
		CreateAt: now,
		UpdateAt: now,
		DeleteAt: 0,
	}
}

func (f *FolderModel) checkCreateRequestEntity(ctx context.Context, req entity.CreateFolderRequest, parentFolder *entity.FolderItem, operator *entity.Operator) error {
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

	if !req.Partition.Valid() {
		log.Warn(ctx, "invalid partition", log.Any("req", req))
		return ErrInvalidPartition
	}

	//check duplicate name
	err := f.checkDuplicateFolderName(ctx, req.OwnerType, req.Partition, req.Name, parentFolder, operator)
	if err != nil {
		return err
	}

	return nil
}

func (f *FolderModel) checkFolderEmpty(ctx context.Context, folderItem *entity.FolderItem) error {
	//若不是folder，返回
	//if it is not folder ,return
	if !folderItem.ItemType.IsFolder() {
		log.Info(ctx, "item is not folder", log.Any("folderItem", folderItem), log.String("path", string(folderItem.DirPath)))
		return nil
	}
	if folderItem.ItemsCount > 0 {
		return ErrFolderIsNotEmpty
	}
	return nil
}

func (f *FolderModel) checkDuplicateFolderName(ctx context.Context, ownerType entity.OwnerType, partition entity.FolderPartition, name string, parentFolder *entity.FolderItem, operator *entity.Operator) error {
	//check get all sub folders from parent parentFolder
	//folder下folder名唯一
	//folder name duplicate checking
	condition := da.FolderCondition{
		IDs:       entity.NullStrings{Valid: false},
		ItemType:  int(entity.FolderItemTypeFolder),
		OwnerType: int(ownerType),
		Owner:     ownerType.Owner(operator),
		Partition: partition,
		Name:      name,
	}
	folders, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		log.Error(ctx, "count parentFolder for check duplicate parentFolder failed",
			log.Err(err), log.Any("condition", condition))
		return err
	}

	for i := range folders {
		//handle with Character cases
		if folders[i].Name == name {
			log.Warn(ctx, "check duplicate name failed",
				log.Err(err),
				log.String("name", name),
				log.Any("folders", folders))
			return ErrDuplicateFolderName
		}
	}

	return nil
}

func (f *FolderModel) checkDuplicateFolderNameForUpdate(ctx context.Context, name string, folder *entity.FolderItem, operator *entity.Operator) error {
	//check get all sub folders from parent folder
	//folder下folder名唯一
	//folder name duplicate checking
	condition := da.FolderCondition{
		IDs:       entity.NullStrings{Valid: false},
		ItemType:  int(entity.FolderItemTypeFolder),
		OwnerType: int(folder.OwnerType),
		Partition: folder.Partition,
		Owner:     folder.Owner,
		Name:      name,
	}
	folders, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		log.Error(ctx, "count folder for check duplicate folder failed",
			log.Err(err), log.Any("condition", condition))
		return err
	}
	//check duplicate folder name
	for i := range folders {
		//handle with Character cases
		if folders[i].Name == name {
			//if owner type is organization,folder can be the same
			//in different partition
			return ErrDuplicateFolderName
		}
	}

	return nil
}

func (f *FolderModel) updateMoveFolderItemCount(ctx context.Context, tx *dbo.DBContext, ids []string) error {
	//update folder items count
	err := f.batchRepairFolderItemsCountByIDs(ctx, tx, ids)
	if err != nil {
		log.Error(ctx, "batchRepairFolderItemsCountByIDs failed", log.Err(err),
			log.Strings("ids", ids))
		return err
	}

	return nil
}

func (f *FolderModel) getDescendantItemsIDs(ctx context.Context, folder *entity.FolderItem) ([]string, error) {
	//若为文件，则直接返回
	//check folder type, folder case return
	if !folder.ItemType.IsFolder() {
		return nil, nil
	}
	items, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		DirDescendant: string(folder.ChildrenPath()),
	})
	if err != nil {
		log.Error(ctx, "search folder failed", log.Err(err), log.String("path", string(folder.DirPath)))
		return nil, err
	}
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}

	return ids, nil
}

func (f *FolderModel) getDescendantItemsMapByFolders(ctx context.Context, folders []*entity.FolderItem) (*entity.FolderDescendantItemsAndPath, error) {
	pathList := make([]string, 0)
	for i := range folders {
		if folders[i].ItemType.IsFolder() {
			pathList = append(pathList, folders[i].ChildrenPath().String())
		}
	}
	items, err := da.GetFolderDA().SearchFolder(ctx, dbo.MustGetDB(ctx), da.FolderCondition{
		DirDescendantList: pathList,
	})
	if err != nil {
		log.Error(ctx, "search folder failed", log.Err(err), log.Strings("path", pathList))
		return nil, err
	}
	folderMap := make(map[string][]*entity.FolderItem)
	folderPath := make(map[string]entity.Path)
	for i := range folders {
		subFolders := make([]*entity.FolderItem, 0)
		for j := range items {
			if strings.HasPrefix(items[j].DirPath.String(), folders[i].ChildrenPath().String()) {
				subFolders = append(subFolders, items[j])
			}
		}
		folderMap[folders[i].ID] = subFolders
		folderPath[folders[i].ID] = folders[i].ChildrenPath()
	}

	return &entity.FolderDescendantItemsAndPath{
		DescendantFoldersMap:  folderMap,
		FolderChildrenPathMap: folderPath,
	}, nil
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
			log.Error(ctx, "get folder info failed", log.Err(err), log.Strings("ids", folderItems))
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
		log.Error(ctx, "get folder items failed", log.Err(err), log.String("id", parentID))
		return nil, err
	}
	return items, nil
}
func (f *FolderModel) getFolder(ctx context.Context, tx *dbo.DBContext, fid string) (*entity.FolderItem, error) {
	parentFolder, err := da.GetFolderDA().GetFolderByID(ctx, tx, fid)
	if err != nil || parentFolder == nil {
		log.Error(ctx, "no such folder id", log.Err(err), log.String("fid", fid))
		return nil, ErrInvalidParentFolderId
	}
	return parentFolder, nil
}
func (f *FolderModel) getFolderMaybeRoot(ctx context.Context, tx *dbo.DBContext, fid string, ownerType entity.OwnerType, partition entity.FolderPartition, operator *entity.Operator) (*entity.FolderItem, error) {
	if fid == constant.FolderRootPath {
		return f.rootFolder(ctx, ownerType, partition, operator), nil
	}
	parentFolder, err := da.GetFolderDA().GetFolderByID(ctx, tx, fid)
	if err != nil || parentFolder == nil {
		log.Error(ctx, "no such folder id", log.Err(err), log.String("fid", fid))
		return nil, ErrInvalidParentFolderId
	}
	return parentFolder, nil
}

func (f *FolderModel) rootFolder(ctx context.Context, ownerType entity.OwnerType, partition entity.FolderPartition, operator *entity.Operator) *entity.FolderItem {
	return &entity.FolderItem{
		ID:        constant.FolderRootPath,
		OwnerType: ownerType,
		Owner:     ownerType.Owner(operator),
		ItemType:  entity.FolderItemTypeFolder,
		DirPath:   constant.FolderRootPath,
		Partition: partition,
	}
}

type FolderItem struct {
	Name              string
	Thumbnail         string
	VisibilitySetting []string
}

func (f *FolderModel) handleMoveSharedContentFolderRecursion(ctx context.Context, tx *dbo.DBContext, contentLinks []string, fromRootFolder, distFolder *entity.FolderItem, operator *entity.Operator) error {
	//TODO:shared folder
	//1. search items from folders
	//TODO:check root path
	log.Info(ctx, "handle move shared content",
		log.Strings("links", contentLinks),
		log.Any("from", fromRootFolder),
		log.Any("dist", distFolder))

	//2. search folders by parent ids
	records, err := da.GetSharedFolderDA().Search(ctx, tx, da.SharedFolderCondition{
		FolderIDs: []string{fromRootFolder.ID, distFolder.ID},
	})
	if err != nil {
		log.Error(ctx, "Can't search folder records",
			log.Err(err),
			log.String("FromFolderID", fromRootFolder.ID))
		return err
	}
	fromRecords := make([]*entity.SharedFolderRecord, 0)
	toRecords := make([]*entity.SharedFolderRecord, 0)
	for i := range records {
		if records[i].FolderID == fromRootFolder.ID {
			fromRecords = append(fromRecords, records[i])
		} else if records[i].FolderID == distFolder.ID {
			toRecords = append(toRecords, records[i])
		}
	}

	//delete auth content from from_folder
	if len(fromRecords) > 0 && fromRootFolder.ID != constant.FolderRootPath {
		//folder has shared to orgs
		//need to remove
		orgIDs := make([]string, len(fromRecords))
		for i := range fromRecords {
			orgIDs[i] = fromRecords[i].OrgID
		}
		oids := utils.SliceDeduplication(orgIDs)
		//delete content
		err = GetAuthedContentRecordsModel().BatchDelete(ctx, tx, entity.BatchDeleteAuthedContentByOrgsRequest{
			OrgIDs:     oids,
			FolderIDs:  []string{fromRootFolder.ID},
			ContentIDs: contentLinks,
		}, operator)
		if err != nil {
			log.Error(ctx, "Can't delete auth content records",
				log.Err(err),
				log.Strings("contentLinks", contentLinks),
				log.Strings("orgIDs", oids),
			)
			return err
		}
	}
	//add auth content to to_folder
	if len(toRecords) > 0 && distFolder.ID != constant.FolderRootPath {
		orgIDs := make([]string, len(toRecords))
		for i := range toRecords {
			orgIDs[i] = toRecords[i].OrgID
		}
		data := make([]*entity.AddAuthedContentRequest, 0)
		oids := utils.SliceDeduplication(orgIDs)
		for i := range oids {
			for j := range contentLinks {
				data = append(data, &entity.AddAuthedContentRequest{
					OrgID:        oids[i],
					FromFolderID: distFolder.ID,
					ContentID:    contentLinks[j],
				})
			}
		}

		err = GetAuthedContentRecordsModel().BatchAddByOrgIDs(ctx, tx, data, operator)
		if err != nil {
			log.Error(ctx, "batch add auth content records",
				log.Err(err),
				log.Any("data", data),
				log.Strings("orgIDs", oids),
			)
			return err
		}
	}

	return nil
}

func (f *FolderModel) handleMoveSharedContent(ctx context.Context,
	tx *dbo.DBContext,
	cidList []string,
	fromFolders []*entity.FolderItem,
	distFolder *entity.FolderItem,
	operator *entity.Operator) error {
	//1.Check content type, filter assets
	log.Info(ctx, "handle move shared content",
		log.Strings("cids", cidList),
		log.Any("from", fromFolders),
		log.Any("dist", distFolder))
	contents, err := GetContentModel().GetRawContentByIDList(ctx, tx, cidList)
	if err != nil {
		return err
	}
	//if content is asset, no need to share
	sharedContentIDs := make([]string, 0)
	for i := range contents {
		if !contents[i].ContentType.IsAsset() {
			sharedContentIDs = append(sharedContentIDs, contents[i].ID)
		}
	}

	log.Info(ctx, "pending shared content ids",
		log.Strings("cids", sharedContentIDs))
	if len(sharedContentIDs) < 1 {
		log.Info(ctx, "No need to share",
			log.Strings("cids", cidList))
		return nil
	}

	//2.Get from path & to path
	folderIDs := make([]string, 0)
	fromFolderIDs := make([]string, len(fromFolders))
	for i := range fromFolders {
		if fromFolders[i].ID != constant.FolderRootPath {
			folderIDs = append(folderIDs, fromFolders[i].ID)
		}
		fromFolderIDs[i] = fromFolders[i].ID
	}

	if distFolder.ID != constant.FolderRootPath {
		folderIDs = append(folderIDs, distFolder.ID)
	}

	log.Info(ctx, "pending update folder IDs",
		log.Strings("folderIDs", folderIDs))

	if len(folderIDs) < 1 {
		//in this case from root to root
		//no need to share
		return nil
	}

	condition := da.SharedFolderCondition{
		FolderIDs: folderIDs,
	}
	records, err := da.GetSharedFolderDA().Search(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "search shared folder failed",
			log.Err(err),
			log.Any("condition", condition))
		return err
	}

	log.Info(ctx, "shared folders result",
		log.Any("records", records))

	//3.Get orgs from folder
	fromOrgs := make([]string, 0)
	toOrgs := make([]string, 0)
	for i := range records {
		if records[i].FolderID == distFolder.ID {
			toOrgs = append(toOrgs, records[i].OrgID)
			continue
		}
		for j := range fromFolders {
			if records[i].FolderID == fromFolders[j].ID {
				fromOrgs = append(fromOrgs, records[i].OrgID)
			}
		}
	}
	log.Info(ctx, "from orgs and to orgs",
		log.Strings("fromOrgs", fromOrgs),
		log.Strings("toOrgs", toOrgs))
	//4.Update content auth records (remove/add)
	//remove from content id list
	//when from folder is not root
	if len(fromOrgs) != 0 {
		err = GetAuthedContentRecordsModel().BatchDelete(ctx, tx, entity.BatchDeleteAuthedContentByOrgsRequest{
			OrgIDs:     fromOrgs,
			FolderIDs:  fromFolderIDs,
			ContentIDs: sharedContentIDs,
		}, operator)
		if err != nil {
			log.Error(ctx, "Batch delete auth content failed",
				log.Err(err),
				log.Strings("fromOrgs", fromOrgs),
				log.Strings("fromContentIDs", sharedContentIDs))
			return err
		}
	}

	//when to folder is not root
	if len(toOrgs) != 0 {
		//add to content id list
		toContentIDSize := len(sharedContentIDs)
		toOrgsSize := len(toOrgs)
		data := make([]*entity.AddAuthedContentRequest, toOrgsSize*toContentIDSize)
		for i := range toOrgs {
			for j := range sharedContentIDs {
				data[i*toContentIDSize+j] = &entity.AddAuthedContentRequest{
					OrgID:        toOrgs[i],
					FromFolderID: distFolder.ID,
					ContentID:    sharedContentIDs[j],
				}
			}
		}
		err = GetAuthedContentRecordsModel().BatchAddByOrgIDs(ctx, tx, data, operator)
		if err != nil {
			log.Error(ctx, "Batch add auth content failed",
				log.Err(err),
				log.Any("data", data))
			return err
		}
	}

	return nil
}

func (f *FolderModel) updateLinkedItemPath(ctx context.Context,
	tx *dbo.DBContext,
	ids []string,
	path entity.Path,
	fromFolder []*entity.FolderItem,
	toFolder *entity.FolderItem,
	operator *entity.Operator) error {
	err := GetContentModel().BatchUpdateContentPath(ctx, tx, ids, path)
	if err != nil {
		log.Warn(ctx, "can't update content path by id", log.Err(err),
			log.Strings("ids", ids), log.String("path", path.String()))
		return err
	}
	err = f.handleMoveSharedContent(ctx, tx, ids, fromFolder, toFolder, operator)
	if err != nil {
		log.Error(ctx, "can't handle move shared content", log.Err(err),
			log.Any("fromFolder", fromFolder),
			log.Any("toFolder", toFolder),
			log.Strings("ids", ids),
			log.String("path", path.String()))
		return err
	}
	return nil
}

func (f *FolderModel) listContentInFolder(ctx context.Context, tx *dbo.DBContext, path string) ([]string, error) {
	//search all content in the folders
	condition := &da.ContentConditionInternal{DirPathRecursion: path}
	_, contentList, err := da.GetContentDA().SearchContentInternal(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "SearchContentInternal failed", log.Err(err),
			log.String("itemType", "content"),
			log.String("path", path))
		return nil, err
	}
	contentLinkIds := make([]string, len(contentList))
	for i := range contentList {
		contentLinkIds[i] = contentList[i].ID
	}
	return contentLinkIds, nil
}

func (f *FolderModel) listContentMapInFolders(ctx context.Context, tx *dbo.DBContext, pathMap map[string]entity.Path) (map[string][]string, error) {
	pathList := make([]string, len(pathMap))

	i := 0
	for _, v := range pathMap {
		pathList[i] = v.String()
		i++
	}
	//search all content in the folders
	condition := &da.ContentConditionInternal{DirPathRecursionList: pathList}
	_, contentList, err := da.GetContentDA().SearchContentInternal(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "SearchContentInternal failed", log.Err(err),
			log.String("itemType", "content"),
			log.Any("path", pathList))
		return nil, err
	}
	contentLinkMap := make(map[string][]string)
	for folderID, path := range pathMap {
		linkIDs := make([]string, 0)
		for j := range contentList {
			if strings.HasPrefix(contentList[j].DirPath.String(), path.String()) {
				linkIDs = append(linkIDs, contentList[j].ID)
			}
		}
		contentLinkMap[folderID] = linkIDs
	}

	return contentLinkMap, nil
}

func (f *FolderModel) replaceLinkedContentPath(ctx context.Context, tx *dbo.DBContext,
	req *entity.ReplaceLinkedContentPathRequest) error {
	linkPath := constant.FolderPathSeparator + req.FromFolder.ID
	if req.DistFolder.ID != constant.FolderRootPath {
		linkPath = string(req.DistFolder.ChildrenPath()) + constant.FolderPathSeparator + req.FromFolder.ID
	}

	err := GetContentModel().BatchReplaceContentPath(ctx, tx,
		req.ContentIDs,
		req.FromFolder.ChildrenPath().String(),
		linkPath)
	if err != nil {
		log.Error(ctx, "can't update content path by id", log.Err(err),
			log.Any("req", req), log.String("path", linkPath))
		return err
	}

	return nil
}

func createFolderItemByID(ctx context.Context, tx *dbo.DBContext, link string, user *entity.Operator) (*FolderItem, error) {
	fileType, id, err := parseLink(ctx, link)
	if err != nil {
		return nil, err
	}
	switch fileType {
	case entity.FolderFileTypeContent:
		content, err := GetContentModel().GetContentByID(ctx, tx, id, user)
		if err != nil {
			log.Warn(ctx, "can't find content by id", log.Err(err),
				log.String("itemType", string(fileType)), log.String("id", id))
			return nil, err
		}
		return &FolderItem{
			Name:              content.Name,
			Thumbnail:         content.Thumbnail,
			VisibilitySetting: content.PublishScope,
		}, nil
	}
	log.Warn(ctx, "unsupported file type",
		log.String("itemType", string(fileType)), log.String("id", id))
	return nil, ErrInvalidFolderItemType
}

func parseLink(ctx context.Context, link string) (entity.FolderFileType, string, error) {
	linkPairs := strings.Split(link, "-")
	if len(linkPairs) != 2 {
		log.Warn(ctx, "link is invalid", log.Err(ErrInvalidItemLink),
			log.String("link", link))
		return "", "", ErrInvalidItemLink
	}
	fileType := entity.NewFolderFileType(linkPairs[0])
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
