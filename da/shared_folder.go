package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type ISharedFolderDA interface {
	Add(ctx context.Context, tx *dbo.DBContext, req entity.SharedFolderRecord) error
	BatchAdd(ctx context.Context, tx *dbo.DBContext, req []*entity.SharedFolderRecord) error
	BatchDelete(ctx context.Context, tx *dbo.DBContext, orgID string, folderIDs []string) error
	BatchDeleteByOrgIDs(ctx context.Context, tx *dbo.DBContext, folderID string, orgIDs []string) error
	InsertBatchesTx(ctx context.Context, tx *dbo.DBContext, req []*entity.SharedFolderRecord) error

	Search(ctx context.Context, tx *dbo.DBContext, condition SharedFolderCondition) ([]*entity.SharedFolderRecord, error)
}

type SharedFolderDA struct {
	s dbo.BaseDA
}

func (sf *SharedFolderDA) Add(ctx context.Context, tx *dbo.DBContext, req entity.SharedFolderRecord) error {
	if req.ID == "" {
		req.ID = utils.NewID()
	}
	req.CreateAt = time.Now().Unix()
	_, err := sf.s.InsertTx(ctx, tx, &req)
	if err != nil {
		return err
	}
	return nil
}

func (sf *SharedFolderDA) InsertBatchesTx(ctx context.Context, tx *dbo.DBContext, req []*entity.SharedFolderRecord) error {
	if len(req) < 1 {
		return nil
	}
	createAt := time.Now().Unix()
	var models []entity.SharedFolderRecord
	for _, item := range req {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		models = append(models, entity.SharedFolderRecord{ID: item.ID, OrgID: item.OrgID, FolderID: item.FolderID,
			Creator: item.Creator, CreateAt: createAt, UpdateAt: createAt})
	}
	_, err := sf.s.InsertInBatchesTx(ctx, tx, &models, constant.ShareAllBatchSize)
	if err != nil {
		log.Error(ctx, "batch insert cms_authed_contents: batch insert failed",
			log.Err(err),
			log.Any("items", models),
		)
		return err
	}
	return nil
}

func (sf *SharedFolderDA) BatchAdd(ctx context.Context, tx *dbo.DBContext, req []*entity.SharedFolderRecord) error {
	if len(req) < 1 {
		return nil
	}
	createAt := time.Now().Unix()
	var models []entity.SharedFolderRecord
	for _, item := range req {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		models = append(models, entity.SharedFolderRecord{ID: item.ID, OrgID: item.OrgID, FolderID: item.FolderID,
			Creator: item.Creator, CreateAt: createAt, UpdateAt: createAt})
	}
	_, err := sf.s.InsertTx(ctx, tx, &models)
	if err != nil {
		log.Error(ctx, "batch insert cms_authed_contents: batch insert failed",
			log.Err(err),
			log.Any("items", req),
		)
		return err
	}
	return nil
}
func (sf *SharedFolderDA) BatchDelete(ctx context.Context, tx *dbo.DBContext, orgID string, folderIDs []string) error {
	now := time.Now().Unix()
	err := tx.Model(&entity.SharedFolderRecord{}).Where("org_id = ? and folder_id in (?)", orgID, folderIDs).Updates(entity.SharedFolderRecord{DeleteAt: now}).Error
	if err != nil {
		log.Error(ctx, "batch delete cms_authed_contents: batch delete failed",
			log.Err(err),
			log.String("orgID", orgID),
			log.Strings("folderIDs", folderIDs),
		)
		return err
	}
	return nil
}

func (sf *SharedFolderDA) BatchDeleteByOrgIDs(ctx context.Context, tx *dbo.DBContext, folderID string, orgIDs []string) error {
	now := time.Now().Unix()
	step := constant.ShareAllBatchSize

	for start := 0; start < len(orgIDs); {
		end := start + step
		if end >= len(orgIDs) {
			end = len(orgIDs)
		}
		ids := orgIDs[start:end]
		err := tx.Model(&entity.SharedFolderRecord{}).Where("org_id in (?) and folder_id = ?", ids, folderID).Updates(entity.SharedFolderRecord{DeleteAt: now}).Error
		if err != nil {
			log.Error(ctx, "batch delete cms_authed_contents: batch delete failed",
				log.Err(err),
				log.Int("start", start),
				log.Int("end", end),
				log.Strings("orgIDs", ids),
				log.String("folderID", folderID),
			)
			return err
		}
		start = end
	}

	return nil
}

func (sf *SharedFolderDA) Search(ctx context.Context, tx *dbo.DBContext, condition SharedFolderCondition) ([]*entity.SharedFolderRecord, error) {
	objs := make([]*entity.SharedFolderRecord, 0)
	err := sf.s.QueryTx(ctx, tx, &condition, &objs)

	if err != nil {
		log.Error(ctx, "query cms_shared_folder_records failed",
			log.Err(err),
			log.Any("condition", condition),
		)
		return nil, err
	}
	return objs, nil
}

type SharedFolderCondition struct {
	IDs       []string
	OrgIDs    []string
	FolderIDs []string
	Creator   []string
}

func (s *SharedFolderCondition) GetConditions() ([]string, []interface{}) {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	if len(s.IDs) > 0 {
		condition := "id in (?)"
		conditions = append(conditions, condition)
		params = append(params, s.IDs)
	}
	if len(s.OrgIDs) > 0 {
		condition := "org_id in (?)"
		conditions = append(conditions, condition)
		params = append(params, s.OrgIDs)
	}
	if len(s.FolderIDs) > 0 {
		condition := "folder_id in (?)"
		conditions = append(conditions, condition)
		params = append(params, s.FolderIDs)
	}
	if len(s.Creator) > 0 {
		condition := "creator in (?)"
		conditions = append(conditions, condition)
		params = append(params, s.Creator)
	}

	conditions = append(conditions, " delete_at = 0")
	return conditions, params
}
func (s *SharedFolderCondition) GetPager() *dbo.Pager {
	return &dbo.NoPager
}
func (s *SharedFolderCondition) GetOrderBy() string {
	return ""
}

var (
	_sharedFolderDA     ISharedFolderDA
	_sharedFolderDAOnce sync.Once
)

func GetSharedFolderDA() ISharedFolderDA {
	_sharedFolderDAOnce.Do(func() {
		_sharedFolderDA = new(SharedFolderDA)
	})

	return _sharedFolderDA
}
