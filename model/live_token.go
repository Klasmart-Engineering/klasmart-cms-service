package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrGoLiveTimeNotUp = errors.New("go live time not up")
	ErrGoLiveNotAllow  = errors.New("go live not allow")
)

type ILiveTokenModel interface {
	MakeScheduleLiveToken(ctx context.Context, op *entity.Operator, scheduleID string, tokenType entity.LiveTokenType) (string, error)
	MakeContentLiveToken(ctx context.Context, op *entity.Operator, contentID string) (string, error)
}

func (s *liveTokenModel) MakeScheduleLiveToken(ctx context.Context, op *entity.Operator, scheduleID string, tokenType entity.LiveTokenType) (string, error) {
	schedule, err := GetScheduleModel().GetPlainByID(ctx, scheduleID)
	if err != nil {
		log.Error(ctx, "MakeScheduleLiveToken:GetScheduleModel.GetPlainByID error",
			log.Err(err),
			log.Any("op", op),
			log.String("scheduleID", scheduleID))
		return "", err
	}

	if tokenType == entity.LiveTokenTypeLive && schedule.ClassType != entity.ScheduleClassTypeHomework {
		now := time.Now().Unix()
		diff := utils.TimeStampDiff(schedule.StartAt, now)
		if diff >= constant.ScheduleAllowGoLiveTime {
			log.Warn(ctx, "MakeScheduleLiveToken: go live time not up",
				log.Any("op", op),
				log.String("scheduleID", scheduleID),
				log.Int64("schedule.StartAt", schedule.StartAt),
				log.Int64("time.Now", now),
			)
			return "", ErrGoLiveTimeNotUp
		}
		if schedule.Status.GetScheduleStatus(entity.ScheduleStatusInput{
			EndAt:     schedule.EndAt,
			DueAt:     schedule.DueAt,
			ClassType: schedule.ClassType,
		}) == entity.ScheduleStatusClosed {
			log.Warn(ctx, "MakeScheduleLiveToken:go live not allow",
				log.Any("op", op),
				log.Any("schedule", schedule),
				log.Int64("schedule.StartAt", schedule.StartAt),
				log.Int64("time.Now", now),
			)
			return "", ErrGoLiveNotAllow
		}
	}
	classType := schedule.ClassType.ConvertToLiveClassType()
	if classType == entity.LiveClassTypeInvalid {
		log.Error(ctx, "MakeScheduleLiveToken:ConvertToLiveClassType invalid",
			log.Any("op", op),
			log.String("scheduleID", scheduleID),
			log.Any("schedule.ClassType", schedule.ClassType),
		)
		return "", constant.ErrInvalidArgs
	}
	liveTokenInfo := entity.LiveTokenInfo{
		UserID:    op.UserID,
		Type:      tokenType, //entity.LiveTokenTypeLive,
		RoomID:    scheduleID,
		ClassType: classType,
		OrgID:     op.OrgID,
	}
	liveTokenInfo.ScheduleID = schedule.ID

	name, err := s.getUserName(ctx, op)
	if err != nil {
		log.Error(ctx, "MakeScheduleLiveToken:get user name by id error",
			log.Err(err),
			log.Any("op", op),
			log.String("scheduleID", scheduleID))
		return "", err
	}
	liveTokenInfo.Name = name
	isTeacher, err := s.isTeacherByScheduleID(ctx, op, scheduleID)
	if err != nil {
		log.Error(ctx, "MakeScheduleLiveToken:judge is teacher error",
			log.Err(err),
			log.Any("op", op))
		return "", err
	}
	liveTokenInfo.Teacher = isTeacher
	if schedule.ClassType == entity.ScheduleClassTypeTask || (schedule.ClassType == entity.ScheduleClassTypeHomework && schedule.IsHomeFun) {
		liveTokenInfo.Materials = make([]*entity.LiveMaterial, 0)
	} else {
		_, err = GetScheduleModel().VerifyLessonPlanAuthed(ctx, op, schedule.LessonPlanID)
		if err != nil {
			log.Error(ctx, "MakeScheduleLiveToken:GetScheduleModel.VerifyLessonPlanAuthed error",
				log.Err(err),
				log.Any("op", op),
				log.Any("schedule", schedule))
			return "", err
		}
		materialInput := &entity.MaterialInput{
			ScheduleID: scheduleID,
			TokenType:  tokenType,
			ContentID:  schedule.LessonPlanID,
		}
		liveTokenInfo.Materials, err = s.getMaterials(ctx, op, materialInput)
		if err != nil {
			log.Error(ctx, "MakeScheduleLiveToken:get material error",
				log.Err(err),
				log.Any("op", op),
				log.Any("liveTokenInfo", liveTokenInfo),
				log.Any("schedule", schedule))
			return "", err
		}
	}

	expiresAt := time.Now().Add(constant.LiveTokenExpiresAt).Unix()
	if liveTokenInfo.ClassType == entity.LiveClassTypeLive {
		expiresAt = schedule.EndAt + int64(constant.LiveClassTypeLiveTokenExpiresAt.Seconds())
	}

	token, err := s.createJWT(ctx, liveTokenInfo, expiresAt)
	if err != nil {
		log.Error(ctx, "MakeScheduleLiveToken:create jwt error",
			log.Err(err),
			log.Any("op", op),
			log.Any("liveTokenInfo", liveTokenInfo),
			log.Any("schedule", schedule),
			log.Any("expiresAt", expiresAt))
		return "", err
	}
	return token, nil
}

