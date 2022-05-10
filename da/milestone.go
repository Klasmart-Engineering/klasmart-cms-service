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
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IMilestoneDA interface {
	dbo.DataAccesser

	// return the min shortcode in each idle interval, excluding deleted
	GetIdleShortcode(ctx context.Context, tx *dbo.DBContext, orgID string) ([]int, error)
	GetLargerShortcode(ctx context.Context, tx *dbo.DBContext, orgID string, shortcodeNum, size int) ([]int, error)

	// TODO: remove duplicate interfaces
	Create(ctx context.Context, tx *dbo.DBContext, milestone *entity.Milestone) error
	UpdateMilestone(ctx context.Context, tx *dbo.DBContext, milestone *entity.Milestone) error
	Search(ctx context.Context, tx *dbo.DBContext, condition *MilestoneCondition) (int, []*entity.Milestone, error)
	GetByID(ctx context.Context, tx *dbo.DBContext, ID string) (*entity.Milestone, error)
	UpdateLatest(ctx context.Context, tx *dbo.DBContext, ancestorID, latestID string) error

	BatchPublish(ctx context.Context, tx *dbo.DBContext, publishIDs []string) error
	BatchHide(ctx context.Context, tx *dbo.DBContext, hideIDs []string) error
	BatchUnLock(ctx context.Context, tx *dbo.DBContext, unLockIDs []string) error
	BatchUpdateLatest(ctx context.Context, tx *dbo.DBContext, ancestorLatest map[string]string) error
	BatchDelete(ctx context.Context, tx *dbo.DBContext, milestoneIDs []string) error

	UnbindOutcomes(ctx context.Context, tx *dbo.DBContext, outcomeAncestors []string) error
}

type milestoneDA struct {
	dbo.BaseDA
}

var _milestoneDA IMilestoneDA
var _milestoneOnce sync.Once

func GetMilestoneDA() IMilestoneDA {
	_milestoneOnce.Do(func() {
		_milestoneDA = new(milestoneDA)
	})
	return _milestoneDA
}

func (m milestoneDA) GetLargerShortcode(ctx context.Context, tx *dbo.DBContext, orgID string, shortcodeNum, size int) ([]int, error) {
	tx.ResetCondition()

	var result []int
	err := tx.Table(entity.MilestoneTable).
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

func (m milestoneDA) GetIdleShortcode(ctx context.Context, tx *dbo.DBContext, orgID string) ([]int, error) {
	tx.ResetCondition()

	var result []int
	sql := `
	SELECT DISTINCT
		a.shortcode_num 
	FROM
		milestones AS a
		LEFT JOIN milestones AS b ON b.shortcode_num = a.shortcode_num + 1 
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
		log.Error(ctx, "GetIdleShortcode error",
			log.Err(err),
			log.String("orgID", orgID),
			log.String("sql", sql))
		return result, err
	}

	for i := range result {
		result[i] = result[i] + 1
	}

	return result, nil
}

func (m milestoneDA) Create(ctx context.Context, tx *dbo.DBContext, milestone *entity.Milestone) error {
	_, err := m.InsertTx(ctx, tx, milestone)
	if err != nil {
		log.Error(ctx, "Create: InsertTx failed",
			log.Err(err),
			log.Any("milestone", milestone))
		return err
	}
	return nil
}

func (m milestoneDA) GetByID(ctx context.Context, tx *dbo.DBContext, ID string) (*entity.Milestone, error) {
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

func (m milestoneDA) UpdateMilestone(ctx context.Context, tx *dbo.DBContext, milestone *entity.Milestone) error {
	milestone.UpdateAt = time.Now().Unix()
	_, err := m.BaseDA.UpdateTx(ctx, tx, milestone)
	if err != nil {
		log.Error(ctx, "Update: UpdateTx failed",
			log.Any("milestone", milestone))
		return err
	}
	return nil
}

func (m milestoneDA) UpdateLatest(ctx context.Context, tx *dbo.DBContext, ancestorID, latestID string) error {
	sql := fmt.Sprintf("UPDATE %s SET latest_id = ? WHERE ancestor_id = ? AND delete_at = 0", entity.Milestone{}.TableName())
	err := tx.Exec(sql, latestID, ancestorID).Error
	if err != nil {
		log.Error(ctx, "UpdateLatest failed",
			log.Err(err),
			log.String("ancestorID", ancestorID),
			log.String("latestID", latestID),
			log.String("sql", sql))
		return err
	}

	return nil
}

func (m milestoneDA) Search(ctx context.Context, tx *dbo.DBContext, condition *MilestoneCondition) (int, []*entity.Milestone, error) {
	var milestones []*entity.Milestone
	total, err := m.BaseDA.PageTx(ctx, tx, condition, &milestones)
	if err != nil {
		log.Error(ctx, "Search: PageTx failed",
			log.Any("condition", condition))
		return 0, nil, err
	}
	return total, milestones, nil
}

func (m milestoneDA) BatchPublish(ctx context.Context, tx *dbo.DBContext, publishIDs []string) error {
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

func (m milestoneDA) BatchHide(ctx context.Context, tx *dbo.DBContext, hideIDs []string) error {
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

func (m milestoneDA) BatchUnLock(ctx context.Context, tx *dbo.DBContext, unLockIDs []string) error {
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

func (m milestoneDA) BatchUpdateLatest(ctx context.Context, tx *dbo.DBContext, ancestorLatest map[string]string) error {
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

func (m milestoneDA) BatchDelete(ctx context.Context, tx *dbo.DBContext, milestoneIDs []string) error {
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

func (m milestoneDA) UnbindOutcomes(ctx context.Context, tx *dbo.DBContext, outcomeAncestors []string) error {
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
	SourceIDs   dbo.NullStrings
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

	if c.SourceIDs.Valid {
		wheres = append(wheres, "source_id in (?)")
		params = append(params, c.SourceIDs.Strings)
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
