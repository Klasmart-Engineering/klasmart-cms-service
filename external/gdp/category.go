package gdp

type CategoryFilter struct {
	Status *StringFilter    `json:"status,omitempty" gqls:"status,omitempty"`
	System *BooleanFilter   `json:"system,omitempty" gqls:"system,omitempty"`
	AND    []CategoryFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR     []CategoryFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (CategoryFilter) FilterType() FilterOfType {
	return CategoriesConnectionType
}

type CategoryConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}

type CategoriesConnectionEdge struct {
	Cursor string                 `json:"cursor" gqls:"cursor"`
	Node   CategoryConnectionNode `json:"node" gqls:"node"`
}

type CategoriesConnectionResponse struct {
	TotalCount int                        `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo         `json:"pageInfo" gqls:"pageInfo"`
	Edges      []CategoriesConnectionEdge `json:"edges" gqls:"edges"`
}

func (ccr CategoriesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &ccr.PageInfo
}
