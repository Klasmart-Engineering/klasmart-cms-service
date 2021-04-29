package da

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
)

type ShortcodeMySqlDA struct {
	dbo.BaseDA
}

type ShortcodeCondition struct {
	OrgID          sql.NullString
	NotAncestorID  sql.NullString
	Shortcode      sql.NullString
	IncludeDeleted bool
}

func (c *ShortcodeCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.OrgID.Valid {
		wheres = append(wheres, "organization_id = ?")
		params = append(params, c.OrgID.String)
	}

	if c.Shortcode.Valid {
		wheres = append(wheres, "shortcode = ?")
		params = append(params, c.Shortcode.String)
	}

	if c.NotAncestorID.Valid {
		wheres = append(wheres, "ancestor_id <> ?")
		params = append(params, c.NotAncestorID.String)
	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "delete_at=0")
	}
	return wheres, params
}

func (scd ShortcodeMySqlDA) Search(ctx context.Context, tx *dbo.DBContext, kind IShortcodeKind, condition *ShortcodeCondition) ([]*entity.ShortcodeElement, error) {
	wheres, parameters := condition.GetConditions()
	db := tx.Clone()
	if len(wheres) > 0 {
		db.DB = db.Where(strings.Join(wheres, " and "), parameters...)
	}
	var table string
	switch kind.Kind() {
	case entity.KindOutcome:
		table = entity.Outcome{}.TableName()
	case entity.KindMileStone:
		table = entity.Milestone{}.TableName()
	}
	var shortcodes []*entity.ShortcodeElement
	err := db.Table(table).Select("shortcode").Find(&shortcodes).Error
	if err != nil {
		log.Error(ctx, "Search: failed",
			log.Err(err),
			log.String("kind", string(kind.Kind())),
			log.Any("condition", condition))
		return nil, err
	}
	return shortcodes, nil
}
