package entity

type Operator struct {
	UserID string `json:"user_id"`
	OrgID  string `json:"org_id"`
	Token  string `json:"-"` // hide token in log
}
