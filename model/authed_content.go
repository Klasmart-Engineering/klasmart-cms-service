package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/contentdata"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

//TODO:For authed content => implement the interface => done
type IAuthedContent interface {
	Add(ctx context.Context, tx *dbo.DBContext, req entity.AddAuthedContentRequest, op *entity.Operator) error
	BatchAdd(ctx context.Context, tx *dbo.DBContext, req entity.BatchAddAuthedContentRequest, op *entity.Operator) error
	BatchAddByOrgIDs(ctx context.Context, tx *dbo.DBContext, reqs []*entity.AddAuthedContentRequest, op *entity.Operator) error

	//Delete(ctx context.Context, tx *dbo.DBContext, req entity.DeleteAuthedContentRequest, op *entity.Operator) error
	BatchDelete(ctx context.Context, tx *dbo.DBContext, req entity.BatchDeleteAuthedContentByOrgsRequest, op *entity.Operator) error

	SearchRecordsList(ctx context.Context, tx *dbo.DBContext, condition entity.SearchAuthedContentRequest, op *entity.Operator) (int, []*entity.AuthedContentRecord, error)
	SearchDetailsList(ctx context.Context, tx *dbo.DBContext, condition entity.SearchAuthedContentRequest, op *entity.Operator) (int, []*entity.AuthedContentRecordInfo, error)
	QueryRecordsList(ctx context.Context, tx *dbo.DBContext, condition entity.SearchAuthedContentRequest, op *entity.Operator) ([]*entity.AuthedContentRecord, error)

	BatchUpdateVersion(ctx context.Context, tx *dbo.DBContext, oldID []string, newID string) error

	GetContentAuthByIDList(ctx context.Context, cids []string, operator *entity.Operator) (map[string]entity.ContentAuth, error)
}

type AuthedContent struct {
}

//AddAuthedContent add authed content record to org
func (ac *AuthedContent) Add(ctx context.Context, tx *dbo.DBContext, req entity.AddAuthedContentRequest, op *entity.Operator) error {
	//check duplicate
	contentIDs, err := ac.expandRelatedContentID(ctx, tx, []string{req.ContentID})
	if err != nil{
		return err
	}
	if len(contentIDs) > 0 {
		return ac.BatchAdd(ctx, tx, entity.BatchAddAuthedContentRequest{
			OrgID:      req.OrgID,
			FolderID:   req.FromFolderID,
			ContentIDs: contentIDs,
		}, op)
	}

	//lock org id for batch add
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixContentAuth, req.OrgID)
	if err != nil {
		return err
	}
	locker.Lock()
	defer locker.Unlock()


	condition := da.AuthedContentCondition{
		ContentIDs: []string{req.ContentID},
		FromFolderIDs: []string{req.FromFolderID},
		OrgIDs:     []string{req.OrgID},
	}
	total, err := da.GetAuthedContentRecordsDA().CountAuthedContentRecords(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "count authed content records failed",
			log.Err(err),
			log.Any("condition", condition))
		return err
	}
	if total > 0 {
		//if record is already in recorded
		//continue
		return nil
	}
	data := entity.AuthedContentRecord{
		OrgID:     req.OrgID,
		ContentID: req.ContentID,
		FromFolderID: req.FromFolderID,
		Creator:   op.UserID,
	}
	err = da.GetAuthedContentRecordsDA().AddAuthedContent(ctx, tx, data)
	if err != nil {
		log.Error(ctx, "add authed content records failed",
			log.Err(err),
			log.Any("data", data))
		return err
	}
	return nil
}

