package model

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

type ILiveTokenModel interface {
	MakeLiveToken(ctx context.Context, op *entity.Operator, scheduleID string) (string, error)
}

func (s *liveTokenModel) MakeLiveToken(ctx context.Context, op *entity.Operator, scheduleID string) (string, error) {
	liveTokenInfo := entity.LiveTokenInfo{
		UserID: op.UserID,
		Type:   entity.LiveTokenTypeLive,
	}
	schedule, err := GetScheduleModel().GetPlainByID(ctx, scheduleID)
	if err != nil {
		return "", err
	}
	liveTokenInfo.ScheduleID = schedule.ID
	type ContentTemp struct {
		Name       string
		MaterialID string
	}
	contentList := make([]ContentTemp, 0)
	liveTokenInfo.Materials = make([]*entity.LiveMaterial, len(contentList))
	for i, item := range contentList {
		materialItem := &entity.LiveMaterial{
			Name:     item.Name,
			TypeName: entity.MaterialTypeH5P,
		}
		materialItem.URL = fmt.Sprintf("/%v/h5p-www/play/%v",
			entity.LiveTokenEnvPath,
			item.MaterialID,
		)
		liveTokenInfo.Materials[i] = materialItem
	}
	switch op.Role {
	case entity.RoleTeacher:
		teacherService, err := external.GetTeacherServiceProvider()
		if err != nil {
			log.Error(ctx, "MakeLiveToken:GetTeacherServiceProvider error",
				log.Err(err),
				log.Any("op", op),
				log.String("scheduleID", scheduleID))
			return "", err
		}
		teacherInfos, err := teacherService.BatchGet(ctx, []string{op.UserID})
		if err != nil {
			log.Error(ctx, "MakeLiveToken:GetTeacherServiceProvider BatchGet error",
				log.Err(err),
				log.Any("op", op),
				log.String("scheduleID", scheduleID))
			return "", err
		}
		if len(teacherInfos) <= 0 {
			log.Error(ctx, "MakeLiveToken:teacher info not found",
				log.Err(err),
				log.Any("op", op),
				log.String("scheduleID", scheduleID))
			return "", constant.ErrRecordNotFound
		}
		liveTokenInfo.Name = teacherInfos[0].Name
		liveTokenInfo.Teacher = true
	case entity.RoleStudent:
		// TODO
	default:
		liveTokenInfo.Name = op.Role
	}
	now := time.Now()
	stdClaims := &jwt.StandardClaims{
		Audience:  "kidsloop-live",
		ExpiresAt: now.Add(time.Hour * 24 * entity.ValidDays).Unix(),
		IssuedAt:  now.Add(-30 * time.Second).Unix(),
		Issuer:    "KidsLoopUser-live",
		NotBefore: 0,
		Subject:   "authorization",
	}

	claims := &entity.LiveTokenClaims{
		StandardClaims: stdClaims,
		LiveTokenInfo:  liveTokenInfo,
	}
	token, err := utils.CreateJWT(claims)
	if err != nil {
		log.Error(ctx, "MakeLiveToken:create jwt error",
			log.Err(err),
			log.Any("op", op),
			log.String("scheduleID", scheduleID))
		return "", err
	}
	return token, nil
}
func (s *liveTokenModel) MakeLivePreviewToken(ctx context.Context, op *entity.Operator, contentID string) (string, error) {
	return "", nil
}

type liveTokenModel struct{}

var (
	_liveTokenOnce  sync.Once
	_liveTokenModel ILiveTokenModel
)

func GetLiveTokenModel() ILiveTokenModel {
	_liveTokenOnce.Do(func() {
		_liveTokenModel = &liveTokenModel{}
	})
	return _liveTokenModel
}
