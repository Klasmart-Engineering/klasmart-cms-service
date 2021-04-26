package da

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type SetSqlDA struct {
	dbo.BaseDA
}

type SetCondition struct {
	ID             sql.NullString
	IDs            dbo.NullStrings
	Name           sql.NullString
	Names          dbo.NullStrings
	OrganizationID sql.NullString
	//FuzzyName      sql.NullString

	IncludeDeleted bool
	OrderBy        SetOrderBy `json:"order_by"`
	Pager          dbo.Pager
}

func (c *SetCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	//if c.FuzzyName.Valid {
	//	wheres = append(wheres, "match(name) against(? in boolean mode)")
	//	params = append(params, strings.TrimSpace(c.FuzzyName.String))
	//}

	if c.ID.Valid {
		wheres = append(wheres, "id=?")
		params = append(params, c.ID.String)
	}

	if c.IDs.Valid {
		wheres = append(wheres, fmt.Sprintf("id in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.IDs.ToInterfaceSlice()...)
	}

	if c.Name.Valid {
		wheres = append(wheres, "match(name) against(? in boolean mode)")
		params = append(params, strings.TrimSpace(c.Name.String))
		//wheres = append(wheres, "name=?")
		//params = append(params, strings.TrimSpace(c.Name.String))
	}

	if c.Names.Valid {
		wheres = append(wheres, fmt.Sprintf("name in (%s)", c.IDs.SQLPlaceHolder()))
		params = append(params, c.Names.ToInterfaceSlice()...)
	}

	if c.OrganizationID.Valid {
		wheres = append(wheres, "organization_id=?")
		params = append(params, c.OrganizationID.String)
	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "delete_at=0")
	}
	return wheres, params
}

type SetOrderBy int

const (
	_ = iota
	OrderBySetName
	OrderBySetNameDesc
	OrderBySetCreatedAt
	OrderBySetCreatedAtDesc
)

func (c *SetCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *SetCondition) GetOrderBy() string {
	switch c.OrderBy {
	case OrderBySetName:
		return "name"
	case OrderBySetNameDesc:
		return "name desc"
	case OrderBySetCreatedAt:
		return "create_at"
	case OrderBySetCreatedAtDesc:
		return "create_at desc"
	default:
		return "update_at desc"
	}
}

type OutcomeSetCondition struct {
	OutcomeIDs dbo.NullStrings
	SetIDs     dbo.NullStrings

	IncludeDeleted bool
	OrderBy        SetOrderBy `json:"order_by"`
	Pager          dbo.Pager
}

func (c *OutcomeSetCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.OutcomeIDs.Valid {
		wheres = append(wheres, fmt.Sprintf("outcome_id in (%s)", c.OutcomeIDs.SQLPlaceHolder()))
		params = append(params, c.OutcomeIDs.ToInterfaceSlice()...)
	}

	if c.SetIDs.Valid {
		wheres = append(wheres, fmt.Sprintf("set_id in (%s)", c.SetIDs.SQLPlaceHolder()))
		params = append(params, c.SetIDs.ToInterfaceSlice()...)
	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "delete_at is null")
	}
	return wheres, params
}
func (c *OutcomeSetCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *OutcomeSetCondition) GetOrderBy() string {
	switch c.OrderBy {
	case OrderBySetName:
		return "name"
	case OrderBySetNameDesc:
		return "name desc"
	case OrderBySetCreatedAt:
		return "create_at"
	case OrderBySetCreatedAtDesc:
		return "create_at desc"
	default:
		return "update_at desc"
	}
}
func (o SetSqlDA) CreateSet(ctx context.Context, tx *dbo.DBContext, set *entity.Set) (err error) {
	now := time.Now().Unix()
	if set.CreateAt == 0 {
		set.CreateAt = now
	}
	if set.UpdateAt == 0 {
		set.UpdateAt = now
	}
	_, err = o.InsertTx(ctx, tx, set)
	if err != nil {
		log.Error(ctx, "CreateSet: InsertTx failed", log.Err(err), log.Any("outcome_set", set))
		return
	}
	return
}

func (o SetSqlDA) UpdateOutcomeSet(ctx context.Context, tx *dbo.DBContext, set *entity.Set) (err error) {
	if set.UpdateAt == 0 {
		set.UpdateAt = time.Now().Unix()
	}
	_, err = o.UpdateTx(ctx, tx, set)
	if err != nil {
		log.Error(ctx, "UpdateOutcomeSet: UpdateTx failed", log.Err(err), log.Any("outcome_set", set))
	}
	return
}

func (o SetSqlDA) SearchSet(ctx context.Context, tx *dbo.DBContext, condition *SetCondition) (total int, sets []*entity.Set, err error) {
	total, err = o.PageTx(ctx, tx, condition, &sets)
	if err != nil {
		log.Error(ctx, "SearchSet failed",
			log.Err(err),
			log.Any("condition", condition))
	}
	return
}

