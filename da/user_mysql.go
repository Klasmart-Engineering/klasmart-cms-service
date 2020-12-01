package da

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type UserSqlDA struct {
	dbo.BaseDA
}

type UserCondition struct {
	UserIDs      dbo.NullStrings
	UserName     sql.NullString
	FuzzyKey     sql.NullString
	FuzzyAuthors dbo.NullStrings

	IncludeDeleted bool
	OrderBy        OutcomeOrderBy `json:"order_by"`
	Pager          dbo.Pager
}

func (c *UserCondition) GetUserConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	if c.FuzzyKey.Valid && !c.FuzzyAuthors.Valid {
		wheres = append(wheres, "match(name, keywords, description, shortcode) against(? in boolean mode)")
		params = append(params, strings.TrimSpace(c.FuzzyKey.String))
	}

	if c.FuzzyKey.Valid && c.FuzzyAuthors.Valid {
		wheres = append(wheres, fmt.Sprintf("((match(name, keywords, description, shortcode) against(? in boolean mode)) or (id in (%s)))", c.FuzzyAuthors.SQLPlaceHolder()))
		params = append(params, strings.TrimSpace(c.FuzzyKey.String), c.FuzzyAuthors.ToInterfaceSlice())

	}

	if !c.IncludeDeleted {
		wheres = append(wheres, "delete_at=0")
	}
	return wheres, params
}

type UserOrderBy int

const (
	_ = iota
	UserByName
)

const defaultUserPageIndex = 1
const defaultUserPageSize = 20

func NewUserPage(page, pageSize int) dbo.Pager {
	if page == -1 {
		return dbo.NoPager
	}
	if page == 0 {
		page = defaultUserPageIndex
	}
	if pageSize == 0 {
		pageSize = defaultUserPageSize
	}
	return dbo.Pager{
		Page:     page,
		PageSize: pageSize,
	}
}

func (c *UserCondition) GetUserPager() *dbo.Pager {
	return &c.Pager
}

func NewUserOrderBy(name string) UserOrderBy {
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

func (c *UserCondition) GetUserOrderBy() string {
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

func (u *UserSqlDA) GetUserByAccount(ctx context.Context, tx *dbo.DBContext, account string) (*entity.User, error) {
	var user entity.User
	sql := fmt.Sprintf("select * from %s where phone = %s or email = %s", user.TableName(), account, account)
	err := tx.Raw(sql).Find(&user).Error

	if gorm.IsRecordNotFoundError(err) {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetUserByAccount: Find failed", log.String("account", account), log.Err(err))
		return nil, err
	}

	return &user, nil
}

func (u *UserSqlDA) InsertTx(ctx context.Context, db *dbo.DBContext, value *entity.User) (*entity.User, error) {
	now := time.Now().Unix()
	value.CreateAt, value.UpdateAt = now, now
	res, err := u.BaseDA.InsertTx(ctx, db, value)
	if err != nil {
		log.Error(ctx, "InsertTx: failed", log.Err(err))
	}
	if err == dbo.ErrDuplicateRecord {
		return nil, constant.ErrDuplicateRecord
	}
	return res.(*entity.User), err
}
