package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IUserSettingModel interface {
	SetByOperator(ctx context.Context, op *entity.Operator, jsonData *entity.UserSettingJsonContent) (string, error)
	GetByOperator(ctx context.Context, op *entity.Operator) (*entity.UserSettingJsonContent, error)
}

type userSettingModel struct {
}

func (u *userSettingModel) SetByOperator(ctx context.Context, op *entity.Operator, jsonData *entity.UserSettingJsonContent) (string, error) {
	var userSettings []*entity.UserSetting
	err := da.GetUserSettingDA().Query(ctx, da.UserSettingCondition{
		UserID: sql.NullString{
			String: op.UserID,
			Valid:  true,
		}}, &userSettings)
	if err != nil {
		log.Error(ctx, "SetByUserID:GetUserSettingDA.Query error",
			log.Err(err),
			log.Any("jsonData", jsonData),
			log.Any("op", op),
		)
		return "", err
	}
	b, err := json.Marshal(jsonData)
	if err != nil {
		log.Error(ctx, "SetByUserID:json.Marshal error",
			log.Err(err),
			log.Any("jsonData", jsonData),
			log.Any("op", op),
		)
		return "", err
	}
	settingJson := string(b)
	if len(userSettings) <= 0 {
		// add
		data := new(entity.UserSetting)
		data.ID = utils.NewID()
		data.UserID = op.UserID
		data.SettingJson = settingJson
		_, err := da.GetUserSettingDA().Insert(ctx, data)
		if err != nil {
			log.Error(ctx, "SetByUserID:GetUserSettingDA.Insert error",
				log.Err(err),
				log.Any("new", data),
				log.Any("op", op),
			)
			return "", err
		}
		return data.ID, nil
	}
	// update
	old := userSettings[0]
	old.SettingJson = settingJson
	_, err = da.GetUserSettingDA().Update(ctx, old)
	if err != nil {
		log.Error(ctx, "SetByUserID:GetUserSettingDA.Update error",
			log.Err(err),
			log.Any("old", old),
			log.Any("jsonData", jsonData),
			log.Any("op", op),
		)
		return "", err
	}
	return old.ID, nil
}

func (u *userSettingModel) GetByOperator(ctx context.Context, op *entity.Operator) (*entity.UserSettingJsonContent, error) {
	// get default setting
	var defaultSettings []*entity.UserSetting
	err := da.GetUserSettingDA().Query(ctx, da.UserSettingCondition{
		UserID: sql.NullString{
			String: entity.DefaultSettingID,
			Valid:  true,
		}}, &defaultSettings)
	if err != nil {
		log.Error(ctx, "GetByOperator:GetUserSettingDA.Query error",
			log.Err(err),
			log.Any("op", op),
		)
		return nil, err
	}
	if len(defaultSettings) <= 0 {
		return nil, constant.ErrRecordNotFound
	}
	defaultSetting := defaultSettings[0]
	defaultJsonContent := new(entity.UserSettingJsonContent)
	err = json.Unmarshal([]byte(defaultSetting.SettingJson), defaultJsonContent)
	if err != nil {
		log.Error(ctx, "GetByOperator:json.Unmarshal error",
			log.Err(err),
			log.Any("op", op),
			log.Any("defaultSetting", defaultSetting),
		)
		return nil, err
	}

	// get user setting
	var userSettings []*entity.UserSetting
	err = da.GetUserSettingDA().Query(ctx, da.UserSettingCondition{
		UserID: sql.NullString{
			String: op.UserID,
			Valid:  true,
		}}, &userSettings)
	if err != nil {
		log.Error(ctx, "GetByOperator:GetUserSettingDA.Query error",
			log.Err(err),
			log.Any("op", op),
		)
		return nil, err
	}
	if len(userSettings) <= 0 {
		return defaultJsonContent, nil
	}
	userSetting := userSettings[0]
	userJsonContent := new(entity.UserSettingJsonContent)
	err = json.Unmarshal([]byte(userSetting.SettingJson), userJsonContent)
	if err != nil {
		log.Error(ctx, "GetByOperator:json.Unmarshal error",
			log.Err(err),
			log.Any("op", op),
			log.Any("userSetting", userSetting),
		)
		return nil, err
	}
	err = utils.JsonMerge(userJsonContent, defaultJsonContent)
	if err != nil {
		log.Error(ctx, "GetByOperator:utils.JsonMerge error",
			log.Err(err),
			log.Any("op", op),
			log.Any("userJsonContent", userJsonContent),
			log.Any("defaultJsonContent", defaultJsonContent),
		)
		return nil, err
	}
	return userJsonContent, nil
}

var (
	_userSettingOnce  sync.Once
	_userSettingModel IUserSettingModel
)

func GetUserSettingModel() IUserSettingModel {
	_userSettingOnce.Do(func() {
		_userSettingModel = &userSettingModel{}
	})
	return _userSettingModel
}
