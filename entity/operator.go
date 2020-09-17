package entity

import "fmt"

const (
	RoleTeacher = "teacher"
	RoleAdmin   = "admin"
	RoleStudent = "student"
)

type Operator struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	OrgID  string `json:"org_id"`
}

func (op *Operator) ToLivePayload() string {
	return fmt.Sprintf("%s,%s,%s", op.UserID, op.Role, op.OrgID)
}
