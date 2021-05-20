package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

type IVisibilitySettingModel interface {
	Query(ctx context.Context, contentType int, operator *entity.Operator) ([]*entity.VisibilitySetting, error)
	GetByID(ctx context.Context, id string, operator *entity.Operator) (*entity.VisibilitySetting, error)
}

type visibilitySettingModel struct {
}

func (m *visibilitySettingModel) Query(ctx context.Context, contentType int, operator *entity.Operator) ([]*entity.VisibilitySetting, error) {
	//err := da.GetVisibilitySettingDA().Query(ctx, condget permission failedition, &result)
	//if err != nil {
	//	log.Error(ctx, "query error", log.Err(err), log.Any("condition", condition))
	//	return nil, err
	//}
	ret, err := GetContentPermissionMySchoolModel().GetPermissionOrgs(ctx, external.CreateContentPage201, operator)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
		return nil, err
	}
	orgMap := make(map[string]*entity.VisibilitySetting)
	for i := range ret {
		//user has org publish permission, so he has all publish permissions
		if ret[i].ID == operator.OrgID {
			//query all the schools in the org
			schools, err := external.GetSchoolServiceProvider().GetByOrganizationID(ctx, operator, operator.OrgID)
			if err != nil {
				log.Error(ctx, "query school error",
					log.Err(err),
					log.Any("operator", operator))
				return nil, err
			}
			//Add schools as visibility settings
			result := make([]*entity.VisibilitySetting, len(schools)+1)
			for i := range schools {
				result[i] = &entity.VisibilitySetting{
					ID:   schools[i].ID,
					Name: schools[i].Name,
				}
			}
			//Add org as visibility settings
			result[len(schools)] = &entity.VisibilitySetting{
				ID:   ret[i].ID,
				Name: ret[i].Name,
			}
			return result, nil
		}

		orgMap[ret[i].ID] = &entity.VisibilitySetting{
			ID:   ret[i].ID,
			Name: ret[i].Name,
		}
	}

	if contentType == entity.ContentTypePlan {
		ret2, err := GetContentPermissionMySchoolModel().GetPermissionOrgs(ctx, external.CreateLessonPlan221, operator)
		if err != nil {
			log.Error(ctx, "query lesson error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
			return nil, err
		}
		for i := range ret2 {
			orgMap[ret2[i].ID] = &entity.VisibilitySetting{
				ID:   ret2[i].ID,
				Name: ret2[i].Name,
			}
		}
	} else if contentType == entity.ContentTypeMaterial {
		ret2, err := GetContentPermissionMySchoolModel().GetPermissionOrgs(ctx, external.CreateLessonMaterial220, operator)
		if err != nil {
			log.Error(ctx, "query material error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
			return nil, err
		}
		for i := range ret2 {
			orgMap[ret2[i].ID] = &entity.VisibilitySetting{
				ID:   ret2[i].ID,
				Name: ret2[i].Name,
			}
		}
	}

	result := make([]*entity.VisibilitySetting, len(orgMap))
	i := 0
	for _, v := range orgMap {
		result[i] = v
		i++
	}

	return result, nil
}

func (m *visibilitySettingModel) GetByID(ctx context.Context, id string, operator *entity.Operator) (*entity.VisibilitySetting, error) {
	ret, err := GetContentPermissionMySchoolModel().GetPermissionOrgs(ctx, external.CreateContentPage201, operator)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.String("id", id), log.Any("operator", operator))
		return nil, err
	}
	ret2, err := GetContentPermissionMySchoolModel().GetPermissionOrgs(ctx, external.CreateLessonPlan221, operator)
	if err != nil {
		log.Error(ctx, "query lesson error", log.Err(err), log.String("id", id), log.Any("operator", operator))
		return nil, err
	}
	ret = append(ret, ret2...)
	ret3, err := GetContentPermissionMySchoolModel().GetPermissionOrgs(ctx, external.CreateLessonMaterial220, operator)
	if err != nil {
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
