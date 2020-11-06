package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"sync"
)

type IVisibilitySettingModel interface {
	Query(ctx context.Context, contentType int, operator *entity.Operator) ([]*entity.VisibilitySetting, error)
	GetByID(ctx context.Context, id string, operator *entity.Operator) (*entity.VisibilitySetting, error)
}

type visibilitySettingModel struct {
}

func (m *visibilitySettingModel) Query(ctx context.Context, contentType int, operator *entity.Operator) ([]*entity.VisibilitySetting, error) {
	//err := da.GetVisibilitySettingDA().Query(ctx, condition, &result)
	//if err != nil {
	//	log.Error(ctx, "query error", log.Err(err), log.Any("condition", condition))
	//	return nil, err
	//}
	ret, err := GetContentPermissionModel().GetPermissionedOrgs(ctx, external.CreateContentPage201, operator)
	if err != nil{
		log.Error(ctx, "query error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
		return nil, err
	}
	if contentType == entity.ContentTypeLesson {
		ret2, err := GetContentPermissionModel().GetPermissionedOrgs(ctx, external.CreateLessonPlan221, operator)
		if err != nil{
			log.Error(ctx, "query lesson error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
			return nil, err
		}
		ret = append(ret, ret2...)
	}else if contentType == entity.ContentTypeMaterial {
		ret2, err := GetContentPermissionModel().GetPermissionedOrgs(ctx, external.CreateLessonMaterial220, operator)
		if err != nil{
			log.Error(ctx, "query material error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
			return nil, err
		}
		ret = append(ret, ret2...)
	}


	result := make([]*entity.VisibilitySetting, len(ret))
	for i := range ret {
		result[i] = &entity.VisibilitySetting{
			ID:   ret[i].ID,
			Name: ret[i].Name,
		}
	}


	return result, nil
}

func (m *visibilitySettingModel) GetByID(ctx context.Context, id string, operator *entity.Operator) (*entity.VisibilitySetting, error) {
	ret, err := GetContentPermissionModel().GetPermissionedOrgs(ctx, external.CreateContentPage201, operator)
	if err != nil{
		log.Error(ctx, "query error", log.Err(err), log.String("id", id), log.Any("operator", operator))
		return nil, err
	}
	ret2, err := GetContentPermissionModel().GetPermissionedOrgs(ctx, external.CreateLessonPlan221, operator)
	if err != nil{
		log.Error(ctx, "query lesson error", log.Err(err), log.String("id", id), log.Any("operator", operator))
		return nil, err
	}
	ret = append(ret, ret2...)
	ret3, err := GetContentPermissionModel().GetPermissionedOrgs(ctx, external.CreateLessonMaterial220, operator)
	if err != nil{
		log.Error(ctx, "query material error", log.Err(err), log.String("id", id), log.Any("operator", operator))
		return nil, err
	}
	ret = append(ret, ret3...)

	for i := range ret {
		if id == ret[i].ID {
			return &entity.VisibilitySetting{
				ID:   ret[i].ID,
				Name: ret[i].Name,
			}, nil
		}
	}

	return nil, ErrInvalidVisibleScope
}

var (
	_visibilitySettingOnce  sync.Once
	_visibilitySettingModel IVisibilitySettingModel
)

func GetVisibilitySettingModel() IVisibilitySettingModel {
	_visibilitySettingOnce.Do(func() {
		_visibilitySettingModel = &visibilitySettingModel{}
	})
	return _visibilitySettingModel
}
