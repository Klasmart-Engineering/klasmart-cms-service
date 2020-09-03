package da

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type OutcomeSqlDA struct {
	dbo.BaseDA
}

type OutcomeCondition struct {
	Name          sql.NullString
	Description   sql.NullString
	Keyword       sql.NullString
	Shortcode     sql.NullString
	PublishStatus sql.NullString

	OrderBy OutcomeOrderBy `json:"order_by"`
	Pager   utils.Pager
}

type OutcomeOrderBy int

const (
	_ = iota
	OrderByName
	OrderByNameDesc
	OrderByCreatedAt
	OrderByCreatedAtDesc
)

func (s *OutcomeCondition) GetConditions() ([]string, []interface{}) {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	conditions = append(conditions, " delete_at = 0")
	return conditions, params
}
func (s *OutcomeCondition) GetPager() *dbo.Pager {
	return &dbo.Pager{
		Page:     int(s.Pager.PageIndex),
		PageSize: int(s.Pager.PageSize),
	}
}
func (s *OutcomeCondition) GetOrderBy() string {
	switch s.OrderBy {
	case OrderByName:
		return "name"
	case OrderByNameDesc:
		return "name desc"
	case OrderByCreatedAt:
		return "create_at"
	case OrderByCreatedAtDesc:
		return "create_at desc"
	default:
		return "create_at desc"
	}
}

func (o OutcomeSqlDA) CreateOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) error {
	panic("implement me")
}

func (o OutcomeSqlDA) UpdateOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) error {
	panic("implement me")
}

func (o OutcomeSqlDA) DeleteOutcome(ctx context.Context, tx *dbo.DBContext, id string) error {
	panic("implement me")
}

func (o OutcomeSqlDA) GetOutcomeByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Outcome, error) {
	panic("implement me")
}

func (o OutcomeSqlDA) SearchOutcome(ctx context.Context, tx *dbo.DBContext, condition *OutcomeCondition) (total int, outcomes []*entity.Outcome, err error) {
	total, err = o.PageTx(ctx, tx, condition, &outcomes)
	if err != nil {
		log.Error(ctx, "SearchOutcome failed",
			log.Err(err),
			log.Any("condition", condition))
	}
	return
}
