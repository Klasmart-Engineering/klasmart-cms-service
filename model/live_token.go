package model

import (
	"context"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/contentdata"

	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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
	if schedule.Status == entity.ScheduleStatusNotStart {
		err := GetScheduleModel().UpdateScheduleStatus(ctx, dbo.MustGetDB(ctx), schedule.ID, entity.ScheduleStatusStarted)
		if err != nil {
			return "", err
		}
	}
	liveTokenInfo := entity.LiveTokenInfo{
		UserID: op.UserID,
		Type:   entity.LiveTokenTypeLive,
		RoomID: scheduleID,
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
	liveTokenInfo.Teacher = true //op.Role == entity.RoleTeacher

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
		RoomID: utils.NewID(),
	}

	name, err := s.getUserName(ctx, op)
	if err != nil {
		log.Error(ctx, "MakeLivePreviewToken:get user name by id error",
			log.Err(err),
			log.Any("op", op))
		return "", err
	}
	liveTokenInfo.Name = name
	liveTokenInfo.Teacher = true //op.Role == entity.RoleTeacher

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
		teacherService := external.GetTeacherServiceProvider()
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
		ExpiresAt: now.Add(constant.LiveTokenExpiresAt).Unix(),
		IssuedAt:  now.Add(-constant.LiveTokenIssuedAt).Unix(),
		Issuer:    "KidsLoopUser-live",
		NotBefore: 0,
		Subject:   "authorization",
	}

	claims := &entity.LiveTokenClaims{
		StandardClaims: stdClaims,
		LiveTokenInfo:  liveTokenInfo,
	}
	token, err := utils.CreateJWT(ctx, claims, config.Get().LiveTokenConfig.PrivateKey)
	if err != nil {
		log.Error(ctx, "MakeLiveToken:create jwt error",
			log.Err(err),
			log.Any("liveTokenInfo", liveTokenInfo))
		return "", err
	}
	return token, nil
}

func (s *liveTokenModel) getMaterials(ctx context.Context, contentID string) ([]*entity.LiveMaterial, error) {
	contentList, err := GetContentModel().GetContentSubContentsByID(ctx, dbo.MustGetDB(ctx), contentID)
	log.Debug(ctx, "content data", log.Any("contentList", contentList))
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "getMaterials:get content sub by id not found",
			log.Err(err),
			log.String("contentID", contentID))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "getMaterials:get content sub by id error",
			log.Err(err),
			log.String("contentID", contentID))
		return nil, err
	}
	materials := make([]*entity.LiveMaterial, 0, len(contentList))
	for _, item := range contentList {
		if item == nil {
			continue
		}
		materialItem := &entity.LiveMaterial{
			Name: item.Name,
		}
		mData, ok := item.Data.(*contentdata.MaterialData)
		if !ok {
			log.Debug(ctx, "content data convert materialdata error", log.Any("item", item))
			continue
		}
		// material type
		switch mData.FileType {
		case entity.FileTypeImage:
			materialItem.TypeName = entity.MaterialTypeImage
		case entity.FileTypeAudio:
			materialItem.TypeName = entity.MaterialTypeAudio
		case entity.FileTypeVideo:
			materialItem.TypeName = entity.MaterialTypeVideo
		case entity.FileTypeH5p:
			materialItem.TypeName = entity.MaterialTypeH5P
		default:
			log.Warn(ctx, "content material type is invalid", log.Any("materialData", mData))
			continue
		}
		// material url
		if materialItem.TypeName == entity.MaterialTypeH5P {
			materialItem.URL = fmt.Sprintf("/h5p/play/%v", mData.Source)
		} else {
			materialItem.URL, err = GetResourceUploaderModel().GetResourcePath(ctx, string(mData.Source))
			if err != nil {
				log.Error(ctx, "getMaterials:get resource path error",
					log.Err(err),
					log.String("contentID", contentID),
					log.Any("mData", mData))
				return nil, err
			}
		}

		materials = append(materials, materialItem)
	}
	return materials, nil
}

type liveTokenModel struct {
	PrivateKey interface{}
}

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
