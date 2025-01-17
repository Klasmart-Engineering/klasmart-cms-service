package model

import (
	"context"
	"sync"

	"github.com/KL-Engineering/common-log/log"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
)

type ISchedulePermissionModel interface {
	//GetClassIDs(ctx context.Context, op *entity.Operator) ([]string, error)
	HasScheduleEditPermission(ctx context.Context, op *entity.Operator, classID string) error
	HasScheduleOrgPermission(ctx context.Context, op *entity.Operator, permissionName external.PermissionName) error
	HasScheduleOrgPermissions(ctx context.Context, op *entity.Operator, permissionNames []external.PermissionName) (map[external.PermissionName]bool, error)
	HasClassesPermission(ctx context.Context, op *entity.Operator, classIDs []string) error
	GetOnlyUnderOrgUsers(ctx context.Context, op *entity.Operator) ([]*external.User, error)
	GetUnDefineClass(ctx context.Context, op *entity.Operator, permissionMap map[external.PermissionName]bool) (*entity.ScheduleFilterClass, error)
}

type schedulePermissionModel struct {
}

func (s *schedulePermissionModel) GetUnDefineClass(ctx context.Context, op *entity.Operator, permissionMap map[external.PermissionName]bool) (*entity.ScheduleFilterClass, error) {
	hasUnDefineClass, err := s.HasUnDefineClass(ctx, op, permissionMap)
	if err != nil {
		return nil, err
	}
	if !hasUnDefineClass {
		return nil, constant.ErrRecordNotFound
	}
	result := &entity.ScheduleFilterClass{
		ID:               entity.ScheduleFilterUndefinedClass,
		Name:             entity.ScheduleFilterUndefinedClass,
		OperatorRoleType: entity.ScheduleRoleTypeUnknown,
	}

	return result, nil
}

func (s *schedulePermissionModel) HasUnDefineClass(ctx context.Context, op *entity.Operator, permissionMap map[external.PermissionName]bool) (bool, error) {
	cacheData, err := da.GetScheduleRedisDA().GetScheduleFilterUndefinedClass(ctx, op.OrgID, permissionMap)
	if err == nil {
		log.Info(ctx, "Query:using cache",
			log.Any("op", op),
			log.Any("permissionMap", permissionMap),
			log.Err(err),
		)
		return cacheData, nil
	}

	userInfos, err := s.GetOnlyUnderOrgUsers(ctx, op)
	if err != nil {
		log.Error(ctx, "GetOnlyUnderOrgUsers error", log.Any("op", op))
		return false, err
	}

	if len(userInfos) <= 0 {
		return false, nil
	}

	userIDs := make([]string, len(userInfos))
	hasOperator := false
	for i, item := range userInfos {
		userIDs[i] = item.ID
		if item.ID == op.UserID {
			hasOperator = true
		}
	}
	if permissionMap[external.ScheduleViewOrgCalendar] {
		// org permission
		log.Debug(ctx, "org permission", log.Strings("userIDs", userIDs), log.Any("op", op))
	} else if permissionMap[external.ScheduleViewMyCalendar] && hasOperator {
		userIDs = []string{op.UserID}
	} else {
		return false, nil
	}

	hasSchedule, err := GetScheduleRelationModel().HasScheduleByRelationIDs(ctx, op, userIDs)
	if err != nil {
		log.Error(ctx, "has schedule by relation ids", log.Strings("userIDs", userIDs), log.Any("op", op))
		return false, err
	}

	if err = da.GetScheduleRedisDA().Set(ctx, op.OrgID, &da.ScheduleCacheCondition{
		PermissionMap: permissionMap,
		DataType:      da.ScheduleFilterUndefinedClass,
	}, hasSchedule); err != nil {
		log.Warn(ctx, "set cache error",
			log.Err(err),
			log.Any("permissionMap", permissionMap),
			log.Any("data", hasSchedule))
	}

	return hasSchedule, nil
}

