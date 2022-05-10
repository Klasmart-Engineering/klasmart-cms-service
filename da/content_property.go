package da

import (
	"context"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IContentPropertyDA interface {
	BatchAdd(ctx context.Context, tx *dbo.DBContext, co []*entity.ContentProperty) error
	CleanByContentID(ctx context.Context, tx *dbo.DBContext, contentID string) error

	BatchGetByContentIDList(ctx context.Context, tx *dbo.DBContext, contentID []string) ([]*entity.ContentProperty, error)
	BatchGetByContentIDsMapResult(ctx context.Context, tx *dbo.DBContext, contentID []string) (map[string][]*entity.ContentProperty, error)
}
type ContentPropertyDA struct {
	s dbo.BaseDA
}

func (c *ContentPropertyDA) BatchAdd(ctx context.Context, tx *dbo.DBContext, co []*entity.ContentProperty) error {
	_, err := c.s.InsertTx(ctx, tx, &co)
	if err != nil {
		log.Error(ctx, "db exec sql error\"",
			log.Err(err),
			log.Any("items", co),
		)
		return err
	}
	return nil
}

func (c *ContentPropertyDA) CleanByContentID(ctx context.Context, tx *dbo.DBContext, contentID string) error {
	err := tx.Where("content_id = ?", contentID).Delete(entity.ContentProperty{}).Error
	if err != nil {
		log.Error(ctx, "db exec sql error", log.String("contentID", contentID), log.Err(err))
		return err
	}
	return nil
}

func (c *ContentPropertyDA) BatchGetByContentIDsMapResult(ctx context.Context, tx *dbo.DBContext, contentIDs []string) (map[string][]*entity.ContentProperty, error) {
	list, err := c.BatchGetByContentIDList(ctx, tx, contentIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]*entity.ContentProperty)
	for _, it := range list {
		result[it.ContentID] = append(result[it.ContentID], it)
	}
	return result, nil
}

func (c *ContentPropertyDA) BatchGetByContentIDList(ctx context.Context, tx *dbo.DBContext, contentID []string) ([]*entity.ContentProperty, error) {
	if len(contentID) < 1 {
		return nil, nil
	}
	condition := ContentPropertyCondition{ContentIDs: dbo.NullStrings{
		Strings: contentID,
		Valid:   true,
	}}
	objs := make([]*entity.ContentProperty, 0)
	err := c.s.QueryTx(ctx, tx, &condition, &objs)
	if err != nil {
		return nil, err
	}

	return objs, nil
}

type ContentPropertyCondition struct {
	ContentIDs dbo.NullStrings
}

func (c *ContentPropertyCondition) GetOrderBy() string {
	return ""
}

func (c *ContentPropertyCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.ContentIDs.Valid {
		wheres = append(wheres, "content_id in (?)")
		params = append(params, c.ContentIDs.Strings)
	}

	return wheres, params
}

func (c *ContentPropertyCondition) GetPager() *dbo.Pager {
	return &dbo.NoPager
}

var (
	contentPropertyDA    IContentPropertyDA
	_contentPropertyOnce sync.Once
)

func GetContentPropertyDA() IContentPropertyDA {
	_contentPropertyOnce.Do(func() {
		contentPropertyDA = new(ContentPropertyDA)
	})

	return contentPropertyDA
}