func (ac *AuthedContent) BatchAddByOrgIDs(ctx context.Context, tx *dbo.DBContext, reqs []*entity.AddAuthedContentRequest, op *entity.Operator) error {
	contentIDs := make([]string, len(reqs))
	folderIDs := make([]string, len(reqs))
	orgIDs := make([]string, len(reqs))

	for i := range reqs {
		contentIDs[i] = reqs[i].ContentID
		folderIDs[i] = reqs[i].FromFolderID
		orgIDs[i] = reqs[i].OrgID
	}

	allContentIDsMap, err := ac.expandRelatedContentIDMap(ctx, tx, contentIDs)
	if err != nil{
		return err
	}
	log.Info(ctx, "query already exists records results",
		log.Any("reqs", reqs),
		log.Any("allContentIDsMap", allContentIDsMap),
		log.Strings("contentIDs", contentIDs),
		log.Strings("folderIDs", folderIDs),
		log.Strings("orgIDs", orgIDs))

	for k, _ := range allContentIDsMap {
		allContentIDsMap[k] = utils.SliceDeduplication(allContentIDsMap[k])
	}
	folderIDs = utils.SliceDeduplication(folderIDs)
	orgIDs = utils.SliceDeduplication(orgIDs)

	log.Info(ctx, "duplicated records results",
		log.Any("allContentIDsMap", allContentIDsMap),
		log.Strings("contentIDs", contentIDs),
		log.Strings("folderIDs", folderIDs),
		log.Strings("orgIDs", orgIDs))
	//lock org id for batch add
	for i := range orgIDs {
		locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixContentAuth, orgIDs[i])
		if err != nil {
			return err
		}
		locker.Lock()
		defer locker.Unlock()
	}

	condition := da.AuthedContentCondition{
		ContentIDs: contentIDs,
		FromFolderIDs: folderIDs,
		OrgIDs:     orgIDs,
	}
	log.Info(ctx, "query authed content records",
		log.Any("condition", condition))
	authRecords, err := da.GetAuthedContentRecordsDA().QueryAuthedContentRecords(ctx, tx, condition)
	if err != nil{
		log.Error(ctx, "query authed content records failed",
			log.Err(err),
			log.Any("condition", condition))
		return err
	}
	log.Info(ctx, "query authed content records results",
		log.Any("authRecords", authRecords))
	filteredReq := make([]*entity.AddAuthedContentRequest, 0)
	for i := range reqs {
		for j := range allContentIDsMap[reqs[i].ContentID] {
			alreadyRecord := false
			contentID := allContentIDsMap[reqs[i].ContentID][j]
			for k := range authRecords {
				//if already auth the content
				//ignore
				if reqs[i].OrgID == authRecords[k].OrgID &&
					reqs[i].FromFolderID == authRecords[k].FromFolderID &&
					contentID == authRecords[k].ContentID {
					alreadyRecord = true
					break
				}
			}
			//only if alreadyRecord == false, (means auth is not in record)
			if !alreadyRecord {
				filteredReq = append(filteredReq, reqs[i])
			}
		}
	}

	log.Info(ctx, "filtered records results",
		log.Any("filteredReq", filteredReq))

	data := make([]*entity.AuthedContentRecord, len(filteredReq))
	for i := range filteredReq {
		data[i] = &entity.AuthedContentRecord{
			OrgID:     filteredReq[i].OrgID,
			ContentID: filteredReq[i].ContentID,
			FromFolderID: filteredReq[i].FromFolderID,
			Creator:   op.UserID,
		}
	}
	log.Info(ctx, "pending add records",
		log.Any("data", data))
	err = da.GetAuthedContentRecordsDA().BatchAddAuthedContent(ctx, tx, data)
	if err != nil {
		log.Error(ctx, "add authed content records failed",
			log.Err(err),
			log.Any("data", data))
		return err
	}
	return nil
}