func (s *schedulePermissionModel) GetOnlyUnderOrgUsers(ctx context.Context, op *entity.Operator) ([]*external.User, error) {
	users, err := da.GetUserRedisDA().GetUsersByOrg(ctx, op.OrgID)
	if err == nil && len(users) > 0 {
		return s.getOnlyUnderOrgUsers(users)
	}

	users, err = s.getUsersByOrg(ctx, op)
	if err != nil {
		log.Error(ctx, "s.getUsersByOrg error",
			log.Any("op", op),
			log.Err(err))
	}

	return s.getOnlyUnderOrgUsers(users)
}

func (s *schedulePermissionModel) GetOnlyUnderOrgClasses(ctx context.Context, op *entity.Operator, permissionMap map[external.PermissionName]bool) ([]*entity.ScheduleFilterClass, error) {
	classInfos, err := external.GetClassServiceProvider().GetOnlyUnderOrgClasses(ctx, op, op.OrgID)
	if err != nil {
		log.Error(ctx, "get only under org classes error", log.Any("op", op))
		return nil, err
	}
	underOrgClassIDs := make([]string, 0, len(classInfos))
	for _, classItem := range classInfos {
		if classItem.Valid {
			underOrgClassIDs = append(underOrgClassIDs, classItem.ID)
		}
	}

	var result []*entity.ScheduleFilterClass
	if permissionMap[external.ScheduleViewOrgCalendar] {
		result = make([]*entity.ScheduleFilterClass, 0, len(classInfos))

		for _, item := range classInfos {
			if item.Valid {
				result = append(result, &entity.ScheduleFilterClass{
					ID:               item.ID,
					Name:             item.Name,
					OperatorRoleType: entity.ScheduleRoleTypeUnknown,
				})
			}
		}
	} else if permissionMap[external.ScheduleViewMyCalendar] {
		operatorClasses, err := external.GetClassServiceProvider().GetByUserID(ctx, op, op.UserID)
		if err != nil {
			log.Error(ctx, "get operator classes IDs error", log.Any("op", op))
			return nil, err
		}
		operatorClassMap := make(map[string]*external.Class, len(operatorClasses))
		for _, item := range operatorClasses {
			operatorClassMap[item.ID] = item
		}

		result = make([]*entity.ScheduleFilterClass, 0, len(operatorClasses))
		for _, item := range classInfos {
			if _, ok := operatorClassMap[item.ID]; ok {
				result = append(result, &entity.ScheduleFilterClass{
					ID:               item.ID,
					Name:             item.Name,
					OperatorRoleType: entity.ScheduleRoleTypeUnknown,
				})
			}
		}
	}
	log.Debug(ctx, "no permission", log.Any("result", result), log.Any("permissionMap", permissionMap))
	if len(result) <= 0 {
		return nil, constant.ErrRecordNotFound
	}

	classStuMap, err := external.GetStudentServiceProvider().GetByClassIDs(ctx, op, underOrgClassIDs)
	if err != nil {
		log.Error(ctx, "GetStudentServiceProvider.GetByClassIDs error", log.Any("op", op), log.Strings("underOrgClassIDs", underOrgClassIDs))
		return nil, err
	}
	classTeacherMap, err := external.GetTeacherServiceProvider().GetByClasses(ctx, op, underOrgClassIDs)
	if err != nil {
		log.Error(ctx, "GetTeacherServiceProvider.GetByClasses error", log.Any("op", op), log.Strings("underOrgClassIDs", underOrgClassIDs))
		return nil, err
	}

	for _, item := range result {
		teacherList := classTeacherMap[item.ID]
		for _, teacher := range teacherList {
			if teacher.ID == op.UserID {
				item.OperatorRoleType = entity.ScheduleRoleTypeTeacher
				break
			}
		}
		if item.OperatorRoleType == entity.ScheduleRoleTypeTeacher {
			continue
		}
		stuList := classStuMap[item.ID]
		for _, stu := range stuList {
			if stu.ID == op.UserID {
				item.OperatorRoleType = entity.ScheduleRoleTypeStudent
				break
			}
		}
	}
	return result, nil
}

