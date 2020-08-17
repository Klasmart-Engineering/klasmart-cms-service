package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/dbo"
)

var (
	_contentVersionOnce sync.Once
	contentVersionDA    IContentVersionDA
)

type IContentVersionDA interface {
	AddContentVersionRecord(ctx context.Context, tx *dbo.DBContext, cv entity.ContentVersion) (int64, error)
	RemoveContentVersionRecord(ctx context.Context, tx *dbo.DBContext, rid int64) error
	GetContentVersionById(ctx context.Context, tx *dbo.DBContext, rid int64) (*entity.ContentVersion, error)

	SearchContentVersion(ctx context.Context, tx *dbo.DBContext, condition ContentVersionCondition) (int, []*entity.ContentVersion, error)
}

type ContentVersionCondition struct {
	Ids        []int64
	ContentIds []string
	LastIds    []string
	MainIds    []string
	Versions   []int

	Page     int
	PageSize int
}

func (s *ContentVersionCondition) GetConditions() ([]string, []interface{}) {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)
	if len(s.Ids) > 0 {
		condition := " id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.Ids)
	}

	if len(s.ContentIds) > 0 {
		condition := " content_id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.ContentIds)
	}

	if len(s.LastIds) > 0 {
		condition := " last_id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.LastIds)
	}

	if len(s.MainIds) > 0 {
		condition := " main_id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.MainIds)
	}

	if len(s.Versions) > 0 {
		condition := " version in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.Versions)
	}

	conditions = append(conditions, " deleted_at IS NULL ")
	return conditions, params
}
func (s *ContentVersionCondition) GetPager() *dbo.Pager {
	return &dbo.Pager{
		Page:     s.Page,
		PageSize: s.PageSize,
	}
}
func (s *ContentVersionCondition) GetOrderBy() string {
	return "version"
}

type DBContentVersionDA struct {
	s dbo.BaseDA
}

func (cd *DBContentVersionDA) AddContentVersionRecord(ctx context.Context, tx *dbo.DBContext, cv entity.ContentVersion) (int64, error) {
	now := time.Now()
	cv.UpdatedAt = &now
	cv.CreatedAt = &now
	err := cd.s.SaveTx(ctx, tx, &cv)
	if err != nil {
		return -1, err
	}
	return cv.Id, nil
}
func (cd *DBContentVersionDA) RemoveContentVersionRecord(ctx context.Context, tx *dbo.DBContext, rid int64) error {
	now := time.Now()
	cv := new(entity.ContentVersion)
	cv.Id = rid
	cv.DeletedAt = &now
	_, err := cd.s.UpdateTx(ctx, tx, cv)
	if err != nil {
		return err
	}
	return nil
}
func (cd *DBContentVersionDA) GetContentVersionById(ctx context.Context, tx *dbo.DBContext, rid int64) (*entity.ContentVersion, error) {
	obj := new(entity.ContentVersion)
	err := cd.s.GetTx(ctx, tx, rid, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (cd *DBContentVersionDA) SearchContentVersion(ctx context.Context, tx *dbo.DBContext, condition ContentVersionCondition) (int, []*entity.ContentVersion, error) {
	objs := make([]*entity.ContentVersion, 0)
	count, err := cd.s.PageTx(ctx, tx, &condition, &objs)
	if err != nil {
		return 0, nil, err
	}

	return count, objs, nil
}

func GetContentVersionDA() IContentVersionDA {
	_contentVersionOnce.Do(func() {
		contentVersionDA = new(DBContentVersionDA)
	})

	return contentVersionDA
}