//BatchAddAuthedContent add a list of authed content records to org
//if some of the list records (content_id and org_id) is exists, ignore
func (ac *AuthedContent) BatchAdd(ctx context.Context, tx *dbo.DBContext, req entity.BatchAddAuthedContentRequest, op *entity.Operator) error {
	contentIDs, err := ac.expandRelatedContentID(ctx, tx, req.ContentIDs)
	if err != nil{
		return err
	}
	condition := da.AuthedContentCondition{
		ContentIDs: contentIDs,
		OrgIDs:     []string{req.OrgID},
		FromFolderIDs: []string{req.FolderID},
	}
	//lock org id for batch add
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixContentAuth, req.OrgID)
	if err != nil {
		return err
	}
	locker.Lock()
	defer locker.Unlock()


	log.Info(ctx, "query already exists records",
		log.Any("condition", condition))
	objs, err := da.GetAuthedContentRecordsDA().QueryAuthedContentRecords(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "count authed content records failed",
			log.Err(err),
			log.Any("condition", condition))
		return err
	}

	log.Info(ctx, "query already exists records results",
		log.Any("objs", objs),
		log.Strings("contentIDs", contentIDs))

	//remove duplicate
	pendingAddIDs := make([]string, 0)
	if len(objs) > 0 {
		for i := range contentIDs {
			flag := false
			for j := range objs {
				if contentIDs[i] == objs[j].ContentID {
					flag = true
					break
				}
			}

			if !flag {
				pendingAddIDs = append(pendingAddIDs, contentIDs[i])
			}
		}
	} else {
		pendingAddIDs = contentIDs
	}
	log.Info(ctx, "pendingAddIDs",
		log.Any("objs", objs),
		log.Strings("pendingAddIDs", pendingAddIDs),
		log.Strings("contentIDs", contentIDs))

	data := make([]*entity.AuthedContentRecord, len(pendingAddIDs))
	for i := range pendingAddIDs {
		data[i] = &entity.AuthedContentRecord{
			OrgID:     req.OrgID,
			ContentID: pendingAddIDs[i],
			FromFolderID: req.FolderID,
			Creator:   op.UserID,
		}
	}
	log.Info(ctx, "add records",
		log.Any("data", data))
	err = da.GetAuthedContentRecordsDA().BatchAddAuthedContent(ctx, tx, data)
	if err != nil {
		log.Error(ctx, "add authed content records failed",
			log.Err(err),
			log.Any("data", data))
		return err
	}
	return nil
}

//DeleteAuthedContent delete record by org_id and content_id
//func (ac *AuthedContent) Delete(ctx context.Context, tx *dbo.DBContext, req entity.DeleteAuthedContentRequest, op *entity.Operator) error {
//	err := da.GetAuthedContentRecordsDA().BatchDeleteAuthedContent(ctx, tx, req.OrgID, []string{req.ContentID})
//	if err != nil {
//		log.Error(ctx, "remove authed content records failed",
//			log.Err(err),
//			log.Any("operator", op),
//			log.Any("req", req))
//		return err
//	}
//	return nil
//}

//BatchDeleteAuthedContent delete record by org_id and (content_id list)
//func (ac *AuthedContent) BatchDelete(ctx context.Context, tx *dbo.DBContext, req entity.BatchDeleteAuthedContentRequest, op *entity.Operator) error {
//	err := da.GetAuthedContentRecordsDA().BatchDeleteAuthedContent(ctx, tx, req.OrgID, req.ContentIDs)
//	if err != nil {
//		log.Error(ctx, "remove authed content records failed",
//			log.Err(err),
//			log.Any("operator", op),
//			log.Any("req", req))
//		return err
//	}
//	return nil
//}


//BatchDeleteAuthedContent delete record by org_id and (content_id list)
func (ac *AuthedContent) BatchDelete(ctx context.Context, tx *dbo.DBContext, req entity.BatchDeleteAuthedContentByOrgsRequest, op *entity.Operator) error {
	err := da.GetAuthedContentRecordsDA().BatchDeleteAuthedContentByOrgs(ctx, tx, req.OrgIDs, req.ContentIDs, req.FolderIDs)
	if err != nil {
		log.Error(ctx, "remove authed content records failed",
			log.Err(err),
			log.Any("operator", op),
			log.Any("req", req))
		return err
	}
	return nil
}

