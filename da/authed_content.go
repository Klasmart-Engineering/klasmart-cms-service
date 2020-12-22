package da

import (
	"context"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IAuthedContentDA interface {
	AddAuthedContent(ctx context.Context, tx *dbo.DBContext, req entity.AuthedContentRecord) error
	BatchAddAuthedContent(ctx context.Context, tx *dbo.DBContext, req []*entity.AuthedContentRecord) error
	BatchDeleteAuthedContent(ctx context.Context, tx *dbo.DBContext, orgID string, contentIDs []string) error
	BatchDeleteAuthedContentByOrgs(ctx context.Context, tx *dbo.DBContext, orgID []string, contentIDs []string) error

	ReplaceContentID(ctx context.Context, tx *dbo.DBContext, oldContentIDs []string, newContentID string) error
	SearchAuthedContentRecords(ctx context.Context, tx *dbo.DBContext, condition AuthedContentCondition) (int, []*entity.AuthedContentRecord, error)
	CountAuthedContentRecords(ctx context.Context, tx *dbo.DBContext, condition AuthedContentCondition) (int, error)
	QueryAuthedContentRecords(ctx context.Context, tx *dbo.DBContext, condition AuthedContentCondition) ([]*entity.AuthedContentRecord, error)
}

type AuthedContentDA struct {
	s dbo.BaseDA
}

func (ac *AuthedContentDA) AddAuthedContent(ctx context.Context, tx *dbo.DBContext, req entity.AuthedContentRecord) error {
	req.ID = utils.NewID()
	req.CreateAt = time.Now().Unix()
	_, err := ac.s.InsertTx(ctx, tx, &req)
	if err != nil {
		return err
	}
	return nil
}
func (ac *AuthedContentDA) BatchAddAuthedContent(ctx context.Context, tx *dbo.DBContext, req []*entity.AuthedContentRecord) error {
	if len(req) < 1 {
		return nil
	}
	createAt := time.Now().Unix()
	columns := []string{
		"id", "org_id", "content_id", "creator", "create_at", "duration",
	}
	var values [][]interface{}
	for _, item := range req {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		values = append(values, []interface{}{item.ID, item.OrgID, item.ContentID, item.Creator, createAt, item.Duration})
	}
	template := SQLBatchInsert(entity.AuthedContentRecord{}.TableName(), columns, values)
	if err := tx.Exec(template.Format, template.Values...).Error; err != nil {
		log.Error(ctx, "batch insert cms_authed_contents: batch insert failed",
			log.Err(err),
			log.Any("items", values),
		)
		return err
	}
	return nil
}
func (ac *AuthedContentDA) BatchDeleteAuthedContent(ctx context.Context, tx *dbo.DBContext, orgID string, contentIDs []string) error {
	now := time.Now().Unix()
	err := tx.Model(&entity.AuthedContentRecord{}).Where("org_id = ? and content_id in (?)", orgID, contentIDs).Updates(entity.AuthedContentRecord{DeleteAt: now}).Error
	if err != nil {
		log.Error(ctx, "batch delete cms_authed_contents: batch delete failed",
			log.Err(err),
			log.String("orgID", orgID),
			log.Strings("contentIDs", contentIDs),
		)
		return err
	}
	return nil
}
func (ac *AuthedContentDA) BatchDeleteAuthedContentByOrgs(ctx context.Context, tx *dbo.DBContext, orgIDs []string, contentIDs []string) error {
	now := time.Now().Unix()
	err := tx.Model(&entity.AuthedContentRecord{}).Where("org_id in (?) and content_id in (?)", orgIDs, contentIDs).Updates(entity.AuthedContentRecord{DeleteAt: now}).Error
	if err != nil {
		log.Error(ctx, "batch delete cms_authed_contents: batch delete failed",
			log.Err(err),
			log.Strings("orgIDs", orgIDs),
			log.Strings("contentIDs", contentIDs),
		)
		return err
	}
	return nil
}
func (ac *AuthedContentDA) ReplaceContentID(ctx context.Context, tx *dbo.DBContext, oldContentIDs []string, newContentID string) error {
	err := tx.Model(&entity.AuthedContentRecord{}).Where(" content_id in (?)", oldContentIDs).Updates(entity.AuthedContentRecord{ContentID: newContentID}).Error
	if err != nil {
		log.Error(ctx, "batch replace cms_authed_contents: replace failed",
			log.Err(err),
			log.String("newContentID", newContentID),
			log.Strings("oldContentIDs", oldContentIDs),
		)
		return err
	}
	return nil
}

func (ac *AuthedContentDA) SearchAuthedContentRecords(ctx context.Context, tx *dbo.DBContext, condition AuthedContentCondition) (int, []*entity.AuthedContentRecord, error) {
	objs := make([]*entity.AuthedContentRecord, 0)
	count, err := ac.s.PageTx(ctx, tx, &condition, &objs)
	if err != nil {
		log.Error(ctx, "query cms_authed_contents failed",
			log.Err(err),
			log.Any("condition", condition),
		)
		return 0, nil, err
	}

	return count, objs, nil
}

func (ac *AuthedContentDA) CountAuthedContentRecords(ctx context.Context, tx *dbo.DBContext, condition AuthedContentCondition) (int, error) {
	total, err := ac.s.CountTx(ctx, tx, &condition, entity.AuthedContentRecord{})
	if err != nil {
		log.Error(ctx, "count cms_authed_contents failed",
			log.Err(err),
			log.Any("condition", condition),
		)
		return 0, err
	}
	return total, nil
}

func (ac *AuthedContentDA) QueryAuthedContentRecords(ctx context.Context, tx *dbo.DBContext, condition AuthedContentCondition) ([]*entity.AuthedContentRecord, error) {
	objs := make([]*entity.AuthedContentRecord, 0)
	err := ac.s.QueryTx(ctx, tx, &condition, &objs)
	if err != nil {
		log.Error(ctx, "query cms_authed_contents failed",
			log.Err(err),
			log.Any("condition", condition),
		)
		return nil, err
	}
	return objs, nil
}

type AuthedContentCondition struct {
	IDs           []string
	OrgIDs        []string
	ContentIDs    []string
	Creator       []string
	FromFolderIDs []string

	Pager *utils.Pager
}

func (s *AuthedContentCondition) GetConditions() ([]string, []interface{}) {
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
	if len(s.FromFolderIDs) > 0 {
		condition := "from_folder_id in (?)"
		conditions = append(conditions, condition)
		params = append(params, s.FromFolderIDs)
	}
	if len(s.ContentIDs) > 0 {
		condition := "content_id in (?)"
		conditions = append(conditions, condition)
		params = append(params, s.ContentIDs)
	}
	if len(s.Creator) > 0 {
		condition := "creator in (?)"
		conditions = append(conditions, condition)
		params = append(params, s.Creator)
	}

	conditions = append(conditions, " delete_at = 0")
	return conditions, params
}
func (s *AuthedContentCondition) GetPager() *dbo.Pager {
	if s.Pager == nil {
		return &dbo.NoPager
	}
	return &dbo.Pager{
		Page:     int(s.Pager.PageIndex),
		PageSize: int(s.Pager.PageSize),
	}
}
func (s *AuthedContentCondition) GetOrderBy() string {
	return ""
}

var (
	authedContentDA    IAuthedContentDA
	_authedContentOnce sync.Once
)

func GetAuthedContentRecordsDA() IAuthedContentDA {
	_authedContentOnce.Do(func() {
		authedContentDA = new(AuthedContentDA)
	})

	return authedContentDA
}
