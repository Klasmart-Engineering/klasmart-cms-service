package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"sync"
)

type ISchedulePermissionModel interface {
	//GetClassIDs(ctx context.Context, op *entity.Operator) ([]string, error)
	HasScheduleEditPermission(ctx context.Context, op *entity.Operator, classID string) error
	HasScheduleOrgPermission(ctx context.Context, op *entity.Operator, permissionName external.PermissionName) error
	HasScheduleOrgPermissions(ctx context.Context, op *entity.Operator, permissionNames []external.PermissionName) (map[external.PermissionName]bool, error)
	HasClassesPermission(ctx context.Context, op *entity.Operator, classIDs []string) error
	GetSchoolsByOperator(ctx context.Context, op *entity.Operator) ([]*entity.ScheduleFilterSchool, error)
	GetClassesByOperator(ctx context.Context, op *entity.Operator, schoolID string) ([]*entity.ScheduleFilterClass, error)
}

type schedulePermissionModel struct {
	testSchedulePermissionRepeatFlag bool
}

func (s *schedulePermissionModel) GetClassesByOperator(ctx context.Context, op *entity.Operator, schoolID string) ([]*entity.ScheduleFilterClass, error) {
	permissionMap, err := s.HasScheduleOrgPermissions(ctx, op, []external.PermissionName{
		external.ScheduleViewOrgCalendar,
		external.ScheduleViewSchoolCalendar,
		external.ScheduleViewMyCalendar,
	})
	if err == constant.ErrForbidden {
		log.Error(ctx, "get Classes forbidden", log.Any("op", op))
		return nil, err
	}
	if err != nil {
		log.Error(ctx, "get Classes error", log.Any("op", op))
		return nil, err
	}

	classMap := make(map[string]*entity.ScheduleFilterClass)
	classIDs := make([]string, 0)
	if schoolID != "" {
		// get class by school id
		schoolClassMap, err := external.GetClassServiceProvider().GetBySchoolIDs(ctx, op, []string{schoolID})
		if err != nil {
			log.Error(ctx, "GetClassServiceProvider.GetBySchoolIDs error", log.Any("op", op), log.String("SchoolID", schoolID))
			return nil, err
		}
		if len(schoolClassMap) > 0 {
			classList := schoolClassMap[schoolID]
			for _, item := range classList {
				classMap[item.ID] = &entity.ScheduleFilterClass{
					ID:             item.ID,
					Name:           item.Name,
					HasStudentFlag: false,
				}
				classIDs = append(classIDs, item.ID)
			}
		}
	} else {
		// TODO:school id is empty,it means user selected others

	}

	if permissionMap[external.ScheduleViewOrgCalendar] || permissionMap[external.ScheduleViewSchoolCalendar] {
		if permissionMap[external.ScheduleViewMyCalendar] {
			classStuMap, err := external.GetStudentServiceProvider().GetByClassIDs(ctx, op, classIDs)
			if err != nil {
				log.Error(ctx, "GetStudentServiceProvider.GetByClassIDs error", log.Any("op", op), log.Strings("classIDs", classIDs))
				return nil, err
			}
			// Judge you are a student in the class
			for _, item := range classMap {
				stuList := classStuMap[item.ID]
				for _, stu := range stuList {
					if stu.ID == op.UserID {
						item.HasStudentFlag = true
						break
					}
				}
			}
		}

		result := make([]*entity.ScheduleFilterClass, 0, len(classMap))
		for _, item := range classMap {
			result = append(result, item)
		}
		return result, nil
	}
	if permissionMap[external.ScheduleViewMyCalendar] {
		classList, err := external.GetClassServiceProvider().GetByUserID(ctx, op, op.UserID)
		if err != nil {
			log.Error(ctx, "GetClassServiceProvider.GetByUserID error", log.Any("op", op))
			return nil, err
		}
		result := make([]*entity.ScheduleFilterClass, 0)
		for _, item := range classList {
			if classItem, ok := classMap[item.ID]; ok {
				result = append(result, classItem)
			}
		}
		return result, nil
	}

	return nil, constant.ErrForbidden
}

func (s *schedulePermissionModel) GetSchoolsByOperator(ctx context.Context, op *entity.Operator) ([]*entity.ScheduleFilterSchool, error) {
	permissionMap, err := s.HasScheduleOrgPermissions(ctx, op, []external.PermissionName{
		external.ScheduleViewOrgCalendar,
		external.ScheduleViewSchoolCalendar,
		external.ScheduleViewMyCalendar,
	})
	if err == constant.ErrForbidden {
		log.Error(ctx, "get schools forbidden", log.Any("op", op))
		return nil, err
	}
	if err != nil {
		log.Error(ctx, "get schools error", log.Any("op", op))
		return nil, err
	}
	if permissionMap[external.ScheduleViewOrgCalendar] {
		schoolList, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, external.ScheduleViewOrgCalendar)
		if err != nil {
			log.Error(ctx, "GetSchoolServiceProvider.GetByPermission error", log.Any("op", op), log.String("permissionName", external.ScheduleViewOrgCalendar.String()))
			return nil, err
		}
		return schoolList, nil
	}
	if permissionMap[external.ScheduleViewSchoolCalendar] {
		schoolList, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, external.ScheduleViewSchoolCalendar)
		if err != nil {
			log.Error(ctx, "GetSchoolServiceProvider.GetByPermission error", log.Any("op", op), log.String("permissionName", external.ScheduleViewSchoolCalendar.String()))
			return nil, err
		}
		return schoolList, nil
	}
	if permissionMap[external.ScheduleViewMyCalendar] {
		schoolList, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, external.ScheduleViewMyCalendar)
		if err != nil {
			log.Error(ctx, "GetSchoolServiceProvider.GetByPermission error", log.Any("op", op), log.String("permissionName", external.ScheduleViewMyCalendar.String()))
			return nil, err
		}
		return schoolList, nil
	}
	return nil, constant.ErrForbidden
}

