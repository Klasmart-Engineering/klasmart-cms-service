package da

import (
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type IUserSettingDA interface {
	dbo.DataAccesser
}

type userSettingDA struct {
	dbo.BaseDA
}

var (
	_userSettingOnce sync.Once
	_userSettingDA   IUserSettingDA
)

func GetUserSettingDA() IUserSettingDA {
	_userSettingOnce.Do(func() {
		_userSettingDA = &userSettingDA{}
	})
	return _userSettingDA
}

type UserSettingCondition struct {
	UserID sql.NullString
}

func (c UserSettingCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.UserID.Valid {
		wheres = append(wheres, "user_id = ?")
		params = append(params, c.UserID.String)
	}

	return wheres, params
}

func (c UserSettingCondition) GetOrderBy() string {
	return ""
}

func (c UserSettingCondition) GetPager() *dbo.Pager {
	return nil
}
