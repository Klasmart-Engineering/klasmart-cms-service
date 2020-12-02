package model

import (
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func Test_GetUserSettingJson(t *testing.T) {

	data := &entity.UserSettingJsonContent{
		CMSPageSize: 20,
	}
	b, err := json.Marshal(data)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(string(b))
	setting := &entity.UserSetting{
		ID:          entity.DefaultUserSettingID,
		UserID:      entity.DefaultUserSettingID,
		SettingJson: string(b),
	}
	t.Log(*setting)
}
