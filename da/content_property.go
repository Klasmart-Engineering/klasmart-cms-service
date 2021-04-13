package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IContentPropertyDA interface {
	BatchAdd(ctx context.Context, tx *dbo.DBContext, co []*entity.ContentProperty) error
	CleanByContentID(ctx context.Context, tx *dbo.DBContext, contentID string) error

	BatchGetByContentIDList(ctx context.Context, tx *dbo.DBContext, contentID []string) ([]*entity.ContentProperty, error)
}
type ContentPropertyDA struct {
	s dbo.BaseDA
}

func (c *ContentPropertyDA) BatchAdd(ctx context.Context, tx *dbo.DBContext, co []*entity.ContentProperty) error {
	var data [][]interface{}
	for _, item := range co {
		data = append(data, []interface{}{
			item.PropertyType,
			item.ContentID,
			item.PropertyID,
			item.Sequence,
		})
	}
	format, values := SQLBatchInsert(entity.ContentProperty{}.TableName(), []string{
		"`property_type`",
		"`content_id`",
		"`property_id`",
		"`sequence`",
	}, data)
	execResult := tx.Exec(format, values...)
	if execResult.Error != nil {
		logger.Error(ctx, "db exec sql error", log.Any("format", format), log.Any("values", values), log.Err(execResult.Error))
		return execResult.Error
	}
	return nil
}

func (c *ContentPropertyDA) CleanByContentID(ctx context.Context, tx *dbo.DBContext, contentID string) error {
	err := tx.Where("content_id = ?", contentID).Delete(entity.ContentProperty{}).Error
	if err != nil {
		logger.Error(ctx, "db exec sql error", log.String("contentID", contentID), log.Err(err))
		return err
	}
	return nil
}

func (c *ContentPropertyDA) BatchGetByContentIDList(ctx context.Context, tx *dbo.DBContext, contentID []string) ([]*entity.ContentProperty, error) {
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
