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
	MakeLivePreviewToken(ctx context.Context, op *entity.Operator, contentID string) (string, error)
}

func (s *liveTokenModel) MakeLiveToken(ctx context.Context, op *entity.Operator, scheduleID string) (string, error) {
	schedule, err := GetScheduleModel().GetPlainByID(ctx, scheduleID)
	if err != nil {
		return "", err
	}

	liveTokenInfo := entity.LiveTokenInfo{
		UserID: op.UserID,
		Type:   entity.LiveTokenTypeLive,
	}
	liveTokenInfo.ScheduleID = schedule.ID

	name, err := s.getUserName(ctx, op)
	if err != nil {
		log.Error(ctx, "MakeLiveToken:get user name by id error",
			log.Err(err),
			log.Any("op", op),
			log.String("scheduleID", scheduleID))
		return "", err
	}
	liveTokenInfo.Name = name

	liveTokenInfo.Materials, err = s.getMaterials(ctx, schedule.LessonPlanID)
	if err != nil {
		log.Error(ctx, "MakeLiveToken:get material error",
			log.Err(err),
			log.Any("op", op),
			log.Any("liveTokenInfo", liveTokenInfo),
			log.Any("schedule", schedule))
		return "", err
	}

	token, err := s.createJWT(ctx, liveTokenInfo)
	if err != nil {
		log.Error(ctx, "MakeLiveToken:create jwt error",
			log.Err(err),
			log.Any("op", op),
			log.Any("liveTokenInfo", liveTokenInfo),
			log.Any("schedule", schedule))
		return "", err
	}
	return token, nil
}
func (s *liveTokenModel) MakeLivePreviewToken(ctx context.Context, op *entity.Operator, contentID string) (string, error) {
	liveTokenInfo := entity.LiveTokenInfo{
		UserID: op.UserID,
		Type:   entity.LiveTokenTypePreview,
	}

	name, err := s.getUserName(ctx, op)
	if err != nil {
		log.Error(ctx, "MakeLivePreviewToken:get user name by id error",
			log.Err(err),
			log.Any("op", op))
		return "", err
	}
	liveTokenInfo.Name = name

	liveTokenInfo.Materials, err = s.getMaterials(ctx, contentID)
	if err != nil {
		log.Error(ctx, "MakeLivePreviewToken:get material error",
			log.Err(err),
			log.Any("op", op),
			log.Any("liveTokenInfo", liveTokenInfo),
			log.String("contentID", contentID))
		return "", err
	}

	token, err := s.createJWT(ctx, liveTokenInfo)
	if err != nil {
		log.Error(ctx, "MakeLivePreviewToken:create jwt error",
			log.Err(err),
			log.Any("op", op),
			log.Any("liveTokenInfo", liveTokenInfo),
			log.String("contentID", contentID))
		return "", err
	}
	return token, nil
}

func (s *liveTokenModel) getUserName(ctx context.Context, op *entity.Operator) (string, error) {
	switch op.Role {
	case entity.RoleTeacher:
		teacherService, err := external.GetTeacherServiceProvider()
		if err != nil {
			log.Error(ctx, "getUserName:GetTeacherServiceProvider error",
				log.Err(err),
				log.Any("op", op))
			return "", err
		}
		teacherInfos, err := teacherService.BatchGet(ctx, []string{op.UserID})
		if err != nil {
			log.Error(ctx, "getUserName:GetTeacherServiceProvider BatchGet error",
				log.Err(err),
				log.Any("op", op))
			return "", err
		}
		if len(teacherInfos) <= 0 {
			log.Error(ctx, "getUserName:teacher info not found",
				log.Err(err),
				log.Any("op", op))
			return "", constant.ErrRecordNotFound
		}
		return teacherInfos[0].Name, nil
	case entity.RoleStudent:
		return entity.RoleStudent, nil
	case entity.RoleAdmin:
		return entity.RoleAdmin, nil
	default:
		log.Error(ctx, "getUserName:user role invalid", log.Any("op", op))
		return "", constant.ErrRecordNotFound
	}
}

func (s *liveTokenModel) createJWT(ctx context.Context, liveTokenInfo entity.LiveTokenInfo) (string, error) {
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
			log.Any("liveTokenInfo", liveTokenInfo))
		return "", err
	}
	return token, nil
}

func (s *liveTokenModel) getSubContent() []*entity.LiveTokenShort {
	return []*entity.LiveTokenShort{
		&entity.LiveTokenShort{
			ID:   "001",
			Name: "h5p",
		},
	}
}

func (s *liveTokenModel) getMaterials(ctx context.Context, contentID string) ([]*entity.LiveMaterial, error) {
	contentList := s.getSubContent()
	materials := make([]*entity.LiveMaterial, len(contentList))
	for i, item := range contentList {
		materialItem := &entity.LiveMaterial{
			Name:     item.Name,
			TypeName: entity.MaterialTypeH5P,
		}
		materialItem.URL = fmt.Sprintf("/%v/h5p-www/play/%v",
			entity.LiveTokenEnvPath,
			item.ID,
		)
		materials[i] = materialItem
	}
	return materials, nil
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
