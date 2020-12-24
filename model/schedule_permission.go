package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type ISchedulePermissionModel interface {
	GetClassIDs(ctx context.Context, op *entity.Operator) ([]string, error)
	HasScheduleEditPermission(ctx context.Context, op *entity.Operator, classID string) error
	HasScheduleOrgPermission(ctx context.Context, op *entity.Operator, permissionName external.PermissionName) error
}

type schedulePermissionModel struct {
	testSchedulePermissionRepeatFlag bool
}

func (s *schedulePermissionModel) GetClassIDs(ctx context.Context, op *entity.Operator) ([]string, error) {
	schoolClassIDs, err := s.GetClassIDsBySchoolPermission(ctx, op, external.ScheduleViewSchoolCalendar)
	if err != nil {
		log.Error(ctx, "getClassIDsByPermission:getClassIDsBySchoolPermission error",
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}

	orgClassIDs, err := s.GetClassIDsByOrgPermission(ctx, op, external.ScheduleViewOrgCalendar)
	if err != nil {
		log.Error(ctx, "getClassIDsByPermission:getClassIDsByOrgPermission error",
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}

	myClassIDs := make([]string, 0)
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ScheduleViewMyCalendar)
	if err != nil {
		log.Error(ctx, "getScheduleTimeView:GetPermissionServiceProvider.HasOrganizationPermission error",
			log.Err(err),
			log.String("PermissionName", external.ScheduleViewMyCalendar.String()),
			log.Any("op", op),
		)
		return nil, err
	}
	if hasPermission {
		myClassIDs, err = GetScheduleModel().GetOrgClassIDsByUserIDs(ctx, op, []string{op.UserID}, op.OrgID)
		if err != nil {
			log.Error(ctx, "getScheduleTimeView:GetScheduleModel.GetMyOrgClassIDs error",
				log.Err(err),
				log.Any("op", op),
			)
			return nil, err
		}
	}

	log.Info(ctx, "getClassIDs", log.Any("Operator", op),
		log.Strings("schoolClassIDs", schoolClassIDs),
		log.Strings("orgClassIDs", orgClassIDs),
		log.Strings("myClassIDs", myClassIDs),
	)
	classIDs := make([]string, 0, len(schoolClassIDs)+len(orgClassIDs)+len(myClassIDs))
	classIDs = append(classIDs, schoolClassIDs...)
	classIDs = append(classIDs, orgClassIDs...)
	classIDs = append(classIDs, myClassIDs...)

	classIDs = utils.SliceDeduplication(classIDs)

	return classIDs, nil
}

func (s *schedulePermissionModel) GetClassIDsBySchoolPermission(ctx context.Context, op *entity.Operator, permissionName external.PermissionName) ([]string, error) {
	classIDs := make([]string, 0)
	schoolInfoList, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, permissionName)
	if err != nil {
		log.Error(ctx, "check permission error",
			log.String("permission", permissionName.String()),
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}
	schoolIDs := make([]string, len(schoolInfoList))
	for i, item := range schoolInfoList {
		schoolIDs[i] = item.ID
	}
	classMap, err := external.GetClassServiceProvider().GetBySchoolIDs(ctx, op, schoolIDs)
	if err != nil {
		log.Error(ctx, "getClassIDsBySchoolPermission:GetClassServiceProvider GetBySchoolIDs error",
			log.String("permission", permissionName.String()),
			log.Strings("schoolIDs", schoolIDs),
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}
	if classList, ok := classMap[op.OrgID]; ok {
		for _, item := range classList {
			classIDs = append(classIDs, item.ID)
		}
	}
	return classIDs, nil
}

func (s *schedulePermissionModel) GetClassIDsByOrgPermission(ctx context.Context, op *entity.Operator, permissionName external.PermissionName) ([]string, error) {
	//external.ScheduleViewOrgCalendar
	orgInfoList, err := external.GetOrganizationServiceProvider().GetByPermission(ctx, op, permissionName)
	if err != nil {
		log.Error(ctx, "getClassIDsByOrgPermission：check permission error",
			log.String("permission", permissionName.String()),
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}
	orgIDs := make([]string, len(orgInfoList))
	for i, item := range orgInfoList {
		orgIDs[i] = item.ID
	}
	log.Info(ctx, "getClassIDsByOrgPermission", log.Any("orgInfoList", orgInfoList))
	classMap, err := external.GetClassServiceProvider().GetByOrganizationIDs(ctx, op, orgIDs)
	if err != nil {
		log.Error(ctx, "getClassIDsByOrgPermission:GetClassServiceProvider GetByOrganizationIDs error",
			log.String("permission", permissionName.String()),
			log.Strings("orgIDs", orgIDs),
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}
	log.Info(ctx, "getClassIDsByOrgPermission", log.Any("op", op), log.Any("classMap", classMap), log.Any("classMap[op.OrgID]", classMap[op.OrgID]))
	var classIDs []string
	if classList, ok := classMap[op.OrgID]; ok {
		log.Info(ctx, "getClassIDsByOrgPermission", log.Any("classList", classList))
		classIDs = make([]string, len(classList))
		for i, item := range classList {
			classIDs[i] = item.ID
		}
	}
	log.Info(ctx, "getClassIDsByOrgPermission", log.Strings("classIDs", classIDs), log.String("permissionName", permissionName.String()))
	return classIDs, nil
}

func (s *schedulePermissionModel) HasScheduleEditPermission(ctx context.Context, op *entity.Operator, classID string) error {
	permissionName := external.ScheduleCreateEvent
	ok, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permissionName)
	if err != nil {
		return err
	}
	if ok {
		orgClassMap, err := external.GetClassServiceProvider().GetByOrganizationIDs(ctx, op, []string{op.OrgID})
		if err != nil {
			log.Error(ctx, "hasScheduleEditPermission:GetClassServiceProvider.GetByOrganizationIDs error",
				log.String("permission", permissionName.String()),
				log.Any("operator", op),
				log.Err(err),
			)

			return constant.ErrInternalServer
		}
		classList, ok := orgClassMap[op.OrgID]
		if !ok {
			log.Info(ctx, "hasScheduleEditPermission:org has no class",
				log.String("permission", permissionName.String()),
				log.Any("operator", op),
				log.Err(err),
			)

			return constant.ErrForbidden
		}
		for _, item := range classList {
			if item.ID == classID {
				return nil
			}
		}
	}

	permissionName = external.ScheduleCreateMySchoolEvent
	ok, err = external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permissionName)
	if err != nil {
		return err
	}
	if ok {
		schoolList, err := external.GetSchoolServiceProvider().GetByOrganizationID(ctx, op, op.OrgID)
		if err != nil {
			log.Error(ctx, "hasScheduleEditPermission:GetSchoolServiceProvider.GetByOrganizationID error",
				log.String("permission", permissionName.String()),
				log.Any("operator", op),
				log.Err(err),
			)

			return constant.ErrInternalServer
		}
		schoolIDs := make([]string, len(schoolList))
		for i, item := range schoolList {
			schoolIDs[i] = item.ID
		}
		schoolClassMap, err := external.GetClassServiceProvider().GetBySchoolIDs(ctx, op, schoolIDs)
		if err != nil {
			log.Error(ctx, "hasScheduleEditPermission:GetClassServiceProvider.GetBySchoolIDs error",
				log.String("permission", permissionName.String()),
				log.Any("operator", op),
				log.Err(err),
			)

			return constant.ErrInternalServer
		}
		for _, classList := range schoolClassMap {
			for _, item := range classList {
				if item.ID == classID {
					return nil
				}
			}
		}
	}

	permissionName = external.ScheduleCreateMyEvent
	ok, err = external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permissionName)
	if err != nil {
		return err
	}
	if ok {
		classList, err := external.GetClassServiceProvider().GetByUserID(ctx, op, op.UserID)
		if err != nil {
			log.Error(ctx, "hasScheduleEditPermission:GetClassServiceProvider.GetByUserID error",
				log.String("permission", permissionName.String()),
				log.Any("operator", op),
				log.Err(err),
			)
			return constant.ErrInternalServer
		}
		for _, item := range classList {
			if item.ID == classID {
				return nil
			}
		}
	}
	log.Info(ctx, "hasScheduleEditPermission:no permission", log.Any("operator", op), log.String("classID", classID))
	return constant.ErrForbidden
}

func (s *schedulePermissionModel) HasScheduleOrgPermission(ctx context.Context, op *entity.Operator, permissionName external.PermissionName) error {
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permissionName)
	if err != nil {
		log.Error(ctx, "check permission error",
			log.String("permission", string(permissionName)),
			log.Any("operator", op),
			log.Err(err),
		)

		return constant.ErrInternalServer
	}
	if !hasPermission {
		log.Info(ctx, "no permission",
			log.String("permission", string(permissionName)),
			log.Any("Operator", op),
		)

		return constant.ErrForbidden
	}
	return nil
}

var (
	_schedulePermissionOnce  sync.Once
	_schedulePermissionModel ISchedulePermissionModel
)

func GetSchedulePermissionModel() ISchedulePermissionModel {
	_schedulePermissionOnce.Do(func() {
		_schedulePermissionModel = &schedulePermissionModel{}
	})
	return _schedulePermissionModel
}
