package gdp

type SubjectFilter struct {
	ID             *UUIDFilter     `json:"id,omitempty" gqls:"id,omitempty"`
	Name           *StringFilter   `json:"name,omitempty" gqls:"name,omitempty"`
	Status         *StringFilter   `json:"status,omitempty" gqls:"status,omitempty"`
	System         *BooleanFilter  `json:"system,omitempty" gqls:"system,omitempty"`
	OrganizationID *UUIDFilter     `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	CategoryId     *UUIDFilter     `json:"categoryId,omitempty" gqls:"categoryId,omitempty"`
	ClassID        *UUIDFilter     `json:"classId,omitempty" gqls:"classId,omitempty"`
	ProgramID      *UUIDFilter     `json:"programId,omitempty" gqls:"programId,omitempty"`
	AND            []SubjectFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []SubjectFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (SubjectFilter) FilterType() FilterOfType {
	return CategoriesConnectionType
}

type SubjectConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}

type SubjectsConnectionEdge struct {
	Cursor string                `json:"cursor" gqls:"cursor"`
	Node   SubjectConnectionNode `json:"node" gqls:"node"`
}

type SubjectsConnectionResponse struct {
	TotalCount int                      `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo       `json:"pageInfo" gqls:"pageInfo"`
	Edges      []SubjectsConnectionEdge `json:"edges" gqls:"edges"`
}

func (scr SubjectsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &scr.PageInfo
}
