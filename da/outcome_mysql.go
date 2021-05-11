package da

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type OutcomeSQLDA struct {
	dbo.BaseDA
}

type OutcomeCondition struct {
	IDs                dbo.NullStrings
	Name               sql.NullString
	Description        sql.NullString
	Keywords           sql.NullString
	Shortcode          sql.NullString
	ShortcodeCommonKey sql.NullString
	Shortcodes         dbo.NullStrings
	PublishStatus      dbo.NullStrings
	PublishScope       sql.NullString
	AuthorName         sql.NullString
	AuthorID           sql.NullString
	Assumed            sql.NullBool
	OrganizationID     sql.NullString
	SourceID           sql.NullString
	FuzzyKey           sql.NullString
	AuthorIDs          dbo.NullStrings
	AncestorIDs        dbo.NullStrings

	IncludeDeleted bool
	OrderBy        OutcomeOrderBy `json:"order_by"`
	Pager          dbo.Pager
}

func (c *OutcomeCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.FuzzyKey.Valid {
		clauses := []string{"match(name, keywords, description, shortcode) against(? in boolean mode)"}
		params = append(params, strings.TrimSpace(c.FuzzyKey.String))

		if c.AuthorIDs.Valid {
			clauses = append(clauses, fmt.Sprintf("(author_id in (%s))", c.AuthorIDs.SQLPlaceHolder()))
			params = append(params, c.AuthorIDs.ToInterfaceSlice()...)
		}

		if c.IDs.Valid {
			clauses = append(clauses, fmt.Sprintf("(id in (%s))", c.IDs.SQLPlaceHolder()))
			params = append(params, c.IDs.ToInterfaceSlice()...)
		}
		wheres = append(wheres, fmt.Sprintf("(%s)", strings.Join(clauses, " or ")))
	}

	if !c.FuzzyKey.Valid && c.AuthorIDs.Valid {
		wheres = append(wheres, fmt.Sprintf("author_id in (%s)", c.AuthorIDs.SQLPlaceHolder()))
		params = append(params, c.AuthorIDs.ToInterfaceSlice()...)
	}

	if !c.FuzzyKey.Valid && c.IDs.Valid {
		wheres = append(wheres, fmt.Sprintf("id in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}

	if c.Name.Valid {
		wheres = append(wheres, "match(name) against(? in boolean mode)")
		//wheres = append(wheres, "name=?")
		params = append(params, c.Name.String)
	}

	if c.Shortcode.Valid {
		wheres = append(wheres, "match(shortcode) against(? in boolean mode)")
		params = append(params, c.Shortcode.String)
	}

	if c.Keywords.Valid {
		wheres = append(wheres, "match(keywords) against(? in boolean mode)")
		//wheres = append(wheres, "keywords=?")
		params = append(params, c.Keywords.String)
	}

	if c.Description.Valid {
		wheres = append(wheres, "match(description) against(? in boolean mode)")
		//wheres = append(wheres, "description=?")
		params = append(params, c.Description.String)
	}

	if c.ShortcodeCommonKey.Valid {
		wheres = append(wheres, "shortcode=?")
		params = append(params, c.ShortcodeCommonKey.String)
	}

	if c.Shortcodes.Valid {
		wheres = append(wheres, "shortcode in (?)")
		params = append(params, c.Shortcodes.Strings)
	}

	if c.PublishStatus.Valid {
		wheres = append(wheres, fmt.Sprintf("publish_status in (%s)", c.PublishStatus.SQLPlaceHolder()))
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

	if c.AncestorIDs.Valid {
		wheres = append(wheres, "ancestor_id in (?)")
		params = append(params, c.AncestorIDs.Strings)
	}

	if c.Assumed.Valid {
		wheres = append(wheres, "assumed=?")
		params = append(params, c.Assumed.Bool)
	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "delete_at=0")
	}
	return wheres, params
}

func NewOutcomeCondition(condition *entity.OutcomeCondition) *OutcomeCondition {
	return &OutcomeCondition{
		IDs:           dbo.NullStrings{Strings: condition.IDs, Valid: len(condition.IDs) > 0},
		Name:          sql.NullString{String: condition.OutcomeName, Valid: condition.OutcomeName != ""},
		Description:   sql.NullString{String: condition.Description, Valid: condition.Description != ""},
		Keywords:      sql.NullString{String: condition.Keywords, Valid: condition.Keywords != ""},
		Shortcode:     sql.NullString{String: condition.Shortcode, Valid: condition.Shortcode != ""},
		PublishStatus: dbo.NullStrings{Strings: []string{condition.PublishStatus}, Valid: condition.PublishStatus != ""},
		PublishScope:  sql.NullString{String: condition.PublishScope, Valid: condition.PublishScope != ""},
		AuthorID:      sql.NullString{String: condition.AuthorID, Valid: condition.AuthorID != ""},
		//AuthorName:     sql.NullString{String: condition.AuthorName, Valid: condition.AuthorName != ""},
		OrganizationID: sql.NullString{String: condition.OrganizationID, Valid: condition.OrganizationID != ""},
		FuzzyKey:       sql.NullString{String: condition.FuzzyKey, Valid: condition.FuzzyKey != ""},
		AuthorIDs:      dbo.NullStrings{Strings: condition.AuthorIDs, Valid: len(condition.AuthorIDs) > 0},
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
	OrderByUpdateAt
	OrderByUpdateAtDesc
	OrderByShortcode
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
		// require by pm
		//return OrderByCreatedAt
		return OrderByUpdateAt
	case "-created_at":
		// require by pm
		//return OrderByCreatedAtDesc
		return OrderByUpdateAtDesc
	case "updated_at":
		return OrderByUpdateAt
	case "-updated_at":
		return OrderByUpdateAtDesc
	default:
		return OrderByUpdateAtDesc
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
	case OrderByUpdateAt:
		return "update_at"
	case OrderByUpdateAtDesc:
		return "update_at desc"
	default:
		return "update_at desc"
	}
}

func (o OutcomeSQLDA) CreateOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
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
		return
	}
	if outcome.SourceID != "" && outcome.SourceID != constant.LockedByNoBody && outcome.SourceID != outcome.ID {
		GetOutcomeRedis().CleanOutcomeCache(ctx, op, []string{outcome.ID, outcome.SourceID})
	} else {
		GetOutcomeRedis().CleanOutcomeCache(ctx, op, []string{outcome.ID})
	}
	return
}

func (o OutcomeSQLDA) UpdateOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	_, err = o.UpdateTx(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "UpdateOutcome: UpdateTx failed", log.Err(err), log.Any("outcome", outcome))
	}
	GetOutcomeRedis().CleanOutcomeCache(ctx, op, []string{outcome.ID})
	return
}

