package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type OrganizationFilter struct {
	ID            *UUIDFilter          `json:"id,omitempty" gqls:"id,omitempty"`
	Name          *StringFilter        `json:"name,omitempty" gqls:"name,omitempty"`
	Phone         *StringFilter        `json:"phone,omitempty" gqls:"phone,omitempty"`
	Shortcode     *StringFilter        `json:"shortcode,omitempty" gqls:"shortcode,omitempty"`
	Status        *StringFilter        `json:"status,omitempty" gqls:"status,omitempty"`
	OwnerUserID   *UUIDFilter          `json:"ownerUserId,omitempty" gqls:"ownerUserId,omitempty"`
	OwnerUseEmail *StringFilter        `json:"ownerUseEmail,omitempty" gqls:"ownerUseEmail,omitempty"`
	OwnerUseName  *StringFilter        `json:"ownerUseName,omitempty" gqls:"ownerUseName,omitempty"`
	UserID        *UUIDFilter          `json:"userId,omitempty" gqls:"userId,omitempty"`
	AND           []OrganizationFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR            []OrganizationFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (OrganizationFilter) FilterType() FilterOfType {
	return OrganizationsConnectionType
}

type AmsOrganizationConnectionService struct {
	AmsOrganizationService
}

type OrganizationContactInfo struct {
	Address1 string `json:"address1" gqls:"address1"`
	Address2 string `json:"address2" gqls:"address2"`
	Phone    string `json:"phone" gqls:"phone"`
}

type UserSummaryNode struct {
	ID    string `json:"id" gqls:"id"`
	Email string `json:"email" gqls:"email"`
}

type OrganizationConnectionNode struct {
	ID          string                  `json:"id" gqls:"id"`
	Name        string                  `json:"name" gqls:"name"`
	ContactInfo OrganizationContactInfo `json:"contactInfo" gqls:"contactInfo"`
	ShortCode   string                  `json:"shortCode" gqls:"shortCode"`
	Status      string                  `json:"status" gqls:"status"`
	Owners      []UserSummaryNode       `json:"owners" gqls:"owners"`
	Branding    Branding                `json:"branding" gqls:"branding"`
}

type OrganizationsConnectionEdge struct {
	Cursor string                     `json:"cursor" gqls:"cursor"`
	Node   OrganizationConnectionNode `json:"node" gqls:"node"`
}

type OrganizationsConnectionResponse struct {
	TotalCount int                           `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo            `json:"pageInfo" gqls:"pageInfo"`
	Edges      []OrganizationsConnectionEdge `json:"edges" gqls:"edges"`
}

func (pcs OrganizationsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}

func (aocs AmsOrganizationConnectionService) GetByUserID(ctx context.Context, operator *entity.Operator, id string, options ...APOption) ([]*Organization, error) {
	panic("implement me")
}
