package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type RoleFilter struct {
	Name                         *StringFilter  `json:"name,omitempty" gqls:"name,omitempty"`
	Status                       *StringFilter  `json:"status,omitempty" gqls:"status,omitempty"`
	System                       *BooleanFilter `json:"system,omitempty" gqls:"system,omitempty"`
	OrganizationID               *UUIDFilter    `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	SchoolID                     *UUIDFilter    `json:"schoolId,omitempty" gqls:"schoolId,omitempty"`
	SchoolUserID                 *UUIDFilter    `json:"schoolUserId,omitempty" gqls:"schoolUserId,omitempty"`
	MembershipOrganizationUserID *UUIDFilter    `json:"membershipOrganizationUserId,omitempty" gqls:"membershipOrganizationUserId,omitempty"`
	MembershipOrganizationID     *UUIDFilter    `json:"membershipOrganizationId,omitempty" gqls:"membershipOrganizationId,omitempty"`
	AND                          []RoleFilter   `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR                           []RoleFilter   `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (RoleFilter) FilterName() FilterType {
	return RoleFilterType
}

func (RoleFilter) ConnectionName() ConnectionType {
	return RolesConnectionType
}

type AmsRoleConnectionService struct {
	AmsRoleService
}

type RoleConnectionNode struct {
	ID          string `json:"id" gqls:"id"`
	Name        string `json:"name" gqls:"name"`
	Description string `json:"description" gqls:"description"`
	Status      string `json:"status" gqls:"status"`
	System      bool   `json:"system" gqls:"system"`
}

type RolesConnectionEdge struct {
	Cursor string             `json:"cursor" gqls:"cursor"`
	Node   RoleConnectionNode `json:"node" gqls:"node"`
}

type RolesConnectionResponse struct {
	TotalCount int                   `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo    `json:"pageInfo" gqls:"pageInfo"`
	Edges      []RolesConnectionEdge `json:"edges" gqls:"edges"`
}

func (pcs RolesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}

func (arcs *AmsRoleConnectionService) GetRole(ctx context.Context, op *entity.Operator, roleName entity.RoleName) (*entity.Role, error) {
	filter := RoleFilter{
		Name: &StringFilter{Operator: StringOperator(OperatorTypeEq), Value: string(roleName)},
	}

	var pages []AgesConnectionResponse
	err := pageQuery(ctx, op, filter, &pages)
	if err != nil {
		log.Error(ctx, "get role by name failed",
			log.Err(err),
			log.Any("operator", op),
			log.Any("filter", filter))
		return nil, err
	}
	var roles []*entity.Role
	for _, page := range pages {
		for _, v := range page.Edges {
			role := entity.Role{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: v.Node.Status,
			}
			roles = append(roles, &role)
		}
	}
	if len(roles) != 1 {
		log.Error(ctx, "role name is not unique",
			log.Err(err),
			log.Any("roles", roles),
			log.Any("operator", op),
			log.Any("filter", filter))
	}
	return roles[0], nil
}
