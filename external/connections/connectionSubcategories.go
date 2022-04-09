package connections

type SubcategoryFilter struct {
	OrganizationID UUIDFilter          `json:"__organizationId__"`
	CategoryID     UUIDFilter          `json:"__categoryId__"`
	System         BooleanFilter       `json:"__system__"`
	Status         StringFilter        `json:"status"`
	AND            []SubcategoryFilter `json:"__AND__,omitempty"`
	OR             []SubcategoryFilter `json:"__OR__,omitempty"`
}

func (SubcategoryFilter) FilterType() FilterOfType {
	return SubcategoriesConnectionType
}

type SubcategoryConnectionNode struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	System bool   `json:"system"`
}

type SubcategoriesConnectionEdge struct {
	Cursor string                    `json:"cursor"`
	Node   SubcategoryConnectionNode `json:"node"`
}

type SubcategoriesConnectionResponse struct {
	TotalCount int                           `json:"totalCount"`
	PageInfo   ConnectionPageInfo            `json:"pageInfo"`
	Edges      []SubcategoriesConnectionEdge `json:"edges"`
}

func (scs SubcategoriesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &scs.PageInfo
}
