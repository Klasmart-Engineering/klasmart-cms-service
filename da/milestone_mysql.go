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

type MilestoneSQLDA struct {
	dbo.BaseDA
}

func (m MilestoneSQLDA) Create(ctx context.Context, tx *dbo.DBContext, milestone *entity.Milestone) error {
	_, err := m.InsertTx(ctx, tx, milestone)
	if err != nil {
		log.Error(ctx, "Create: InsertTx failed",
			log.Err(err),
			log.Any("milestone", milestone))
		return err
	}
	return nil
}

func (m MilestoneSQLDA) GetByID(ctx context.Context, tx *dbo.DBContext, ID string) (*entity.Milestone, error) {
	var milestone entity.Milestone
	err := m.GetTx(ctx, tx, ID, &milestone)
	if err != nil {
		log.Error(ctx, "GetByID: GetTx failed",
			log.Err(err),
			log.String("milestone", ID))
		return nil, err
	}
	return &milestone, nil
}

func (m MilestoneSQLDA) Update(ctx context.Context, tx *dbo.DBContext, milestone *entity.Milestone) error {
	_, err := m.BaseDA.UpdateTx(ctx, tx, milestone)
	if err != nil {
		log.Error(ctx, "Update: UpdateTx failed",
			log.Any("milestone", milestone))
		return err
	}
	return nil
}

func (m MilestoneSQLDA) Search(ctx context.Context, tx *dbo.DBContext, condition *MilestoneCondition) (int, []*entity.Milestone, error) {
	var milestones []*entity.Milestone
	total, err := m.BaseDA.PageTx(ctx, tx, condition, &milestones)
	if err != nil {
		log.Error(ctx, "Search: PageTx failed",
			log.Any("condition", condition))
		return 0, nil, err
	}
	return total, milestones, nil
}

