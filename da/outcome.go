package da

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IOutcomeDA interface {
	dbo.DataAccesser
	CreateOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) error
	UpdateOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) error
	DeleteOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) error
	BatchLockOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, ids []string) error

	GetOutcomeByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Outcome, error)
	GetOutcomeBySourceID(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, sourceID string) (*entity.Outcome, error)
	SearchOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, condition *OutcomeCondition) (int, []*entity.Outcome, error)

	UpdateLatestHead(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, oldHeader, newHeader string) error

	// return the min shortcode in each idle interval, excluding deleted
	GetIdleShortcode(ctx context.Context, tx *dbo.DBContext, orgID string) ([]int, error)
	GetLargerShortcode(ctx context.Context, tx *dbo.DBContext, orgID string, shortcodeNum, size int) ([]int, error)
}

type outcomeDA struct {
	dbo.BaseDA
}

var _outcomeOnce sync.Once
var _outcomeDA IOutcomeDA

func GetOutcomeDA() IOutcomeDA {
	_outcomeOnce.Do(func() {
		_outcomeDA = new(outcomeDA)
	})
	return _outcomeDA
}

type OutcomeCondition struct {
	IDs                    dbo.NullStrings
	Name                   sql.NullString
	Description            sql.NullString
	Keywords               sql.NullString
	ShortcodeLike          sql.NullString
	Shortcodes             dbo.NullStrings
	PublishStatus          dbo.NullStrings
	PublishScope           sql.NullString
	AuthorName             sql.NullString
	AuthorID               sql.NullString
	Assumed                sql.NullBool
	IsLocked               sql.NullBool
	OrganizationID         sql.NullString
	SourceID               sql.NullString
	SourceIDs              dbo.NullStrings
	FuzzyKey               sql.NullString
	AuthorIDs              dbo.NullStrings
	AncestorIDs            dbo.NullStrings
	RelationProgramIDs     dbo.NullStrings
	RelationSubjectIDs     dbo.NullStrings
	RelationCategoryIDs    dbo.NullStrings
	RelationSubCategoryIDs dbo.NullStrings
	RelationAgeIDs         dbo.NullStrings
	RelationGradeIDs       dbo.NullStrings

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

	if c.ShortcodeLike.Valid {
		wheres = append(wheres, "match(shortcode) against(? in boolean mode)")
		params = append(params, c.ShortcodeLike.String)
	}

	if c.Keywords.Valid {
		wheres = append(wheres, "match(keywords) against(? in boolean mode)")
		params = append(params, c.Keywords.String)
	}

	if c.Description.Valid {
		wheres = append(wheres, "match(description) against(? in boolean mode)")
		params = append(params, c.Description.String)
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

	if c.SourceIDs.Valid {
		wheres = append(wheres, "source_id in (?)")
		params = append(params, c.SourceIDs.Strings)
	}

	if c.AncestorIDs.Valid {
		wheres = append(wheres, "ancestor_id in (?)")
		params = append(params, c.AncestorIDs.Strings)
	}

	if c.Assumed.Valid {
		wheres = append(wheres, "assumed=?")
		params = append(params, c.Assumed.Bool)
	}

	if c.IsLocked.Valid {
		if c.IsLocked.Bool {
			wheres = append(wheres, "locked_by != ? AND locked_by != ?")
			params = append(params, "", constant.LockedByNoBody)
		} else {
			wheres = append(wheres, "(locked_by = ? || locked_by = ?)")
			params = append(params, "", constant.LockedByNoBody)
		}
	}

	if c.RelationProgramIDs.Valid {
		sql := fmt.Sprintf(`exists(SELECT 1 FROM %s WHERE relation_id IN (?) AND relation_type = "%s" AND %s.id = %s.master_id)`,
			entity.OutcomeRelationTable, entity.ProgramType, entity.OutcomeTable, entity.OutcomeRelationTable)
		wheres = append(wheres, sql)
		params = append(params, c.RelationProgramIDs.Strings)
	}

	if c.RelationSubjectIDs.Valid {
		sql := fmt.Sprintf(`exists(SELECT 1 FROM %s WHERE relation_id IN (?) AND relation_type = "%s" AND %s.id = %s.master_id)`,
			entity.OutcomeRelationTable, entity.SubjectType, entity.OutcomeTable, entity.OutcomeRelationTable)
		wheres = append(wheres, sql)
		params = append(params, c.RelationSubjectIDs.Strings)
	}

	if c.RelationCategoryIDs.Valid {
		sql := fmt.Sprintf(`exists(SELECT 1 FROM %s WHERE relation_id IN (?) AND relation_type = "%s" AND %s.id = %s.master_id)`,
			entity.OutcomeRelationTable, entity.CategoryType, entity.OutcomeTable, entity.OutcomeRelationTable)
		wheres = append(wheres, sql)
		params = append(params, c.RelationCategoryIDs.Strings)
	}

	if c.RelationSubCategoryIDs.Valid {
		sql := fmt.Sprintf(`exists(SELECT 1 FROM %s WHERE relation_id IN (?) AND relation_type = "%s" AND %s.id = %s.master_id)`,
			entity.OutcomeRelationTable, entity.SubcategoryType, entity.OutcomeTable, entity.OutcomeRelationTable)
		wheres = append(wheres, sql)
		params = append(params, c.RelationSubCategoryIDs.Strings)
	}

	if c.RelationAgeIDs.Valid {
		sql := fmt.Sprintf(`exists(SELECT 1 FROM %s WHERE relation_id IN (?) AND relation_type = "%s" AND %s.id = %s.master_id)`,
			entity.OutcomeRelationTable, entity.AgeType, entity.OutcomeTable, entity.OutcomeRelationTable)
		wheres = append(wheres, sql)
		params = append(params, c.RelationAgeIDs.Strings)
	}

	if c.RelationGradeIDs.Valid {
		sql := fmt.Sprintf(`exists(SELECT 1 FROM %s WHERE relation_id IN (?) AND relation_type = "%s" AND %s.id = %s.master_id)`,
			entity.OutcomeRelationTable, entity.GradeType, entity.OutcomeTable, entity.OutcomeRelationTable)
		wheres = append(wheres, sql)
		params = append(params, c.RelationGradeIDs.Strings)
	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "delete_at=0")
	}

	return wheres, params
}

