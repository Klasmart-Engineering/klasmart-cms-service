package entity

type Program struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Number    int              `json:"number"`
	OrgType   string           `json:"org_type"`
	GroupName ProgramGroupName `json:"group_name"`
	CreateID  string
	DeleteID  string
	CreateAt  int64
	UpdateAt  int64
	DeleteAt  int64
}
