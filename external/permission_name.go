package external

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type PermissionName string

func (p PermissionName) String() string {
	return string(p)
}

func (p PermissionName) Allow(ctx context.Context, operator *entity.Operator) (bool, error) {
	return GetPermissionServiceProvider().HasPermission(ctx, operator, p)
}

const (
	CreateContentPage201 PermissionName = "create_content_page_201"

	ScheduleViewOrgCalendar    PermissionName = "view_org_calendar__511"
	ScheduleViewSchoolCalendar PermissionName = "view_school_calendar_512"
	ScheduleViewMyCalendar     PermissionName = "view_my_calendar_510"
	ScheduleCreateEvent        PermissionName = "create_event__520"
	ScheduleEditEvent          PermissionName = "edit_event__530"
	ScheduleDeleteEvent        PermissionName = "delete_event_540"
	LiveClassTeacher           PermissionName = "attend_live_class_as_a_teacher_186"
	LiveClassStudent           PermissionName = "attend_live_class_as_a_student_187"
)
