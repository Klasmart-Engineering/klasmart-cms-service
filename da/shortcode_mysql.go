package da

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
)

type ShortcodeSqlDA struct {
	dbo.BaseDA
}

type ShortcodeCondition struct {
	OrgID         sql.NullString
	NotAncestorID sql.NullString
	Shortcode     sql.NullString
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

	return wheres, params
}

func (scd ShortcodeSqlDA) Search(ctx context.Context, tx *dbo.DBContext, table string, condition *ShortcodeCondition) ([]*entity.ShortcodeElement, error) {
	wheres, parameters := condition.GetConditions()
	db := tx.Clone()
	if len(wheres) > 0 {
		db.DB = db.Where(strings.Join(wheres, " and "), parameters...)
	}
	var shortcodes []*entity.ShortcodeElement
	err := db.Table(table).Select("shortcode").Find(&shortcodes).Error
	if err != nil {
		return nil, err
	}
	return shortcodes, nil
}
