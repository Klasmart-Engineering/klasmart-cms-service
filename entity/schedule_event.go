package entity

type ScheduleClassEvent struct {
	Action  ClassActionEvent  `json:"action" enums:"Add,Delete"`
	ClassID string            `json:"class_id"`
	Users   []*ClassUserEvent `json:"users"`
}
type ClassUserEvent struct {
	ID       string                 `json:"id"`
	RoleType ClassUserRoleTypeEvent `json:"role_type" enums:"Student,Teacher"`
}
type ClassUserRoleTypeEvent string

const (
	ClassUserRoleTypeEventTeacher ClassUserRoleTypeEvent = "Teacher"
	ClassUserRoleTypeEventStudent ClassUserRoleTypeEvent = "Student"
)

type ClassActionEvent string

const (
	ClassActionEventAdd    ClassActionEvent = "Add"
	ClassActionEventDelete ClassActionEvent = "Delete"
)
