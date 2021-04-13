package entity

type ClassEventBody struct {
	Token string `json:"token"`
}

type ClassUpdateMembersEvent struct {
	ClassID string         `json:"class_id"`
	Members []*ClassMember `json:"members"`
}

func (s ClassUpdateMembersEvent) Valid() error {
	return nil
}

type ClassMember struct {
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