func (s *schedulePermissionModel) GetClassesBySchoolID(ctx context.Context, op *entity.Operator, schoolID string) ([]*entity.ScheduleFilterClass, error) {
	if schoolID == "" {
		log.Info(ctx, "school id is empty", log.Any("op", op), log.String("schoolID", schoolID))
		return nil, constant.ErrInvalidArgs
	}
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
				ID:               item.ID,
				Name:             item.Name,
				OperatorRoleType: entity.ScheduleRoleTypeUnknown,
			}
			classIDs = append(classIDs, item.ID)
		}
	}

	if permissionMap[external.ScheduleViewOrgCalendar] || permissionMap[external.ScheduleViewSchoolCalendar] {
		if permissionMap[external.ScheduleViewMyCalendar] {
			classStuMap, err := external.GetStudentServiceProvider().GetByClassIDs(ctx, op, classIDs)
			if err != nil {
				log.Error(ctx, "GetStudentServiceProvider.GetByClassIDs error", log.Any("op", op), log.Strings("classIDs", classIDs))
				return nil, err
			}
			classTeacherMap, err := external.GetTeacherServiceProvider().GetByClasses(ctx, op, classIDs)
			if err != nil {
				log.Error(ctx, "GetTeacherServiceProvider.GetByClasses error", log.Any("op", op), log.Strings("classIDs", classIDs))
				return nil, err
			}
			// Judge you are a student in the class
			for _, item := range classMap {
				teacherList := classTeacherMap[item.ID]
				for _, teacher := range teacherList {
					if teacher.ID == op.UserID {
						item.OperatorRoleType = entity.ScheduleRoleTypeTeacher
						break
					}
				}
				if item.OperatorRoleType == entity.ScheduleRoleTypeTeacher {
					continue
				}
				stuList := classStuMap[item.ID]
				for _, stu := range stuList {
					if stu.ID == op.UserID {
						item.OperatorRoleType = entity.ScheduleRoleTypeUnknown
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

func (s *schedulePermissionModel) getUsersByOrg(ctx context.Context, op *entity.Operator) ([]*da.User, error) {
	userInfos, err := external.GetUserServiceProvider().GetByOrganization(ctx, op, op.OrgID)
	if err != nil {
		log.Error(ctx, "GetUserServiceProvider.GetByOrganization error",
			log.Any("op", op),
			log.Err(err))
		return nil, err
	}

	userIDs := make([]string, len(userInfos))
	for i, item := range userInfos {
		userIDs[i] = item.ID
	}

	userSchoolMap, err := external.GetSchoolServiceProvider().GetByUsers(ctx, op, op.OrgID, userIDs)
	if err != nil {
		log.Error(ctx, "GetSchoolServiceProvider.GetByUsers error",
			log.Any("op", op),
			log.Strings("userIDs", userIDs),
			log.Err(err))
		return nil, err
	}

	userClassMap, err := external.GetClassServiceProvider().GetByUserIDs(ctx, op, userIDs)
	if err != nil {
		log.Error(ctx, "GetClassServiceProvider.GetByUserIDs error",
			log.Any("op", op),
			log.Strings("userIDs", userIDs),
			log.Err(err))
		return nil, err
	}

	users := make([]*da.User, len(userInfos))
	for i, item := range userInfos {
		users[i] = &da.User{
			User:    *item,
			Schools: userSchoolMap[item.ID],
			Classes: userClassMap[item.ID],
		}
	}

	go da.GetUserRedisDA().SetUsers(ctx, op.OrgID, users)

	return users, nil
}

func (s *schedulePermissionModel) getOnlyUnderOrgUsers(users []*da.User) ([]*external.User, error) {
	result := make([]*external.User, 0)
	for _, user := range users {
		if len(user.Schools) > 0 || len(user.Classes) > 0 {
			continue
		}
		result = append(result, &user.User)
	}

	return result, nil
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
