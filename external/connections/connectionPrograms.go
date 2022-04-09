package connections

type ProgramFilter struct {
	ID             *UUIDFilter         `json:"__id__,omitempty"`
	Name           *StringFilter       `json:"__name__,omitempty"`
	Status         *StringFilter       `json:"__status__,omitempty"`
	System         *BooleanFilter      `json:"__system__,omitempty"`
	OrganizationID *UUIDFilter         `json:"__organizationId__,omitempty"`
	GradeID        *UUIDFilter         `json:"__gradeId__,omitempty"`
	AgeRangeFrom   *AgeRangeTypeFilter `json:"__ageRangeFrom__,omitempty"`
	AgeRangeTo     *AgeRangeTypeFilter `json:"__ageRangeTo__,omitempty"`
	SubjectID      *UUIDFilter         `json:"__subjectId__,omitempty"`
	SchoolID       *UUIDFilter         `json:"__schoolId,omitempty"`
	ClassID        *UUIDFilter         `json:"__classId__,omitempty"`
	AND            []ProgramFilter     `json:"__AND__,omitempty"`
	OR             []ProgramFilter     `json:"__OR__,omitempty"`
}

func (ProgramFilter) FilterType() FilterOfType {
	return ProgramsConnectionType
}

type ProgramConnectionNode struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	System bool   `json:"system"`
}
type ProgramsConnectionEdge struct {
	Cursor string                `json:"cursor"`
	Node   ProgramConnectionNode `json:"node"`
}

type ProgramsConnectionResponse struct {
	TotalCount int                      `json:"totalCount"`
	PageInfo   ConnectionPageInfo       `json:"pageInfo"`
	Edges      []ProgramsConnectionEdge `json:"edges"`
}

func (pcs ProgramsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}