func (o OutcomeSQLDA) DeleteOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	now := time.Now().Unix()
	outcome.UpdateAt = now
	outcome.DeleteAt = now
	_, err = o.UpdateTx(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "DeleteOutcome: UpdateTx failed", log.Err(err), log.Any("outcome", outcome))
	}
	GetOutcomeRedis().CleanOutcomeCache(ctx, op, []string{outcome.ID})
	return
}

func (o OutcomeSQLDA) GetOutcomeByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Outcome, error) {
	hit := GetOutcomeRedis().GetOutcomeCacheByID(ctx, id)
	if hit != nil {
		return hit, nil
	}
	// not hit
	var outcome entity.Outcome
	err := o.GetTx(ctx, tx, id, &outcome)
	if err != nil {
		log.Error(ctx, "GetOutcomeByID: GetTx failed", log.Err(err), log.Any("outcome", outcome))
		return nil, err
	}
	outcomeSet, err := GetOutcomeSetDA().SearchSetsByOutcome(ctx, tx, []string{outcome.ID})
	if err != nil {
		log.Error(ctx, "GetOutcomeByID: SearchSetsByOutcome failed", log.Err(err), log.Any("outcome", outcome))
		return nil, err
	}
	outcome.Sets = outcomeSet[outcome.ID]
	GetOutcomeRedis().SaveOutcomeCache(ctx, &outcome)
	return &outcome, nil
}

func (o OutcomeSQLDA) GetOutcomeBySourceID(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, sourceID string) (*entity.Outcome, error) {
	condition := OutcomeCondition{SourceID: sql.NullString{String: sourceID, Valid: true}}
	hits := GetOutcomeRedis().GetOutcomeCacheBySearchCondition(ctx, op, &condition)
	if hits != nil && len(hits.OutcomeList) == 1 {
		return hits.OutcomeList[0], nil
	}

	// not hit
	var outcome entity.Outcome
	err := o.QueryTx(ctx, tx, &condition, &outcome)
	if err != nil {
		log.Error(ctx, "GetOutcomeBySourceID: GetTx failed",
			log.Err(err),
			log.String("source_id", sourceID))
		return nil, err
	}
	GetOutcomeRedis().SaveOutcomeCacheListBySearchCondition(ctx, op, &condition, &OutcomeListWithKey{1, []*entity.Outcome{&outcome}})
	return &outcome, nil
}

func (o OutcomeSQLDA) SearchOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, condition *OutcomeCondition) (total int, outcomes []*entity.Outcome, err error) {
	hits := GetOutcomeRedis().GetOutcomeCacheBySearchCondition(ctx, op, condition)
	if hits != nil {
		return hits.Total, hits.OutcomeList, nil
	}
	// not hit
	total, err = o.PageTx(ctx, tx, condition, &outcomes)
	if err != nil {
		log.Error(ctx, "SearchOutcome failed",
			log.Err(err),
			log.Any("condition", condition))
	}
	outcomeIDs := make([]string, len(outcomes))
	for i := range outcomes {
		outcomeIDs[i] = outcomes[i].ID
	}
	if len(outcomeIDs) > 0 {
		outcomeSets, err := GetOutcomeSetDA().SearchSetsByOutcome(ctx, tx, outcomeIDs)
		if err != nil {
			log.Error(ctx, "GetOutcomeByID: SearchSetsByOutcome failed", log.Err(err), log.Strings("outcome", outcomeIDs))
			return 0, nil, err
		}
		for i := range outcomes {
			outcomes[i].Sets = outcomeSets[outcomes[i].ID]
		}
	}

	GetOutcomeRedis().SaveOutcomeCacheListBySearchCondition(ctx, op, condition, &OutcomeListWithKey{total, outcomes})
	return
}

func (o OutcomeSQLDA) UpdateLatestHead(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, oldHeader, newHeader string) error {
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

	// clean cache
	var outcomes []*entity.Outcome
	err = tx.Where("latest_id=? and delete_at=0", newHeader).Find(&outcomes).Error
	if err != nil {
		log.Error(ctx, "UpdateLatestHead: Find failed",
			log.Err(err),
			log.String("old", oldHeader),
			log.String("new", newHeader))
		return err
	}
	if len(outcomes) < 1 {
		log.Info(ctx, "UpdateLatestHead: Find outcomes return empty",
			log.String("old", oldHeader),
			log.String("new", newHeader))
		return nil
	}
	ids := make([]string, len(outcomes))
	for i := range outcomes {
		ids[i] = outcomes[i].ID
	}
	GetOutcomeRedis().CleanOutcomeCache(ctx, op, ids)
	return nil
}