//SearchAuthedContentRecordsList search authed content records
func (ac *AuthedContent) SearchRecordsList(ctx context.Context, tx *dbo.DBContext, condition entity.SearchAuthedContentRequest, op *entity.Operator) (int, []*entity.AuthedContentRecord, error) {
	return da.GetAuthedContentRecordsDA().SearchAuthedContentRecords(ctx, tx, da.AuthedContentCondition{
		IDs:        condition.IDs,
		OrgIDs:     condition.OrgIDs,
		ContentIDs: condition.ContentIDs,
		Creator:    condition.Creator,

		Pager: condition.Pager,
	})
}

func (ac *AuthedContent) QueryRecordsList(ctx context.Context, tx *dbo.DBContext, condition entity.SearchAuthedContentRequest, op *entity.Operator) ([]*entity.AuthedContentRecord, error) {
	return da.GetAuthedContentRecordsDA().QueryAuthedContentRecords(ctx, tx, da.AuthedContentCondition{
		IDs:        condition.IDs,
		OrgIDs:     condition.OrgIDs,
		ContentIDs: condition.ContentIDs,
		Creator:    condition.Creator,

		Pager: condition.Pager,
	})
}

func (cm *AuthedContent) GetContentAuthByIDList(ctx context.Context, cids []string, operator *entity.Operator) (map[string]entity.ContentAuth, error) {
	contents, err := da.GetContentDA().GetContentByIDList(ctx, dbo.MustGetDB(ctx), cids)
	if err != nil {
		log.Error(ctx, "get content failed",
			log.Err(err),
			log.Strings("cids", cids))
		return nil, err
	}
	return cm.getContentAuth(ctx, contents, operator)
}
//SearchAuthedContentDetailsList search authed content records and content details
func (ac *AuthedContent) SearchDetailsList(ctx context.Context, tx *dbo.DBContext, condition entity.SearchAuthedContentRequest, op *entity.Operator) (int, []*entity.AuthedContentRecordInfo, error) {
	total, records, err := ac.SearchRecordsList(ctx, tx, condition, op)
	if err != nil {
		log.Error(ctx, "search authed content records failed",
			log.Err(err),
			log.Any("operator", op),
			log.Any("condition", condition))
		return 0, nil, err
	}
	contentIDs := make([]string, len(records))
	for i := range records {
		contentIDs[i] = records[i].ContentID
	}

	contentDetails, err := GetContentModel().GetContentByIDList(ctx, tx, contentIDs, op)
	if err != nil {
		log.Error(ctx, "get content list records failed",
			log.Err(err),
			log.Any("operator", op),
			log.Strings("contentIDs", contentIDs))
		return 0, nil, err
	}
	contentMap := make(map[string]*entity.ContentInfoWithDetails)
	for i := range contentDetails {
		contentMap[contentDetails[i].ID] = contentDetails[i]
	}

	ret := make([]*entity.AuthedContentRecordInfo, len(records))
	for i := range records {
		ret[i] = &entity.AuthedContentRecordInfo{
			AuthedContentRecord:    *records[i],
			ContentInfoWithDetails: *contentMap[records[i].ContentID],
		}
	}

	return total, ret, nil
}

func (ac *AuthedContent) expandRelatedContentID(ctx context.Context, tx *dbo.DBContext, cids []string) ([]string, error) {
	contents, err := GetContentModel().GetRawContentByIDList(ctx, tx, cids)
	if err != nil{
		return nil, err
	}
	ret := make([]string, 0)
	for i := range contents {
		if contents[i].ContentType == entity.ContentTypePlan {
			data, err := contentdata.CreateContentData(ctx, entity.ContentTypePlan, contents[i].Data)
			if err != nil{
				log.Error(ctx, "create content data failed",
					log.Err(err),
					log.Any("content", contents[i]))
				return nil, err
			}
			relatedIDs := data.SubContentIDs(ctx)
			ret = append(ret, relatedIDs...)
			log.Info(ctx, "expanding content IDs",
				log.Strings("ids", relatedIDs))
		}

		log.Info(ctx, "content IDs",
			log.String("id", contents[i].ID),
			log.Any("content", contents[i]))
		ret = append(ret, contents[i].ID)
	}
	return ret, nil
}