func (m MilestoneSQLDA) BatchPublish(ctx context.Context, tx *dbo.DBContext, publishIDs []string) error {
	if len(publishIDs) > 0 {
		sql := fmt.Sprintf("update %s set status = ?, update_at = ? where id in (?)", entity.Milestone{}.TableName())
		err := tx.Exec(sql, entity.OutcomeStatusPublished, time.Now().Unix(), publishIDs).Error
		if err != nil {
			log.Error(ctx, "BatchPublish: exec sql failed",
				log.Err(err),
				log.Strings("publish", publishIDs),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

func (m MilestoneSQLDA) BatchHide(ctx context.Context, tx *dbo.DBContext, hideIDs []string) error {
	if len(hideIDs) > 0 {
		sql := fmt.Sprintf("update %s set status = ?, locked_by = ?, update_at = ? where id in (?)", entity.Milestone{}.TableName())
		err := tx.Exec(sql, entity.OutcomeStatusHidden, "", time.Now().Unix(), hideIDs).Error
		if err != nil {
			log.Error(ctx, "BatchHide: exec sql failed",
				log.Err(err),
				log.Strings("hide", hideIDs),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

func (m MilestoneSQLDA) BatchUnLock(ctx context.Context, tx *dbo.DBContext, unLockIDs []string) error {
	if len(unLockIDs) > 0 {
		sql := fmt.Sprintf("update %s set locked_by = ?, update_at = ? where id in (?)", entity.Milestone{}.TableName())
		err := tx.Exec(sql, "", time.Now().Unix(), unLockIDs).Error
		if err != nil {
			log.Error(ctx, "BatchUnLock: exec sql failed",
				log.Err(err),
				log.Strings("un_lock", unLockIDs),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

func (m MilestoneSQLDA) BatchUpdateLatest(ctx context.Context, tx *dbo.DBContext, ancestorLatest map[string]string) error {
	if len(ancestorLatest) > 0 {
		var sb strings.Builder
		fmt.Fprintf(&sb, "update %s set update_at= %d, latest_id = case ancestor_id ", entity.Milestone{}.TableName(), time.Now().Unix())
		ancestorIDs := make([]string, len(ancestorLatest))
		i := 0
		for k, v := range ancestorLatest {
			fmt.Fprintf(&sb, " when '%s' then '%s' ", k, v)
			ancestorIDs[i] = k
			i++
		}
		fmt.Fprintf(&sb, " end ")
		fmt.Fprintf(&sb, " where ancestor_id in (?)")
		sql := sb.String()
		err := tx.Exec(sql, ancestorIDs).Error
		if err != nil {
			log.Error(ctx, "BatchPublish: exec sql failed",
				log.Err(err),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

func (m MilestoneSQLDA) BatchDelete(ctx context.Context, tx *dbo.DBContext, milestoneIDs []string) error {
	if len(milestoneIDs) > 0 {
		sql := fmt.Sprintf("update %s set delete_at = %d where id in (?)", entity.Milestone{}.TableName(), time.Now().Unix())
		err := tx.Exec(sql, milestoneIDs).Error
		if err != nil {
			log.Error(ctx, "BatchDelete: exec sql failed",
				log.Err(err),
				log.Strings("delete", milestoneIDs),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

func (m MilestoneSQLDA) UnbindOutcomes(ctx context.Context, tx *dbo.DBContext, outcomeAncestors []string) error {
	if len(outcomeAncestors) > 0 {
		sql := fmt.Sprintf("delete from %s  where outcome_ancestor in (?)", entity.MilestoneOutcome{}.TableName())
		err := tx.Exec(sql, outcomeAncestors).Error
		if err != nil {
			log.Error(ctx, "UnbindOutcomes: exec sql failed",
				log.Err(err),
				log.Strings("delete", outcomeAncestors),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

type MilestoneCondition struct {
	ID          sql.NullString
	IDs         dbo.NullStrings
	AncestorID  sql.NullString
	AncestorIDs dbo.NullStrings
	SourceID    sql.NullString
	Description sql.NullString
	Name        sql.NullString
	Shortcode   sql.NullString
	Shortcodes  dbo.NullStrings
	SearchKey   sql.NullString

	AuthorID  sql.NullString
	AuthorIDs dbo.NullStrings

	Status   sql.NullString
	Statuses dbo.NullStrings

	Type sql.NullString

	OrganizationID sql.NullString
	IncludeDeleted bool
	OrderBy        MilestoneOrderBy `json:"order_by"`
	Pager          dbo.Pager
}

func (c *MilestoneCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.SearchKey.Valid {
		wheres = append(wheres, "match(name, shortcode, `describe`) against(? in boolean mode)")
		params = append(params, strings.TrimSpace(c.SearchKey.String))
	}

	if c.Name.Valid {
		wheres = append(wheres, "match(name) against(? in boolean mode)")
		params = append(params, strings.TrimSpace(c.Name.String))
	}

	if c.Shortcode.Valid {
		wheres = append(wheres, "match(shortcode) against(? in boolean mode)")
		params = append(params, strings.TrimSpace(c.Shortcode.String))
	}

	if c.Description.Valid {
		wheres = append(wheres, "match(`describe`) against(? in boolean mode)")
		params = append(params, strings.TrimSpace(c.Description.String))
	}

	if c.ID.Valid {
		wheres = append(wheres, "id = ?")
		params = append(params, c.ID.String)
	}

	if c.IDs.Valid {
		wheres = append(wheres, "id in (?)")
		params = append(params, c.IDs.Strings)
	}

	if c.Shortcodes.Valid {
		wheres = append(wheres, "shortcode in (?)")
		params = append(params, c.Shortcodes.Strings)
	}

	if c.AncestorID.Valid {
		wheres = append(wheres, "ancestor_id = ?")
		params = append(params, c.AncestorID.String)
	}

	if c.AncestorIDs.Valid {
		wheres = append(wheres, "ancestor_id in (?)")
		params = append(params, c.AncestorIDs.Strings)
	}

	if c.SourceID.Valid {
		wheres = append(wheres, "source_id = ?")
		params = append(params, c.SourceID.String)
	}

	if c.AuthorID.Valid {
		wheres = append(wheres, "author_id = ?")
		params = append(params, c.AuthorID.String)
	}

	if c.AuthorIDs.Valid {
		wheres = append(wheres, "author_id in (?)")
		params = append(params, c.AuthorIDs.Strings)
	}

	if c.OrganizationID.Valid {
		wheres = append(wheres, "organization_id=?")
		params = append(params, c.OrganizationID.String)
	}

	if c.Status.Valid {
		wheres = append(wheres, "status=?")
		params = append(params, c.Status.String)
	}

	if c.Statuses.Valid {
		wheres = append(wheres, "status in (?)")
		params = append(params, c.Statuses.Strings)
	}

	if c.Type.Valid {
		wheres = append(wheres, "type=?")
		params = append(params, c.Type.String)
	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "delete_at=0")
	}
	return wheres, params
}

type MilestoneOrderBy int

const (
	_ MilestoneOrderBy = iota
	OrderByMilestoneName
	OrderByMilestoneNameDesc
	OrderByMilestoneCreatedAt
	OrderByMilestoneCreatedAtDesc
	OrderByMilestoneUpdatedAt
	OrderByMilestoneUpdatedAtDesc
	OrderByMilestoneShortcode
)

func (c *MilestoneCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *MilestoneCondition) GetOrderBy() string {
	switch c.OrderBy {
	case OrderByMilestoneName:
		return "type, name"
	case OrderByMilestoneNameDesc:
		return "type, name desc"
	case OrderByMilestoneCreatedAt:
		return "type, create_at"
	case OrderByMilestoneCreatedAtDesc:
		return "type, create_at desc"
	case OrderByMilestoneUpdatedAt:
		return "type, update_at"
	case OrderByMilestoneUpdatedAtDesc:
		return "type, update_at desc"
	case OrderByMilestoneShortcode:
		return "shortcode"
	default:
		return "type desc, update_at desc"
	}
}

func NewMilestoneOrderBy(name string) MilestoneOrderBy {
	switch name {
	case "name":
		return OrderByMilestoneName
	case "-name":
		return OrderByMilestoneNameDesc
	case "created_at":
		return OrderByMilestoneCreatedAt
	case "-created_at":
		return OrderByMilestoneCreatedAtDesc
	case "updated_at":
		return OrderByMilestoneUpdatedAt
	case "-updated_at":
		return OrderByMilestoneUpdatedAtDesc
	case "shortcode":
		return OrderByMilestoneShortcode
	default:
		return OrderByMilestoneUpdatedAtDesc
	}
}

// ------------------------MilestoneOutcome------------------------------

type MilestoneOutcomeSQLDA struct {
	dbo.BaseDA
}

type MilestoneOutcomeCondition struct {
	MilestoneID      sql.NullString
	MilestoneIDs     dbo.NullStrings
	NotMilestoneID   sql.NullString
	OutcomeAncestor  sql.NullString
	OutcomeAncestors dbo.NullStrings
	IncludeDeleted   bool
	OrderBy          MilestoneOutcomeOrderBy `json:"order_by"`
	Pager            dbo.Pager
}
type MilestoneOutcomeOrderBy int

const (
	_ MilestoneOutcomeOrderBy = iota
	OrderByMilestoneOutcomeUpdatedAt
	OrderByMilestoneOutcomeUpdatedAtDesc
)

func (c *MilestoneOutcomeCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.MilestoneID.Valid {
		wheres = append(wheres, "milestone_id = ?")
		params = append(params, c.MilestoneID.String)
	}

	if c.MilestoneIDs.Valid {
		wheres = append(wheres, "milestone_id in (?)")
		params = append(params, c.MilestoneIDs.Strings)
	}

	if c.NotMilestoneID.Valid {
		wheres = append(wheres, "milestone_id <> ?")
		params = append(params, c.NotMilestoneID.String)
	}

	if c.OutcomeAncestor.Valid {
		wheres = append(wheres, "outcome_ancestor = ?")
		params = append(params, c.OutcomeAncestor.String)
	}

	if c.OutcomeAncestors.Valid {
		wheres = append(wheres, "outcome_ancestor in (?)")
		params = append(params, c.OutcomeAncestors.Strings)
	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "delete_at is null")
	}
	return wheres, params
}

func (c *MilestoneOutcomeCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *MilestoneOutcomeCondition) GetOrderBy() string {
	switch c.OrderBy {
	case OrderByMilestoneOutcomeUpdatedAt:
		return "update_at"
	case OrderByMilestoneOutcomeUpdatedAtDesc:
		return "update_at desc"
	default:
		return "update_at desc"

	}
}

func (mso MilestoneOutcomeSQLDA) SearchTx(ctx context.Context, tx *dbo.DBContext, condition *MilestoneOutcomeCondition) ([]*entity.MilestoneOutcome, error) {
	var result []*entity.MilestoneOutcome
	_, err := mso.BaseDA.PageTx(ctx, tx, condition, &result)
	if err != nil {
		log.Error(ctx, "SearchTx: PageTx failed",
			log.Err(err),
			log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

func (mso MilestoneOutcomeSQLDA) DeleteTx(ctx context.Context, tx *dbo.DBContext, milestoneIDs []string) error {
	table := entity.MilestoneOutcome{}.TableName()
	if len(milestoneIDs) > 0 {
		sql := fmt.Sprintf("delete from %s where milestone_id in (?)", table)
		err := tx.Exec(sql, milestoneIDs).Error
		if err != nil {
			log.Error(ctx, "DeleteTx: exec del sql failed",
				log.Err(err),
				log.Strings("milestone", milestoneIDs),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}
func (mso MilestoneOutcomeSQLDA) InsertTx(ctx context.Context, tx *dbo.DBContext, milestonesOutcomes []*entity.MilestoneOutcome) error {
	table := entity.MilestoneOutcome{}.TableName()
	if len(milestonesOutcomes) > 0 {
		values := make([][]interface{}, len(milestonesOutcomes))
		now := time.Now().Unix()
		for i := range milestonesOutcomes {
			tm := now + int64(i)
			values[i] = []interface{}{
				milestonesOutcomes[i].MilestoneID,
				milestonesOutcomes[i].OutcomeAncestor,
				tm,
				tm,
			}
		}
		sql, results := SQLBatchInsert(table, []string{"milestone_id", "outcome_ancestor", "create_at", "update_at"}, values)
		err := tx.Exec(sql, results...).Error
		if err != nil {
			log.Error(ctx, "InsertTx: exec insert sql failed",
				log.Err(err),
				log.Any("milestonesOutcomes", milestonesOutcomes),
				log.String("sql", sql))
			return err
		}
	}
	return nil
}

func (mso MilestoneOutcomeSQLDA) CountTx(ctx context.Context, tx *dbo.DBContext, generalIDs, normalIDs []string) (map[string]int, error) {
	var generalResults, normalResults []*struct {
		MilestoneID string `gorm:"column:milestone_id" json:"milestone_id"`
		Count       int    `gorm:"column:count" json:"count"`
	}

	generalSql := fmt.Sprintf("select gmo.milestone_id, count(gmo.outcome_ancestor) as count from "+
		"(select any_value(milestone_id) as milestone_id, outcome_ancestor from %s where outcome_ancestor in "+
		"(select outcome_ancestor from %s where milestone_id in (?)) "+
		"group by outcome_ancestor having count(milestone_id) = 1) as gmo group by gmo.milestone_id", entity.MilestoneOutcome{}.TableName(), entity.MilestoneOutcome{}.TableName())

	err := tx.Raw(generalSql, generalIDs).Find(&generalResults).Error
	if err != nil {
		log.Error(ctx, "Count: exec sql failed",
			log.Strings("general", normalIDs),
			log.String("sql", generalSql))
		return nil, err
	}

	normalSql := fmt.Sprintf("select milestone_id, count(outcome_ancestor) as count from %s where milestone_id in (?) group by milestone_id", entity.MilestoneOutcome{}.TableName())
	err = tx.Raw(normalSql, normalIDs).Find(&normalResults).Error
	if err != nil {
		log.Error(ctx, "Count: exec sql failed",
			log.Strings("normal", normalIDs),
			log.String("sql", normalSql))
		return nil, err
	}

	counts := make(map[string]int)
	for i := range generalResults {
		counts[generalResults[i].MilestoneID] = generalResults[i].Count
	}
	for i := range normalResults {
		counts[normalResults[i].MilestoneID] = normalResults[i].Count
	}
	return counts, nil
}
