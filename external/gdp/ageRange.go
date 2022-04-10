package gdp

type AgeRangeFilter struct {
	AgeRangeValueFrom AgeRangeValueFilter `json:"ageRangeValueFrom,omitempty" gqls:"ageRangeValueFrom,omitempty"`
	AgeRangeUnitFrom  AgeRangeUnitFilter  `json:"ageRangeUnitFrom,omitempty" gqls:"ageRangeUnitFrom,omitempty"`
	AgeRangeValueTo   AgeRangeValueFilter `json:"ageRangeValueTo,omitempty" gqls:"ageRangeValueTo,omitempty"`
	AgeRangeUnitTo    AgeRangeUnitFilter  `json:"ageRangeUnitTo,omitempty" gqls:"ageRangeUnitTo,omitempty"`
	Status            *StringFilter       `json:"status,omitempty" gqls:"status,omitempty"`
	System            *BooleanFilter      `json:"system,omitempty" gqls:"system,omitempty"`
	OrganizationID    *UUIDFilter         `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	ProgramID         *UUIDFilter         `json:"programId,omitempty" gqls:"programId,omitempty"`
	AND               []AgeRangeFilter    `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR                []AgeRangeFilter    `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (AgeRangeFilter) FilterType() FilterOfType {
	return AgeRangesConnectionType
}

type AgeRangeConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}

type AgeRangesConnectionEdge struct {
	Cursor string                 `json:"cursor" gqls:"cursor"`
	Node   AgeRangeConnectionNode `json:"node" gqls:"node"`
}

type AgesConnectionResponse struct {
	TotalCount int                       `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo        `json:"pageInfo" gqls:"pageInfo"`
	Edges      []AgeRangesConnectionEdge `json:"edges" gqls:"edges"`
}

func (acs AgesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &acs.PageInfo
}