func NewOutcomeCondition(condition *entity.OutcomeCondition) *OutcomeCondition {
	cond := &OutcomeCondition{
		IDs:            dbo.NullStrings{Strings: condition.IDs, Valid: len(condition.IDs) > 0},
		Name:           sql.NullString{String: condition.OutcomeName, Valid: condition.OutcomeName != ""},
		Description:    sql.NullString{String: condition.Description, Valid: condition.Description != ""},
		Keywords:       sql.NullString{String: condition.Keywords, Valid: condition.Keywords != ""},
		ShortcodeLike:  sql.NullString{String: condition.Shortcode, Valid: condition.Shortcode != ""},
		PublishStatus:  dbo.NullStrings{Strings: []string{condition.PublishStatus}, Valid: condition.PublishStatus != ""},
		PublishScope:   sql.NullString{String: condition.PublishScope, Valid: condition.PublishScope != ""},
		AuthorID:       sql.NullString{String: condition.AuthorID, Valid: condition.AuthorID != ""},
		OrganizationID: sql.NullString{String: condition.OrganizationID, Valid: condition.OrganizationID != ""},
		FuzzyKey:       sql.NullString{String: condition.FuzzyKey, Valid: condition.FuzzyKey != ""},
		AuthorIDs:      dbo.NullStrings{Strings: condition.AuthorIDs, Valid: len(condition.AuthorIDs) > 0},
		Assumed:        sql.NullBool{Bool: condition.Assumed == 1, Valid: condition.Assumed != -1},

		RelationProgramIDs:     dbo.NullStrings{Strings: condition.ProgramIDs, Valid: len(condition.ProgramIDs) > 0},
		RelationSubjectIDs:     dbo.NullStrings{Strings: condition.SubjectIDs, Valid: len(condition.SubjectIDs) > 0},
		RelationCategoryIDs:    dbo.NullStrings{Strings: condition.CategoryIDs, Valid: len(condition.CategoryIDs) > 0},
		RelationSubCategoryIDs: dbo.NullStrings{Strings: condition.SubCategoryIDs, Valid: len(condition.SubCategoryIDs) > 0},
		RelationAgeIDs:         dbo.NullStrings{Strings: condition.AgeIDs, Valid: len(condition.AgeIDs) > 0},
		RelationGradeIDs:       dbo.NullStrings{Strings: condition.GradeIDs, Valid: len(condition.GradeIDs) > 0},

		OrderBy: NewOrderBy(condition.OrderBy),
		Pager:   NewPage(condition.Page, condition.PageSize),
	}

	if condition.IsLocked != nil {
		cond.IsLocked = sql.NullBool{
			Bool:  *condition.IsLocked,
			Valid: true,
		}
	}

	return cond
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

func (o *outcomeDA) GetLargerShortcode(ctx context.Context, tx *dbo.DBContext, orgID string, shortcodeNum, size int) ([]int, error) {
	tx.ResetCondition()

	var result []int
	err := tx.Table(entity.OutcomeTable).
		Where("organization_id = ? and shortcode_num >= ?", orgID, shortcodeNum).
		Order("shortcode_num asc").
		Limit(size).
		Pluck("shortcode_num", &result).Error
	if err != nil {
		log.Error(ctx, "GetLargerShortcode error",
			log.Err(err),
			log.String("orgID", orgID),
			log.Int("shortcodeNum", shortcodeNum),
			log.Int("size", size))
		return nil, err
	}

	return result, nil
}

func (o *outcomeDA) GetIdleShortcode(ctx context.Context, tx *dbo.DBContext, orgID string) ([]int, error) {
	tx.ResetCondition()

	var result []int
	sql := `
	SELECT DISTINCT
		a.shortcode_num 
	FROM
		learning_outcomes AS a
		LEFT JOIN learning_outcomes AS b ON b.shortcode_num = a.shortcode_num + 1 
		AND a.organization_id = b.organization_id 
	WHERE
		a.organization_id = ? 
		AND b.shortcode_num IS NULL 
	ORDER BY
		a.shortcode_num ASC;
	`
	err := tx.Raw(sql, orgID).
		Scan(&result).Error
	if err != nil {
		log.Error(ctx, "GetMaxConsecutiveShortcode error",
			log.Err(err),
			log.String("orgID", orgID),
			log.String("sql", sql))
		return nil, err
	}

	for i := range result {
		result[i] = result[i] + 1
	}

	return result, nil
}

func (o *outcomeDA) CreateOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
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
	return
}