func (s *liveTokenModel) MakeContentLiveToken(ctx context.Context, op *entity.Operator, contentID string) (string, error) {
	liveTokenInfo := entity.LiveTokenInfo{
		UserID:    op.UserID,
		Type:      entity.LiveTokenTypePreview,
		RoomID:    contentID,
		OrgID:     op.OrgID,
		ClassType: entity.LiveClassTypeLive,
	}
	_, err := GetScheduleModel().VerifyLessonPlanAuthed(ctx, op, contentID)
	if err != nil {
		log.Error(ctx, "MakeContentLiveToken:GetScheduleModel.VerifyLessonPlanAuthed error",
			log.Err(err),
			log.Any("op", op),
			log.String("contentID", contentID))
		return "", err
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
	isTeacher, err = s.isTeacherByPermission(ctx, op)
	if err != nil {
		log.Error(ctx, "MakeLivePreviewToken:isTeacherByPermission error",
			log.Err(err),
			log.Any("op", op))
		return "", err
	}
	liveTokenInfo.Teacher = isTeacher
	materialInput := &entity.MaterialInput{
		ContentID: contentID,
		TokenType: entity.LiveTokenTypePreview,
	}
	liveTokenInfo.Materials, err = s.getMaterials(ctx, op, materialInput)
	if err != nil {
		log.Error(ctx, "MakeLivePreviewToken:get material error",
			log.Err(err),
			log.Any("op", op),
			log.Any("liveTokenInfo", liveTokenInfo),
			log.String("contentID", contentID))
		return "", err
	}

	expiresAt := time.Now().Add(constant.LiveTokenExpiresAt).Unix()
	token, err := s.createJWT(ctx, liveTokenInfo, expiresAt)
	if err != nil {
		log.Error(ctx, "MakeLivePreviewToken:create jwt error",
			log.Err(err),
			log.Any("op", op),
			log.Any("liveTokenInfo", liveTokenInfo),
			log.String("contentID", contentID),
			log.Any("expiresAt", expiresAt))
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

func (s *liveTokenModel) createJWT(ctx context.Context, liveTokenInfo entity.LiveTokenInfo, expiresAt int64) (string, error) {
	now := time.Now()

	stdClaims := &jwt.StandardClaims{
		Audience:  "kidsloop-live",
		ExpiresAt: expiresAt,
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

func (s *liveTokenModel) isTeacherByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) (bool, error) {
	isTeacherPermission, err := s.isTeacherByPermission(ctx, op)
	if err != nil {
		log.Error(ctx, "get permissions error", log.Err(err), log.Any("op", op))
		return false, err
	}
	if !isTeacherPermission {
		log.Info(ctx, "has no teacher permission", log.Err(err), log.Any("op", op))
		return false, nil
	}
	isTeacher, err := GetScheduleRelationModel().IsTeacher(ctx, op, scheduleID)
	if err != nil {
		log.Error(ctx, "GetScheduleRelationModel.IsTeacher error", log.Err(err), log.Any("op", op), log.String("scheduleID", scheduleID))
		return false, err
	}
	return isTeacher, nil
}

func (s *liveTokenModel) isTeacherByPermission(ctx context.Context, op *entity.Operator) (bool, error) {
	permissionMap, err := GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, op, []external.PermissionName{
		external.LiveClassTeacher,
		external.LiveClassStudent,
	})
	if err != nil {
		log.Error(ctx, "get permissions error", log.Err(err), log.Any("op", op))
		return false, err
	}
	if permissionMap[external.LiveClassTeacher] {
		return true, nil
	}
	return false, nil
}

func (s *liveTokenModel) getMaterials(ctx context.Context, op *entity.Operator, input *entity.MaterialInput) ([]*entity.LiveMaterial, error) {
	contentList, err := GetContentModel().GetContentSubContentsByID(ctx, dbo.MustGetDB(ctx), input.ContentID, op)
	log.Debug(ctx, "content data", log.Any("contentList", contentList))
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "getMaterials:get content sub by id not found",
			log.Err(err),
			log.Any("input", input))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "getMaterials:get content sub by id error",
			log.Err(err),
			log.Any("input", input))
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
			log.Debug(ctx, "content data convert material data error", log.Any("item", item))
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
		case entity.FileTypeH5p, entity.FileTypeH5pExtend:
			materialItem.TypeName = entity.MaterialTypeH5P
		case entity.FileTypeDocument:
			log.Debug(ctx, "content material doc type", log.Any("op", op), log.Any("content", item))
			//if mData.Source.Ext() != constant.LiveTokenDocumentPDF {
			//	continue
			//}
			materialItem.TypeName = entity.MaterialTypeH5P
		default:
			log.Warn(ctx, "content material type is invalid", log.Any("materialData", mData))
			continue
		}
		// material url
		switch mData.FileType {
		case entity.FileTypeH5pExtend:
			materialItem.URL = fmt.Sprintf("/h5pextend/index.html?org_id=%s&content_id=%s&schedule_id=%s&type=%s#/live-h5p", op.OrgID, item.ID, input.ScheduleID, input.TokenType)
		case entity.FileTypeH5p:
			materialItem.URL = fmt.Sprintf("/h5p/play/%v", mData.Source)
		default:
			source := string(mData.Source)
			parts := strings.Split(source, "-")
			if len(parts) != 2 {
				log.Error(ctx, "invalid resource id", log.String("resourceId", source))
				return nil, constant.ErrInvalidArgs
			}
			materialItem.URL = config.Get().LiveTokenConfig.AssetsUrlPrefix + fmt.Sprintf("/assets/%s", parts[1])
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
