package gqp

type ProgramFilter struct {
	ID             *UUIDFilter         `gqls:"id,omitempty"`
	Name           *StringFilter       `gqls:"name,omitempty"`
	Status         *StringFilter       `gqls:"status,omitempty"`
	System         *BooleanFilter      `gqls:"system,omitempty"`
	OrganizationID *UUIDFilter         `gqls:"organizationId,omitempty"`
	GradeID        *UUIDFilter         `gqls:"gradeId,omitempty"`
	AgeRangeFrom   *AgeRangeTypeFilter `gqls:"ageRangeFrom,omitempty"`
	AgeRangeTo     *AgeRangeTypeFilter `gqls:"ageRangeTo,omitempty"`
	SubjectID      *UUIDFilter         `gqls:"subjectId,omitempty"`
	SchoolID       *UUIDFilter         `gqls:"schoolId,omitempty"`
	ClassID        *UUIDFilter         `gqls:"classId,omitempty"`
	AND            []ProgramFilter     `gqls:"AND,omitempty"`
	OR             []ProgramFilter     `gqls:"OR,omitempty"`
}

func (ProgramFilter) FilterType() FilterOfType {
	return ProgramsConnectionType
}

type ProgramConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}
type ProgramsConnectionEdge struct {
	Cursor string                `json:"cursor" gqls:"cursor"`
	Node   ProgramConnectionNode `json:"node" gqls:"node"`
}

type ProgramsConnectionResponse struct {
	TotalCount int                      `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo       `json:"pageInfo" gqls:"pageInfo"`
	Edges      []ProgramsConnectionEdge `json:"edges" gqls:"edges"`
}

func (pcs ProgramsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}
