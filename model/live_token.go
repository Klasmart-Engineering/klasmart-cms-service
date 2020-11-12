package model

import (
	"context"
	"fmt"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/contentdata"

	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	isTeacher, err := s.isTeacher(ctx, op)
	if err != nil {
		log.Error(ctx, "MakeLivePreviewToken:judge is teacher error",
			log.Err(err),
			log.Any("op", op))
		return "", err
	}
	liveTokenInfo.Teacher = isTeacher
	if schedule.ClassType == entity.ScheduleClassTypeTask {
		liveTokenInfo.Materials = make([]*entity.LiveMaterial, 0)
	} else {
		liveTokenInfo.Materials, err = s.getMaterials(ctx, schedule.LessonPlanID)
		if err != nil {
			log.Error(ctx, "MakeLiveToken:get material error",
				log.Err(err),
				log.Any("op", op),
				log.Any("liveTokenInfo", liveTokenInfo),
				log.Any("schedule", schedule))
			return "", err
		}
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
	isTeacher, err := s.isTeacher(ctx, op)
	if err != nil {
		log.Error(ctx, "MakeLivePreviewToken:judge is teacher error",
			log.Err(err),
			log.Any("op", op))
		return "", err
	}
	liveTokenInfo.Teacher = isTeacher

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
	userInfo, err := external.GetUserServiceProvider().Get(ctx, op.UserID)
	if err != nil {
		log.Error(ctx, "getUserName:get user name error",
			log.Err(err),
			log.Any("op", op),
		)
		return "", err
	}
	return userInfo.Name, nil
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

func (s *liveTokenModel) isTeacher(ctx context.Context, op *entity.Operator) (bool, error) {
	organization, err := external.GetOrganizationServiceProvider().GetByPermission(ctx, op, external.LiveClassTeacher)
	if err != nil {
		log.Error(ctx, "isTeacher:GetOrganizationServiceProvider.GetByPermission error",
			log.String("permission", string(external.LiveClassTeacher)),
			log.Any("operator", op),
			log.Err(err),
		)
		return false, err
	}
	school, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, external.LiveClassTeacher)
	if err != nil {
		log.Error(ctx, "isTeacher:GetSchoolServiceProvider.GetByPermission error",
			log.String("permission", string(external.LiveClassTeacher)),
			log.Any("operator", op),
			log.Err(err),
		)
		return false, err
	}
	log.Info(ctx, "isTeacher:GetSchoolServiceProvider.GetByPermission error",
		log.String("permission", string(external.LiveClassTeacher)),
		log.Any("operator", op),
		log.Any("organization", organization),
		log.Any("school", school),
		log.Err(err),
	)
	if len(organization) != 0 || len(school) != 0 {
		return true, nil
	}
	return false, nil
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