func (s *schedulePermissionModel) HasClassesPermission(ctx context.Context, op *entity.Operator, classIDs []string) error {
	classOrgMap, err := external.GetOrganizationServiceProvider().GetByClasses(ctx, op, classIDs)
	if err != nil {
		log.Error(ctx, "hasScheduleEditPermission:GetOrganizationServiceProvider.GetByClasses error",
			log.Any("operator", op),
			log.Strings("classIDs", classIDs),
			log.Err(err),
		)
		return err
	}
	for _, classID := range classIDs {
		orgInfo, ok := classOrgMap[classID]
		if !ok {
			log.Info(ctx, "hasScheduleEditPermission:class not found org",
				log.Any("operator", op),
				log.Strings("classIDs", classIDs),
				log.String("err classID", classID),
			)
			return constant.ErrForbidden
		}
		if orgInfo.ID != op.OrgID {
			log.Info(ctx, "hasScheduleEditPermission:class org not equal operator org",
				log.Any("operator", op),
				log.Any("orgInfo", orgInfo),
				log.Strings("classIDs", classIDs),
				log.String("err classID", classID),
			)
			return constant.ErrForbidden
		}
	}
	return nil
}

//func (s *schedulePermissionModel) GetClassIDs(ctx context.Context, op *entity.Operator) ([]string, error) {
//	schoolClassIDs, err := s.GetClassIDsBySchoolPermission(ctx, op, external.ScheduleViewSchoolCalendar)
//	if err != nil {
//		log.Error(ctx, "getClassIDsByPermission:getClassIDsBySchoolPermission error",
//			log.Any("operator", op),
//			log.Err(err),
//		)
//		return nil, err
//	}
//
//	orgClassIDs, err := s.GetClassIDsByOrgPermission(ctx, op, external.ScheduleViewOrgCalendar)
//	if err != nil {
//		log.Error(ctx, "getClassIDsByPermission:getClassIDsByOrgPermission error",
//			log.Any("operator", op),
//			log.Err(err),
//		)
//		return nil, err
//	}
//
//	myClassIDs := make([]string, 0)
//	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ScheduleViewMyCalendar)
//	if err != nil {
//		log.Error(ctx, "getScheduleTimeView:GetPermissionServiceProvider.HasOrganizationPermission error",
//			log.Err(err),
//			log.String("PermissionName", external.ScheduleViewMyCalendar.String()),
//			log.Any("op", op),
//		)
//		return nil, err
//	}
//	if hasPermission {
//		myClassIDs, err = GetScheduleModel().GetOrgClassIDsByUserIDs(ctx, op, []string{op.UserID}, op.OrgID)
//		if err != nil {
//			log.Error(ctx, "getScheduleTimeView:GetScheduleModel.GetMyOrgClassIDs error",
//				log.Err(err),
//				log.Any("op", op),
//			)
//			return nil, err
//		}
//	}
//
//	log.Info(ctx, "getClassIDs", log.Any("Operator", op),
//		log.Strings("schoolClassIDs", schoolClassIDs),
//		log.Strings("orgClassIDs", orgClassIDs),
//		log.Strings("myClassIDs", myClassIDs),
//	)
//	classIDs := make([]string, 0, len(schoolClassIDs)+len(orgClassIDs)+len(myClassIDs))
//	classIDs = append(classIDs, schoolClassIDs...)
//	classIDs = append(classIDs, orgClassIDs...)
//	classIDs = append(classIDs, myClassIDs...)
//
//	classIDs = utils.SliceDeduplication(classIDs)
//
//	return classIDs, nil
//}

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
	for _, classList := range classMap {
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
	err := s.HasClassesPermission(ctx, op, []string{classID})
	if err != nil {
		log.Error(ctx, "hasScheduleEditPermission:HasClassesPermission error",
			log.String("classID", classID),
			log.Any("operator", op),
			log.Err(err),
		)
		return err
	}
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

func (s *schedulePermissionModel) HasScheduleOrgPermissions(ctx context.Context, op *entity.Operator, permissionNames []external.PermissionName) (map[external.PermissionName]bool, error) {
	permissionMap, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, permissionNames)
	if err != nil {
		log.Error(ctx, "check permission error",
			log.Any("permission", permissionNames),
			log.Any("operator", op),
			log.Err(err),
		)

		return permissionMap, constant.ErrInternalServer
	}
	hasOne := false
	for _, val := range permissionMap {
		if val {
			hasOne = true
			break
		}
	}
	if !hasOne {
		log.Info(ctx, "no permission",
			log.Any("permission", permissionNames),
			log.Any("Operator", op),
		)

		return permissionMap, constant.ErrForbidden
	}
	log.Debug(ctx, "permission",
		log.Any("permission", permissionNames),
		log.Any("Operator", op),
		log.Any("permissionMap", permissionMap),
	)
	return permissionMap, nil
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