func (o SetSqlDA) SearchOutcomeSet(ctx context.Context, tx *dbo.DBContext, condition *OutcomeSetCondition) (total int, outcomeSets []*entity.OutcomeSet, err error) {
	total, err = o.PageTx(ctx, tx, condition, &outcomeSets)
	if err != nil {
		log.Error(ctx, "SearchOutcomeSet failed",
			log.Err(err),
			log.Any("condition", condition))
	}
	return
}

func (o SetSqlDA) BulkBindOutcomeSet(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcomeSets []*entity.OutcomeSet) error {

	now := time.Now().Unix()
	sql := fmt.Sprintf("insert into %s(outcome_id, set_id, create_at, update_at) values", entity.OutcomeSet{}.TableName())
	for i := range outcomeSets {
		item := fmt.Sprintf("('%s', '%s', %d, %d),", outcomeSets[i].OutcomeID, outcomeSets[i].SetID, now, now)
		sql = sql + item
	}
	sql = strings.TrimSuffix(sql, ",")
	err := tx.Exec(sql).Error
	if err != nil {
		log.Error(ctx, "BulkBindOutcomeSet: failed",
			log.Err(err),
			log.Any("op", op),
			log.String("sql", sql))
		return err
	}
	return nil
}

func (o SetSqlDA) BindOutcomeSet(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcomeSets []*entity.OutcomeSet) error {
	now := time.Now().Unix()
	table := entity.OutcomeSet{}.TableName()
	values := make([][]interface{}, len(outcomeSets))
	for i := range outcomeSets {
		values[i] = []interface{}{
			outcomeSets[i].OutcomeID,
			outcomeSets[i].SetID,
			now + int64(i),
			now + int64(i),
		}
	}
	sql, result := SQLBatchInsert(table, []string{"outcome_id", "set_id", "create_at", "update_at"}, values)
	err := tx.Exec(sql, result...).Error
	if err != nil {
		log.Error(ctx, "BindOutcomeSet: failed",
			log.Err(err),
			log.Any("op", op),
			log.String("sql", sql))
		return err
	}
	return nil
}

func (o SetSqlDA) DeleteBoundOutcomeSet(ctx context.Context, tx *dbo.DBContext, outcomeID string) error {
	sql := fmt.Sprintf("delete from %s where outcome_id = ?", entity.OutcomeSet{}.TableName())
	err := tx.Exec(sql, outcomeID).Error
	if err != nil {
		log.Error(ctx, "DeleteBoundOutcomeSet: exec sql failed",
			log.Err(err),
			log.String("sql", sql))
		return err
	}
	return nil
}

func (o SetSqlDA) SearchOutcomeBySetName(ctx context.Context, op *entity.Operator, name string) ([]*entity.OutcomeSet, error) {
	sql := fmt.Sprintf("select * from %s where set_id in (select id from %s where match(name) against(? in boolean mode) and organization_id = ? and delete_at = 0) and delete_at is null",
		entity.OutcomeSet{}.TableName(), entity.Set{}.TableName())
	var outcomeSets []*entity.OutcomeSet
	err := dbo.MustGetDB(ctx).Raw(sql, name, op.OrgID).Scan(&outcomeSets).Error
	if err != nil {
		log.Error(ctx, "SearchOutcomeBySetName: exec sql failed",
			log.Err(err),
			log.Any("op", op),
			log.String("set_name", name))
		return nil, err
	}
	return outcomeSets, nil
}

func (o SetSqlDA) SearchSetsByOutcome(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string) (map[string][]*entity.Set, error) {
	sql := fmt.Sprintf("select distinct a.outcome_id, a.set_id, b.name from (select * from %s where outcome_id in (?)) as a left join %s as b on (a.set_id=b.id)",
		entity.OutcomeSet{}.TableName(), entity.Set{}.TableName())
	var result []*struct {
		OutcomeID string `gorm:"column:outcome_id"`
		SetID     string `gorm:"column:set_id"`
		Name      string `gorm:"column:name"`
	}
	err := dbo.MustGetDB(ctx).Raw(sql, outcomeIDs).Scan(&result).Error
	if err != nil {
		log.Error(ctx, "SearchSetsByOutcome: exec sql failed",
			log.Err(err),
			log.Strings("outcome", outcomeIDs))
		return nil, err
	}
	outcomesSets := make(map[string][]*entity.Set)
	for i := range result {
		set := entity.Set{
			ID:   result[i].SetID,
			Name: result[i].Name,
		}
		outcomesSets[result[i].OutcomeID] = append(outcomesSets[result[i].OutcomeID], &set)
	}
	return outcomesSets, nil
}

func (o SetSqlDA) IsSetExist(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, name string) (bool, error) {
	sql := fmt.Sprintf("select * from %s where name=? and organization_id=?", entity.Set{}.TableName())
	var sets []*entity.Set
	err := tx.Raw(sql, name, op.OrgID).Scan(&sets).Error
	if err != nil {
		log.Error(ctx, "IsSetExist: exec sql failed",
			log.Err(err),
			log.String("set_name", name))
		return false, err
	}

	for i := range sets {
		if sets[i].Name == name {
			return true, nil
		}
	}

	return false, nil
}
