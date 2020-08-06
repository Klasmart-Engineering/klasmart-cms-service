package da

import (
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type ITagDA interface {
	Insert(ctx context.Context, tag *entity.Tag) error
	Update(ctx context.Context, tag *entity.Tag) error
	Query(ctx context.Context, condition *TagCondition) ([]*entity.Tag, error)
	GetByID(ctx context.Context, id string) (*entity.Tag, error)
	GetByIDs(ctx context.Context, ids []string) ([]*entity.Tag, error)
	Delete(ctx context.Context, id string) error
	Page(ctx context.Context, condition *TagCondition) (int64, []*entity.Tag, error)
}

type TagCondition struct {
	Name entity.NullString

	Pager utils.Pager

	DeleteAt entity.NullInt
}

func (t TagCondition) GetCondition() (expression.Expression, error) {
	var filt expression.ConditionBuilder
	if t.DeleteAt.Valid {
		filt = expression.Name("deleted_at").NotEqual(expression.Value(0))
	} else {
		filt = expression.Name("deleted_at").Equal(expression.Value(0))
	}
	if t.Name.Valid {
		filt = filt.And(expression.Name("name").Equal(expression.Value(t.Name)))
	}
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return expression.Expression{}, err
	}
	return expr, nil
}
