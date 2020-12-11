package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

//TODO:For authed content => implement the interface => done
type IAuthedContent interface {
	AddAuthedContent(ctx context.Context, tx *dbo.DBContext, req entity.AddAuthedContentRequest, op *entity.Operator) error
	BatchAddAuthedContent(ctx context.Context, tx *dbo.DBContext, req entity.BatchAddAuthedContentRequest, op *entity.Operator) error
	DeleteAuthedContent(ctx context.Context, tx *dbo.DBContext, req entity.DeleteAuthedContentRequest, op *entity.Operator) error
	BatchDeleteAuthedContent(ctx context.Context, tx *dbo.DBContext, req entity.BatchDeleteAuthedContentRequest, op *entity.Operator) error

	SearchAuthedContentRecordsList(ctx context.Context, tx *dbo.DBContext, condition entity.SearchAuthedContentRequest, op *entity.Operator) (int, []*entity.AuthedContentRecord, error)
	SearchAuthedContentDetailsList(ctx context.Context, tx *dbo.DBContext, condition entity.SearchAuthedContentRequest, op *entity.Operator) (int, []*entity.AuthedContentRecordInfo, error)

	BatchUpdateAuthedContentVersion(ctx context.Context, tx *dbo.DBContext, oldID []string, newID string) error
}

type AuthedContent struct {
}

//AddAuthedContent add authed content record to org
func (ac *AuthedContent) AddAuthedContent(ctx context.Context, tx *dbo.DBContext, req entity.AddAuthedContentRequest, op *entity.Operator) error {
	//check duplicate
	condition := da.AuthedContentCondition{
		ContentIDs: []string{req.ContentId},
		OrgIDs:     []string{req.OrgId},
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
		OrgID:     req.OrgId,
		ContentID: req.ContentId,
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

//BatchAddAuthedContent add a list of authed content records to org
//if some of the list records (content_id and org_id) is exists, ignore
func (ac *AuthedContent) BatchAddAuthedContent(ctx context.Context, tx *dbo.DBContext, req entity.BatchAddAuthedContentRequest, op *entity.Operator) error {
	condition := da.AuthedContentCondition{
		ContentIDs: req.ContentIds,
		OrgIDs:     []string{req.OrgId},
	}
	objs, err := da.GetAuthedContentRecordsDA().QueryAuthedContentRecords(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "count authed content records failed",
			log.Err(err),
			log.Any("condition", condition))
		return err
	}

	//remove duplicate
	pendingAddIDs := make([]string, 0)
	if len(objs) > 0 {
		for i := range req.ContentIds {
			flag := false
			for j := range objs {
				if req.ContentIds[i] == objs[j].ContentID {
					flag = true
					break
				}
			}

			if !flag {
				pendingAddIDs = append(pendingAddIDs, req.ContentIds[i])
			}
		}
	} else {
		pendingAddIDs = req.ContentIds
	}

	data := make([]entity.AuthedContentRecord, len(pendingAddIDs))
	for i := range pendingAddIDs {
		data[i] = entity.AuthedContentRecord{
			OrgID:     req.OrgId,
			ContentID: pendingAddIDs[i],
			Creator:   op.UserID,
		}
	}
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
func (ac *AuthedContent) DeleteAuthedContent(ctx context.Context, tx *dbo.DBContext, req entity.DeleteAuthedContentRequest, op *entity.Operator) error {
	err := da.GetAuthedContentRecordsDA().BatchDeleteAuthedContent(ctx, tx, req.OrgId, []string{req.ContentId})
	if err != nil {
		log.Error(ctx, "remove authed content records failed",
			log.Err(err),
			log.Any("operator", op),
			log.Any("req", req))
		return err
	}
	return nil
}

//BatchDeleteAuthedContent delete record by org_id and (content_id list)
func (ac *AuthedContent) BatchDeleteAuthedContent(ctx context.Context, tx *dbo.DBContext, req entity.BatchDeleteAuthedContentRequest, op *entity.Operator) error {
	err := da.GetAuthedContentRecordsDA().BatchDeleteAuthedContent(ctx, tx, req.OrgId, req.ContentIds)
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
func (ac *AuthedContent) SearchAuthedContentRecordsList(ctx context.Context, tx *dbo.DBContext, condition entity.SearchAuthedContentRequest, op *entity.Operator) (int, []*entity.AuthedContentRecord, error) {
	return da.GetAuthedContentRecordsDA().SearchAuthedContentRecords(ctx, tx, da.AuthedContentCondition{
		IDs:        condition.IDs,
		OrgIDs:     condition.OrgIds,
		ContentIDs: condition.ContentIds,
		Creator:    condition.Creator,

		Pager: condition.Pager,
	})
}

//SearchAuthedContentDetailsList search authed content records and content details
func (ac *AuthedContent) SearchAuthedContentDetailsList(ctx context.Context, tx *dbo.DBContext, condition entity.SearchAuthedContentRequest, op *entity.Operator) (int, []*entity.AuthedContentRecordInfo, error) {
	total, records, err := ac.SearchAuthedContentRecordsList(ctx, tx, condition, op)
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

	contentDetails, err := GetContentModel().GetContentByIdList(ctx, tx, contentIDs, op)
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

//BatchUpdateAuthedContentVersion replace content id for new version
func (ac *AuthedContent) BatchUpdateAuthedContentVersion(ctx context.Context, tx *dbo.DBContext, oldIDs []string, newID string) error {
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
