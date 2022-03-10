package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) GetLearnerReportOverview(ctx context.Context, op *entity.Operator, cond entity.LearnerReportOverviewCondition) (res entity.LearnerReportOverview, err error) {
	permOrg := external.PermissionName(cond.PermOrg)
	permSchool := external.PermissionName(cond.PermSchool)
	permTeacher := external.PermissionName(cond.PermTeacher)
	permStudent := external.PermissionName(cond.PermStudent)

	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		permOrg,
		permSchool,
		permTeacher,
		permStudent,
	})
	role, err := external.GetRoleServiceProvider().GetRole(ctx, op, entity.RoleStudent)
	if err != nil {
		return
	}
	ucCond := entity.GetUserCountCondition{
		OrgID: entity.NullString{
			String: op.OrgID,
			Valid:  true,
		},
		RoleID: entity.NullString{
			String: role.ID,
			Valid:  true,
		},
	}
	if perms[permOrg] {
		// query count of student in organization
		goto queryUserCount
	}

	if perms[permSchool] {
		// query count of student in schools of op
		ucCond.SchoolIDs.Valid = true
		var schools map[string][]*external.School
		schools, err = external.GetSchoolServiceProvider().GetByUsers(ctx, op, op.OrgID, []string{op.UserID})
		if err != nil {
			return
		}
		for _, s := range schools[op.UserID] {
			ucCond.SchoolIDs.Strings = append(ucCond.SchoolIDs.Strings, s.ID)
		}

		goto queryUserCount
	}
	if perms[permTeacher] {
		// query count of in teacher’s classes
		ucCond.ClassIDs.Valid = true
		var classes []*external.Class
		classes, err = external.GetClassServiceProvider().GetByUserID(ctx, op, op.UserID)
		if err != nil {
			return
		}
		for _, class := range classes {
			ucCond.ClassIDs.Strings = append(ucCond.ClassIDs.Strings, class.ID)
		}
		goto queryUserCount
	}

	if perms[permStudent] {
		// parent and student use the same user_id
		res.Attendees = 1
		ucCond.StudentID = entity.NullString{
			String: op.UserID,
			Valid:  true,
		}
		goto queryAttendance
	}

	return
queryUserCount:
	res.Attendees, err = external.GetUserServiceProvider().GetUserCount(ctx, op, ucCond)
	if err != nil {
		return
	}

queryAttendance:
	r, err := da.GetReportDA().GetLearnerWeeklyReportOverview(ctx, op, cond.TimeRange, ucCond)
	if err != nil {
		return
	}
	res.NumBelow = r.NumBelow
	res.NumMeet = r.NumMeet
	res.NumAbove = r.NumAbove
	return
}