func (o *outcomeDA) UpdateOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	_, err = o.UpdateTx(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "UpdateOutcome: UpdateTx failed", log.Err(err), log.Any("outcome", outcome))
	}
	return
}

func (o *outcomeDA) DeleteOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) (err error) {
	now := time.Now().Unix()
	outcome.UpdateAt = now
	outcome.DeleteAt = now
	_, err = o.UpdateTx(ctx, tx, outcome)
	if err != nil {
		log.Error(ctx, "DeleteOutcome: UpdateTx failed", log.Err(err), log.Any("outcome", outcome))
	}
	return
}

func (o *outcomeDA) BatchLockOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, ids []string) (err error) {
	tx.ResetCondition()

	err = tx.Table(entity.OutcomeTable).
		Where("id in (?)", ids).
		Updates(map[string]interface{}{
			"locked_by": op.UserID,
		}).Error
	if err != nil {
		log.Error(ctx, "BatchLockOutcome error",
			log.Err(err),
			log.Strings("ids", ids))
		return err
	}

	return nil
}

func (o *outcomeDA) GetOutcomeByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Outcome, error) {
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
	return &outcome, nil
}

func (o *outcomeDA) GetOutcomeBySourceID(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, sourceID string) (*entity.Outcome, error) {
	condition := OutcomeCondition{SourceID: sql.NullString{String: sourceID, Valid: true}}
	var outcome entity.Outcome
	err := o.QueryTx(ctx, tx, &condition, &outcome)
	if err != nil {
		log.Error(ctx, "GetOutcomeBySourceID: GetTx failed",
			log.Err(err),
			log.String("source_id", sourceID))
		return nil, err
	}
	return &outcome, nil
}

func (o *outcomeDA) SearchOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, condition *OutcomeCondition) (total int, outcomes []*entity.Outcome, err error) {
	total, err = o.PageTx(ctx, tx, condition, &outcomes)
	if err != nil {
		log.Error(ctx, "SearchOutcome failed",
			log.Err(err),
			log.Any("condition", condition))
		return
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
	return
}

func (o *outcomeDA) UpdateLatestHead(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, oldHeader, newHeader string) error {
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
