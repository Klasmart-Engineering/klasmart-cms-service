package entity

type ScheduleEventBody struct {
	Token string `json:"token"`
}

type ScheduleClassEvent struct {
	ClassID string                    `json:"class_id"`
	Users   []*ScheduleClassUserEvent `json:"users"`
}

func (s ScheduleClassEvent) Valid() error {
	return nil
}

type ScheduleClassUserEvent struct {
	ID       string                 `json:"id"`
	RoleType ClassUserRoleTypeEvent `json:"role_type" enums:"Student,Teacher"`
}

type ClassUserRoleTypeEvent string

const (
	ClassUserRoleTypeEventTeacher ClassUserRoleTypeEvent = "Teacher"
	ClassUserRoleTypeEventStudent ClassUserRoleTypeEvent = "Student"
)

func (t ClassUserRoleTypeEvent) ToScheduleRelationType() ScheduleRelationType {
	switch t {
	case ClassUserRoleTypeEventTeacher:
		return ScheduleRelationTypeClassRosterTeacher
	case ClassUserRoleTypeEventStudent:
		return ScheduleRelationTypeClassRosterStudent
	default:
		return ScheduleRelationTypeInvalid
	}
}

type ScheduleClassEventAction string

const (
	ScheduleClassEventActionAdd    ScheduleClassEventAction = "Add"
	ScheduleClassEventActionDelete ScheduleClassEventAction = "Delete"
)
