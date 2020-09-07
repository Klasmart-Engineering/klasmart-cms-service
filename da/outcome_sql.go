package da

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"time"
)

type OutcomeSqlDA struct {
	dbo.BaseDA
}

type OutcomeCondition struct {
	IDs           dbo.NullStrings
	Name          sql.NullString
	Description   sql.NullString
	Keyword       sql.NullString
	Shortcode     sql.NullString
	PublishStatus dbo.NullStrings
	PublishScope  sql.NullString
	AuthorID      sql.NullString

	OrderBy  OutcomeOrderBy `json:"order_by"`
	Page     int
	PageSize int
}

func NewOutcomeCondition(condition *entity.OutcomeCondition) *OutcomeCondition {
	return &OutcomeCondition{}
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
		Page:     s.Page,
		PageSize: s.PageSize,
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

func (o OutcomeSqlDA) CreateOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	now := time.Now().Unix()
	if outcome.CreateAt == 0 {
		outcome.CreateAt = now
	}
	if outcome.UpdateAt == 0 {
		outcome.UpdateAt = now
	}
	_, err = o.InsertTx(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "CreateOutcome: InsertTx failed", log.Err(err), log.Any("outcome", outcome))
	}
	return
}

func (o OutcomeSqlDA) UpdateOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	now := time.Now().Unix()
	outcome.UpdateAt = now
	_, err = o.UpdateTx(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "UpdateOutcome: UpdateTx failed", log.Err(err), log.Any("outcome", outcome))
	}
	return
}

func (o OutcomeSqlDA) DeleteOutcome(ctx context.Context, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	now := time.Now().Unix()
	outcome.UpdateAt = now
	outcome.DeleteAt = now
	_, err = o.UpdateTx(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "DeleteOutcome: UpdateTx failed", log.Err(err), log.Any("outcome", outcome))
	}
	return
}

func (o OutcomeSqlDA) GetOutcomeByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Outcome, error) {
	var outcome entity.Outcome
	err := o.GetTx(ctx, tx, id, &outcome)
	if err != nil {
		log.Error(ctx, "GetOutcomeByID: GetTx failed", log.Err(err), log.Any("outcome", outcome))
		return nil, err
	}
	return &outcome, nil
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
