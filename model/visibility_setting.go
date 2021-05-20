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
	orgInfoList, err := external.GetOrganizationServiceProvider().BatchGet(ctx, operator, []string{operator.OrgID})
	if err != nil || len(orgInfoList) < 1 {
		log.Error(ctx, "query error", log.Err(err),
			log.Int("contentType", contentType),
			log.Any("operator", operator))
		return nil, err
	}
	orgInfo := orgInfoList[0]
	if contentType == entity.ContentTypeAssets {
		hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.CreateAsset320)
		if err != nil {
			log.Error(ctx, "query error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
			return nil, err
		}
		if !hasPermission {
			return nil, nil
		}
		return []*entity.VisibilitySetting{
			{
				ID:   orgInfo.ID,
				Name: orgInfo.Name,
			},
		}, nil
	}
	ret := make([]*entity.VisibilitySetting, 0)

	schoolInfo, err := GetContentFilterModel().QueryUserSchools(ctx, operator)
	if err != nil {
		log.Error(ctx, "QueryUserSchools error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
		return nil, err
	}

	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.CreateAllSchoolsContent224)
	if err != nil {
		log.Error(ctx, "HasOrganizationPermission error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
		return nil, err
	}
	if hasPermission {
		schools, err := external.GetSchoolServiceProvider().BatchGet(ctx, operator, schoolInfo.AllSchool)
		if err != nil {
			log.Error(ctx, "GetSchoolServiceProvider.BatchGet error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
			return nil, err
		}
		for i := range schools {
			ret = append(ret, &entity.VisibilitySetting{
				ID:   schools[i].ID,
				Name: schools[i].Name,
			})
		}
	} else {
		hasPermission, err = external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.CreateMySchoolsContent223)
		if err != nil {
			log.Error(ctx, "HasOrganizationPermission error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
			return nil, err
		}
		if hasPermission {
			schools, err := external.GetSchoolServiceProvider().BatchGet(ctx, operator, schoolInfo.AllSchool)
			if err != nil {
				log.Error(ctx, "GetSchoolServiceProvider.BatchGet error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
				return nil, err
			}
			for i := range schools {
				ret = append(ret, &entity.VisibilitySetting{
					ID:   schools[i].ID,
					Name: schools[i].Name,
				})
			}
		}
	}

	if contentType == entity.ContentTypePlan {
		hasPermission, err = external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.CreateLessonPlan221)
		if err != nil {
			log.Error(ctx, "query error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
			return nil, err
		}
		if hasPermission {
			ret = append(ret, &entity.VisibilitySetting{
				ID:   orgInfo.ID,
				Name: orgInfo.Name,
			})
		}
	} else if contentType == entity.ContentTypeMaterial {
		hasPermission, err = external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.CreateLessonMaterial220)
		if err != nil {
			log.Error(ctx, "query error", log.Err(err), log.Int("contentType", contentType), log.Any("operator", operator))
			return nil, err
		}
		if hasPermission {
			ret = append(ret, &entity.VisibilitySetting{
				ID:   orgInfo.ID,
				Name: orgInfo.Name,
			})
		}
	}

	return ret, nil
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
