package gql

type GradeFilter struct {
	ID             *UUIDFilter    `gqls:"id,omitempty"`
	Name           *StringFilter  `gqls:"name,omitempty"`
	Status         *StringFilter  `gqls:"status,omitempty"`
	System         *BooleanFilter `gqls:"system,omitempty"`
	OrganizationID *UUIDFilter    `gqls:"organizationId,omitempty"`
	CategoryID     *UUIDFilter    `gqls:"categoryId,omitempty"`
	ClassID        *UUIDFilter    `gqls:"classId,omitempty"`
	ProgramID      *UUIDFilter    `gqls:"programId,omitempty"`
	FromGradeID    *UUIDFilter    `gqls:"fromGradeId,omitempty"`
	ToGradeID      *UUIDFilter    `gqls:"toGradeId,omitempty"`
	AND            []GradeFilter  `gqls:"AND,omitempty"`
	OR             []GradeFilter  `gqls:"OR,omitempty"`
}

func (GradeFilter) FilterType() FilterOfType {
	return GradesConnectionType
}

type GradeSummaryNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}
type GradeConnectionNode struct {
	ID        string           `json:"id" gqls:"id"`
	Name      string           `json:"name" gqls:"id"`
	Status    string           `json:"status" gqls:"status"`
	System    bool             `json:"system" gqls:"system"`
	FromGrade GradeSummaryNode `json:"fromGrade" gqls:"fromGrade"`
	ToGrade   GradeSummaryNode `json:"toGrade" gqls:"toGrade"`
}

type GradesConnectionEdge struct {
	Cursor string              `json:"cursor" gqls:"cursor"`
	Node   GradeConnectionNode `json:"node" gqls:"node"`
}

type GradesConnectionResponse struct {
	TotalCount int                    `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo     `json:"pageInfo" gqls:"pageInfo"`
	Edges      []GradesConnectionEdge `json:"edges" gqls:"edges"`
}

func (scs GradesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &scs.PageInfo
}
