package model

import (
	"encoding/json"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
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
