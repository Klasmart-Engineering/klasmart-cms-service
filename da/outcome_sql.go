package da

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
	"time"
)

type OutcomeSqlDA struct {
	dbo.BaseDA
}

type OutcomeCondition struct {
	IDs            dbo.NullStrings
	Name           sql.NullString
	Description    sql.NullString
	Keywords       sql.NullString
	Shortcode      sql.NullString
	PublishStatus  dbo.NullStrings
	PublishScope   sql.NullString
	AuthorName     sql.NullString
	AuthorID       sql.NullString
	Assumed        sql.NullBool
	OrganizationID sql.NullString
	SourceID       sql.NullString
	FuzzyKey       sql.NullString

	OrderBy OutcomeOrderBy `json:"order_by"`
	Pager   dbo.Pager
}

func (c *OutcomeCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	//var fullMatchValue []string
	//if c.Name.Valid {
	//	fullMatchValue = append(fullMatchValue, "+"+c.Name.String)
	//}
	//if c.AuthorName.Valid {
	//	fullMatchValue = append(fullMatchValue, "+"+c.AuthorName.String)
	//}
	//if c.Keywords.Valid {
	//	fullMatchValue = append(fullMatchValue, "+"+c.Keywords.String)
	//}
	//if c.Shortcode.Valid {
	//	fullMatchValue = append(fullMatchValue, "+"+c.Shortcode.String)
	//}
	//if c.Description.Valid {
	//	fullMatchValue = append(fullMatchValue, "+"+c.Description.String)
	//}
	//
	//if len(fullMatchValue) > 0 {
	//	wheres = append(wheres, "match(name, keywords, description, author_name, shortcode) against(? in boolean mode)")
	//	params = append(params, strings.TrimSpace(strings.Join(fullMatchValue, " ")))
	//}

	if c.FuzzyKey.Valid {
		wheres = append(wheres, "match(name, keywords, description, author_name, shortcode) against(? in boolean mode)")
		params = append(params, strings.TrimSpace(c.FuzzyKey.String))
	}

	if c.IDs.Valid {
		inValue := strings.TrimSuffix(strings.Repeat("?,", len(c.IDs.Strings)), ",")
		wheres = append(wheres, fmt.Sprintf("id in (%s)", inValue))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}

	if c.PublishStatus.Valid {
		inValue := strings.TrimSuffix(strings.Repeat("?,", len(c.PublishStatus.Strings)), ",")
		wheres = append(wheres, fmt.Sprintf("publish_status in (%s)", inValue))
		params = append(params, c.PublishStatus.ToInterfaceSlice()...)
	}

	if c.PublishScope.Valid {
		wheres = append(wheres, "publish_scope=?")
		params = append(params, c.PublishScope.String)
	}

	if c.AuthorID.Valid {
		wheres = append(wheres, "author_id=?")
		params = append(params, c.AuthorID.String)
	}

	if c.OrganizationID.Valid {
		wheres = append(wheres, "organization_id=?")
		params = append(params, c.OrganizationID.String)
	}

	if c.SourceID.Valid {
		wheres = append(wheres, "source_id=?")
		params = append(params, c.SourceID.String)
	}

	if c.Assumed.Valid {
		wheres = append(wheres, "assumed=?")
		params = append(params, c.Assumed.Bool)
	}

	wheres = append(wheres, " delete_at = 0")
	return wheres, params
}

func NewOutcomeCondition(condition *entity.OutcomeCondition) *OutcomeCondition {
	return &OutcomeCondition{
		IDs:            dbo.NullStrings{Strings: condition.IDs, Valid: len(condition.IDs) > 0},
		Name:           sql.NullString{String: condition.OutcomeName, Valid: condition.OutcomeName != ""},
		Description:    sql.NullString{String: condition.Description, Valid: condition.Description != ""},
		Keywords:       sql.NullString{String: condition.Keywords, Valid: condition.Keywords != ""},
		Shortcode:      sql.NullString{String: condition.Shortcode, Valid: condition.Shortcode != ""},
		PublishStatus:  dbo.NullStrings{Strings: []string{condition.PublishStatus}, Valid: condition.PublishStatus != ""},
		PublishScope:   sql.NullString{String: condition.PublishScope, Valid: condition.PublishScope != ""},
		AuthorID:       sql.NullString{String: condition.AuthorID, Valid: condition.AuthorID != ""},
		AuthorName:     sql.NullString{String: condition.AuthorName, Valid: condition.AuthorName != ""},
		OrganizationID: sql.NullString{String: condition.OrganizationID, Valid: condition.OrganizationID != ""},
		FuzzyKey:       sql.NullString{String: condition.FuzzyKey, Valid: condition.FuzzyKey != ""},
		Assumed:        sql.NullBool{Bool: condition.Assumed == 1, Valid: condition.Assumed != -1},
		OrderBy:        NewOrderBy(condition.OrderBy),
		Pager:          NewPage(condition.Page, condition.PageSize),
	}
}

type OutcomeOrderBy int

const (
	_ = iota
	OrderByName
	OrderByNameDesc
	OrderByCreatedAt
	OrderByCreatedAtDesc
)

const defaultPageIndex = 1
const defaultPageSize = 20

func NewPage(page, pageSize int) dbo.Pager {
	if page == -1 {
		return dbo.NoPager
	}
	if page == 0 {
		page = defaultPageIndex
	}
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	return dbo.Pager{
		Page:     page,
		PageSize: pageSize,
	}
}

func (c *OutcomeCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func NewOrderBy(name string) OutcomeOrderBy {
	switch name {
	case "name":
		return OrderByName
	case "-name":
		return OrderByNameDesc
	case "created_at":
		return OrderByCreatedAt
	case "-created_at":
		return OrderByCreatedAtDesc
	default:
		return OrderByCreatedAtDesc
	}
}

func (c *OutcomeCondition) GetOrderBy() string {
	switch c.OrderBy {
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

func (o OutcomeSqlDA) GetOutcomeBySourceID(ctx context.Context, tx *dbo.DBContext, sourceID string) (*entity.Outcome, error) {
	var outcome entity.Outcome
	err := o.QueryTx(ctx, tx, &OutcomeCondition{SourceID: sql.NullString{String: sourceID, Valid: true}}, &outcome)
	//sql := fmt.Sprintf("select * from %s where source_id='%s'", outcome.TableName(), sourceID)
	//err := tx.Raw(sql).Scan(&outcome).Error
	if err != nil {
		log.Error(ctx, "GetOutcomeBySourceID: GetTx failed",
			log.Err(err),
			log.String("source_id", sourceID))
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

func (o OutcomeSqlDA) UpdateLatestHead(ctx context.Context, tx *dbo.DBContext, oldHeader, newHeader string) error {
	sql := fmt.Sprintf("update %s set latest_id='%s' where latest_id='%s' and delete_at=0", entity.Outcome{}.TableName(), newHeader, oldHeader)
	err := tx.Exec(sql).Error
	if err != nil {
		log.Error(ctx, "UpdateLatestHead failed",
			log.Err(err),
			log.String("old", oldHeader),
			log.String("new", newHeader),
			log.String("sql", sql))
		return err
	}
	return nil
}
