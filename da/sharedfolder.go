package da

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type ISharedFolderDA interface {
	AddSharedFolderRecord(ctx context.Context, tx *dbo.DBContext, req entity.SharedFolderRecord) error
	BatchAddSharedFolderRecord(ctx context.Context, tx *dbo.DBContext, req []*entity.SharedFolderRecord) error
	BatchDeleteSharedFolderRecord(ctx context.Context, tx *dbo.DBContext, orgID string, folderIDs []string) error
	BatchDeleteSharedFolderRecordByOrgIDs(ctx context.Context, tx *dbo.DBContext, folderID string, orgIDs []string) error

	SearchSharedFolderRecords(ctx context.Context, tx *dbo.DBContext, condition SharedFolderCondition) ([]*entity.SharedFolderRecord, error)
}

type SharedFolderDA struct {
	s dbo.BaseDA
}

func (sf *SharedFolderDA) AddSharedFolderRecord(ctx context.Context, tx *dbo.DBContext, req entity.SharedFolderRecord) error {
	req.ID = utils.NewID()
	req.CreateAt = time.Now().Unix()
	fmt.Println(req)
	_, err := sf.s.InsertTx(ctx, tx, &req)
	if err != nil {
		return err
	}
	return nil
}
func (sf *SharedFolderDA) BatchAddSharedFolderRecord(ctx context.Context, tx *dbo.DBContext, req []*entity.SharedFolderRecord) error {
	if len(req) < 1 {
		return nil
	}
	createAt := time.Now().Unix()
	columns := []string{
		"id", "org_id", "folder_id", "creator", "create_at", "update_at",
	}
	var values [][]interface{}
	for _, item := range req {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		values = append(values, []interface{}{item.ID, item.OrgID, item.FolderID, item.Creator, createAt, createAt})
	}
	template := SQLBatchInsert(entity.SharedFolderRecord{}.TableName(), columns, values)
	if err := tx.Exec(template.Format, template.Values...).Error; err != nil {
		log.Error(ctx, "batch insert cms_authed_contents: batch insert failed",
			log.Err(err),
			log.Any("items", values),
		)
		return err
	}
	return nil
}
func (sf *SharedFolderDA) BatchDeleteSharedFolderRecord(ctx context.Context, tx *dbo.DBContext, orgID string, folderIDs []string) error {
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

func (sf *SharedFolderDA) BatchDeleteSharedFolderRecordByOrgIDs(ctx context.Context, tx *dbo.DBContext, folderID string, orgIDs []string) error {
	now := time.Now().Unix()
	err := tx.Model(&entity.SharedFolderRecord{}).Where("org_id in (?) and folder_id = ?", orgIDs, folderID).Updates(entity.SharedFolderRecord{DeleteAt: now}).Error
	if err != nil {
		log.Error(ctx, "batch delete cms_authed_contents: batch delete failed",
			log.Err(err),
			log.Strings("orgIDs", orgIDs),
			log.String("folderID", folderID),
		)
		return err
	}
	return nil
}

func (sf *SharedFolderDA) SearchSharedFolderRecords(ctx context.Context, tx *dbo.DBContext, condition SharedFolderCondition) ([]*entity.SharedFolderRecord, error) {
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
