package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) GetLearnerWeeklyReportOverview(ctx context.Context, op *entity.Operator, tr entity.TimeRange) (res entity.LearnerWeeklyReportOverview, err error) {
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.ReportLearningSummmaryOrg,
		external.ReportLearningSummarySchool,
		external.ReportLearningSummaryTeacher,
		external.ReportLearningSummaryStudent,
	})
	role, err := external.GetRoleServiceProvider().GetRole(ctx, op, entity.RoleStudent)
	if err != nil {
		return
	}
	cond := entity.GetUserCountCondition{
		OrgID: entity.NullString{
			String: op.OrgID,
			Valid:  true,
		},
		RoleID: entity.NullString{
			String: role.ID,
			Valid:  true,
		},
	}
	if perms[external.ReportLearningSummmaryOrg] {
		// query count of student in organization
		goto queryUserCount
	}

	if perms[external.ReportLearningSummarySchool] {
		// query count of student in schools of op
		cond.SchoolIDs.Valid = true
		var schools map[string][]*external.School
		schools, err = external.GetSchoolServiceProvider().GetByUsers(ctx, op, op.OrgID, []string{op.UserID})
		if err != nil {
			return
		}
		for _, s := range schools[op.UserID] {
			cond.SchoolIDs.Strings = append(cond.SchoolIDs.Strings, s.ID)
		}

		goto queryUserCount
	}
	if perms[external.ReportLearningSummaryTeacher] {
		// query count of in teacher’s classes
		cond.ClassIDs.Valid = true
		var classes []*external.Class
		classes, err = external.GetClassServiceProvider().GetByUserID(ctx, op, op.UserID)
		if err != nil {
			return
		}
		for _, class := range classes {
			cond.ClassIDs.Strings = append(cond.ClassIDs.Strings, class.ID)
		}
		goto queryUserCount
	}

	if perms[external.ReportLearningSummaryStudent] {
		// parent and student use the same user_id
		res.Attendees = 1
		cond.StudentID = entity.NullString{
			String: op.UserID,
			Valid:  true,
		}
		goto queryAttendance
	}

queryUserCount:
	res.Attendees, err = external.GetUserServiceProvider().GetUserCount(ctx, op, cond)
	if err != nil {
		return
	}

queryAttendance:
	r, err := da.GetReportDA().GetLearnerWeeklyReportOverview(ctx, op, tr, cond)
	if err != nil {
		return
	}
	res.BelowExpectation = r.BelowExpectation
	res.MeetExpectation = r.MeetExpectation
	res.NumAbove = r.NumAbove
	return
}
