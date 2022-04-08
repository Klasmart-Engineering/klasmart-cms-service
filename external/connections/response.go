package connections

type ConnectionPageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

func (pageInfo ConnectionPageInfo) Pager(direction ConnectionDirection, count int) Pager {
	if direction == FORWARD && pageInfo.HasNextPage {
		return Pager{
			PagerDirection: direction,
			PagerCursor:    pageInfo.EndCursor,
			PagerCount:     count,
		}
	}
	if direction == BACKWARD && pageInfo.HasPreviousPage {
		return Pager{
			PagerDirection: direction,
			PagerCursor:    pageInfo.StartCursor,
			PagerCount:     count,
		}
	}
	return nil
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

type IConnectionResponse struct{}

type ProgramsConnectionResponse struct {
	TotalCount int                      `json:"totalCount"`
	PageInfo   ConnectionPageInfo       `json:"pageInfo"`
	Edges      []ProgramsConnectionEdge `json:"edges"`
}

type SubcategoryConnectionNode struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	System bool   `json:"system"`
}

type OrganizationConnectionNode struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ContactInfo string `json:"contactInfo"`
	ShortCode   string `json:"shortCode"`
	Status      string `json:"status"`
	Owners      []int  `json:"owners"`
	Branding    string `json:"branding"`
}

type OrganizationsConnectionEdge struct {
	Cursor string                     `json:"cursor"`
	Node   OrganizationConnectionNode `json:"node"`
}

type OrganizationsConnectionResponse struct {
	TotalCount int                      `json:"totalCount"`
	PageInfo   ConnectionPageInfo       `json:"pageInfo"`
	Edges      []ProgramsConnectionEdge `json:"edges"`
}
