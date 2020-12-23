package model

import (
	"context"
	"errors"
	"fmt"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrGoLiveTimeNotUp = errors.New("go live time not up")
	ErrGoLiveNotAllow  = errors.New("go live not allow")
)

type ILiveTokenModel interface {
	MakeLiveToken(ctx context.Context, op *entity.Operator, scheduleID string) (string, error)
	MakeLivePreviewToken(ctx context.Context, op *entity.Operator, contentID string, classID string) (string, error)
}

func (s *liveTokenModel) MakeLiveToken(ctx context.Context, op *entity.Operator, scheduleID string) (string, error) {
	schedule, err := GetScheduleModel().GetPlainByID(ctx, scheduleID)
	if err != nil {
		log.Error(ctx, "MakeLiveToken:GetScheduleModel.GetPlainByID error",
			log.Err(err),
			log.Any("op", op),
			log.String("scheduleID", scheduleID))
		return "", err
	}
	now := time.Now().Unix()
	diff := utils.TimeStampDiff(schedule.StartAt, now)
	if diff >= constant.ScheduleAllowGoLiveTime {
		log.Warn(ctx, "MakeLiveToken: go live time not up",
			log.Any("op", op),
			log.String("scheduleID", scheduleID),
			log.Int64("schedule.StartAt", schedule.StartAt),
			log.Int64("time.Now", now),
		)
		return "", ErrGoLiveTimeNotUp
	}
	if schedule.Status.GetScheduleStatus(schedule.EndAt) == entity.ScheduleStatusClosed {
		log.Warn(ctx, "MakeLiveToken:go live not allow",
			log.Any("op", op),
			log.Any("schedule", schedule),
			log.Int64("schedule.StartAt", schedule.StartAt),
			log.Int64("time.Now", now),
		)
		return "", ErrGoLiveNotAllow
	}
	classType := schedule.ClassType.ConvertToLiveClassType()
	if classType == entity.LiveClassTypeInvalid {
		log.Error(ctx, "MakeLiveToken:ConvertToLiveClassType invalid",
			log.Any("op", op),
			log.String("scheduleID", scheduleID),
			log.Any("schedule.ClassType", schedule.ClassType),
		)
		return "", constant.ErrInvalidArgs
	}
	liveTokenInfo := entity.LiveTokenInfo{
		UserID:    op.UserID,
		Type:      entity.LiveTokenTypeLive,
		RoomID:    scheduleID,
		ClassType: classType,
		OrgID:     op.OrgID,
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
	isTeacher, err := s.isTeacherByClass(ctx, op, schedule.ClassID)
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
func (s *liveTokenModel) MakeLivePreviewToken(ctx context.Context, op *entity.Operator, contentID string, classID string) (string, error) {
	liveTokenInfo := entity.LiveTokenInfo{
		UserID: op.UserID,
		Type:   entity.LiveTokenTypePreview,
		RoomID: contentID,
		OrgID:  op.OrgID,
	}

	name, err := s.getUserName(ctx, op)
	if err != nil {
		log.Error(ctx, "MakeLivePreviewToken:get user name by id error",
			log.Err(err),
			log.Any("op", op))
		return "", err
	}
	liveTokenInfo.Name = name
	var isTeacher bool
	if classID == "" {
		isTeacher, err = s.isTeacherByPermission(ctx, op)
		if err != nil {
			log.Error(ctx, "MakeLivePreviewToken:isTeacherByPermission error",
				log.Err(err),
				log.Any("op", op))
			return "", err
		}
	} else {
		isTeacher, err = s.isTeacherByClass(ctx, op, classID)
		if err != nil {
			log.Error(ctx, "MakeLivePreviewToken:isTeacherByClass error",
				log.Err(err),
				log.Any("op", op),
				log.String("classID", classID),
			)
			return "", err
		}
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
	userInfo, err := external.GetUserServiceProvider().Get(ctx, op, op.UserID)
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

func (s *liveTokenModel) isTeacherByClass(ctx context.Context, op *entity.Operator, classID string) (bool, error) {
	classTeacherMap, err := external.GetTeacherServiceProvider().GetByClasses(ctx, op, []string{classID})
	if err != nil {
		log.Error(ctx, "isTeacherByClass:GetTeacherServiceProvider.GetByClasses error",
			log.Err(err),
			log.String("classID", classID),
			log.Any("op", op),
		)
		return false, err
	}
	teachers, ok := classTeacherMap[classID]
	if !ok {
		log.Info(ctx, "isTeacherByClass:No teacher under the class",
			log.String("classID", classID),
			log.Any("op", op),
		)
		return false, nil
	}
	log.Debug(ctx, "isTeacherByClass:classTeacherMap info",
		log.String("classID", classID),
		log.Any("op", op),
		log.Any("classTeacherMap", classTeacherMap),
	)
	for _, t := range teachers {
		if t.ID == op.UserID {
			return true, nil
		}
	}
	return false, nil
}

func (s *liveTokenModel) isTeacherByPermission(ctx context.Context, op *entity.Operator) (bool, error) {
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.LiveClassTeacher)
	if err != nil {
		log.Error(ctx, "isTeacherByPermission:GetPermissionServiceProvider.HasOrganizationPermission error",
			log.String("permission", external.LiveClassTeacher.String()),
			log.Any("operator", op),
			log.Err(err),
		)
		return false, err
	}
	return hasPermission, nil
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
		mData, ok := item.Data.(*MaterialData)
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
