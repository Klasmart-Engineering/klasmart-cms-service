package connections

type OrganizationFilter struct {
	ID            *UUIDFilter          `json:"__id__,omitempty"`
	Name          *StringFilter        `json:"__name__,omitempty"`
	Phone         *StringFilter        `json:"__phone__,omitempty"`
	Shortcode     *StringFilter        `json:"__shortcode__,omitempty"`
	Status        *StringFilter        `json:"__status__,omitempty"`
	OwnerUserID   *StringFilter        `json:"__ownerUserId__,omitempty"`
	OwnerUseEmail *StringFilter        `json:"__ownerUseEmail__,omitempty"`
	UserID        *UUIDFilter          `json:"__userId__,omitempty"`
	AND           []OrganizationFilter `json:"__AND__,omitempty"`
	OR            []OrganizationFilter `json:"__OR__,omitempty"`
}

func (OrganizationFilter) FilterType() FilterOfType {
	return OrganizationsConnectionType
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
	TotalCount int                           `json:"totalCount"`
	PageInfo   ConnectionPageInfo            `json:"pageInfo"`
	Edges      []OrganizationsConnectionEdge `json:"edges"`
}

func (pcs OrganizationsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}