func (ac *AuthedContent) expandRelatedContentIDMap(ctx context.Context, tx *dbo.DBContext, cids []string) (map[string][]string, error) {
	contents, err := GetContentModel().GetRawContentByIDList(ctx, tx, cids)
	if err != nil{
		return nil, err
	}
	ret := make(map[string][]string)
	for i := range contents {
		ret[contents[i].ID] = []string{contents[i].ID}
		if contents[i].ContentType == entity.ContentTypePlan {
			data, err := contentdata.CreateContentData(ctx, entity.ContentTypePlan, contents[i].Data)
			if err != nil{
				log.Error(ctx, "create content data failed",
					log.Err(err),
					log.Any("content", contents[i]))
				return nil, err
			}
			relatedIDs := data.SubContentIDs(ctx)
			ret[contents[i].ID] = append(ret[contents[i].ID], relatedIDs...)
			log.Info(ctx, "expanding content IDs",
				log.Strings("ids", relatedIDs))
		}

		log.Info(ctx, "content IDs",
			log.String("id", contents[i].ID),
			log.Any("content", contents[i]))
	}
	return ret, nil
}

//BatchUpdateAuthedContentVersion replace content id for new version
func (ac *AuthedContent) BatchUpdateVersion(ctx context.Context, tx *dbo.DBContext, oldIDs []string, newID string) error {
	err := da.GetAuthedContentRecordsDA().ReplaceContentID(ctx, tx, oldIDs, newID)
	if err != nil {
		log.Error(ctx, "replace authed content records failed",
			log.Err(err),
			log.Strings("oldIDs", oldIDs),
			log.String("newID", newID))
		return err
	}
	return nil
}

func (cm *AuthedContent) getContentAuth(ctx context.Context, contents []*entity.Content, operator *entity.Operator) (map[string]entity.ContentAuth, error) {
	result := make(map[string]entity.ContentAuth)
	pendingAuthContentIDs := make([]string, 0)

	contentLatestIDMap := make(map[string]string)
	contentLatestIDRevert := make(map[string]string)
	for i := range contents {
		if contents[i].LatestID == "" {
			contentLatestIDMap[contents[i].ID] = contents[i].ID
			contentLatestIDRevert[contents[i].ID] = contents[i].ID
		} else {
			contentLatestIDMap[contents[i].ID] = contents[i].LatestID
			contentLatestIDRevert[contents[i].LatestID] = contents[i].ID
		}
	}

	for i := range contents {
		if contents[i].Org == operator.OrgID {
			result[contents[i].ID] = entity.ContentAuthed
		} else {
			result[contents[i].ID] = entity.ContentUnauthed
			//search auth contents use latest id
			pendingAuthContentIDs = append(pendingAuthContentIDs, contentLatestIDMap[contents[i].ID])
		}
	}
	if len(pendingAuthContentIDs) > 0 {
		authRecords, err := GetAuthedContentRecordsModel().QueryRecordsList(ctx, dbo.MustGetDB(ctx), entity.SearchAuthedContentRequest{
			OrgIDs:     []string{operator.OrgID, constant.ShareToAll},
			ContentIDs: pendingAuthContentIDs,
		}, operator)
		if err != nil {
			log.Error(ctx, "query auth contents records failed",
				log.Err(err),
				log.String("orgID", operator.OrgID),
				log.Strings("contentIDs", pendingAuthContentIDs))
			return nil, err
		}
		for i := range authRecords {
			//result revert to current id
			result[contentLatestIDRevert[authRecords[i].ContentID]] = entity.ContentAuthed
		}
	}

	return result, nil
}
var (
	authedContentModel IAuthedContent
	_authedContentOnce sync.Once
)

//GetAuthedContentRecordsModel get record model
func GetAuthedContentRecordsModel() IAuthedContent {
	_authedContentOnce.Do(func() {
		authedContentModel = new(AuthedContent)
	})

	return authedContentModel
}
