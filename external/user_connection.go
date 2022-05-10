package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type UserFilter struct {
	UserID                 *UUIDFilter          `json:"userId,omitempty" gqls:"userId,omitempty"`
	UserStatus             *StringFilter        `json:"userStatus,omitempty" gqls:"userStatus,omitempty"`
	GivenName              *StringFilter        `json:"givenName,omitempty" gqls:"givenName,omitempty"`
	FamilyName             *StringFilter        `json:"familyName,omitempty" gqls:"familyName,omitempty"`
	Avatar                 *StringFilter        `json:"avatar,omitempty" gqls:"avatar,omitempty"`
	Email                  *StringFilter        `json:"email,omitempty" gqls:"email,omitempty"`
	Phone                  *StringFilter        `json:"phone,omitempty" gqls:"phone,omitempty"`
	UserName               *StringFilter        `json:"username,omitempty" gqls:"username,omitempty"`
	OrganizationID         *UUIDFilter          `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	RoleID                 *UUIDFilter          `json:"roleId,omitempty" gqls:"roleId,omitempty"`
	SchoolID               *UUIDExclusiveFilter `json:"schoolId,omitempty" gqls:"schoolId,omitempty"`
	OrganizationUserStatus *StringFilter        `json:"organizationUserStatus,omitempty" gqls:"organizationUserStatus,omitempty"`
	ClassID                *UUIDExclusiveFilter `json:"classId,omitempty" gqls:"classId,omitempty"`
	GradeID                *UUIDFilter          `json:"gradeId,omitempty" gqls:"gradeId,omitempty"`
	AND                    []UserFilter         `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR                     []UserFilter         `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (UserFilter) FilterName() FilterType {
	return UserFilterType
}

func (UserFilter) ConnectionName() ConnectionType {
	return UsersConnectionType
}

type AmsUserConnectionService struct {
	AmsUserService
}

type ContactInfo struct {
	Email    string `json:"email" gqls:"email"`
	Phone    string `json:"phone" gqls:"phone"`
	UserName string `json:"username" gqls:"username"`
}

type UserConnectionNode struct {
	ID          string      `json:"id" gqls:"id"`
	GivenName   string      `json:"givenName" gqls:"givenName"`
	FamilyName  string      `json:"familyName" gqls:"familyName"`
	UserName    string      `json:"username" gqls:"username"`
	Avatar      string      `json:"avatar" gqls:"avatar"`
	ContactInfo ContactInfo `json:"contactInfo" gqls:"contactInfo"`
	Status      APStatus    `json:"status" gqls:"status"`
}

type UsersConnectionEdge struct {
	Cursor string             `json:"cursor" gqls:"cursor"`
	Node   UserConnectionNode `json:"node" gqls:"node"`
}

type UsersConnectionResponse struct {
	TotalCount int                   `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo    `json:"pageInfo" gqls:"pageInfo"`
	Edges      []UsersConnectionEdge `json:"edges" gqls:"edges"`
}

func (ucs UsersConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &ucs.PageInfo
}
func (aucs AmsUserConnectionService) pageNodes(ctx context.Context, operator *entity.Operator, pages []UsersConnectionResponse) []*User {
	if len(pages) == 0 {
		log.Warn(ctx, "pageNodes is empty",
			log.Any("operator", operator))
		return []*User{}
	}
	users := make([]*User, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, edge := range page.Edges {
			if _, ok := exists[edge.Node.ID]; ok {
				log.Warn(ctx, "pageNodes: user exist",
					log.Any("users", edge.Node),
					log.Any("operator", operator))
				continue
			}
			exists[edge.Node.ID] = true
			user := User{
				ID:         edge.Node.ID,
				Name:       edge.Node.GivenName + " " + edge.Node.FamilyName,
				GivenName:  edge.Node.GivenName,
				FamilyName: edge.Node.FamilyName,
				Email:      edge.Node.ContactInfo.Email,
				Avatar:     edge.Node.Avatar,
			}
			users = append(users, &user)
		}
	}
	return users
}

func (aucs AmsUserConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, organizationID string) ([]*User, error) {
	filter := UserFilter{
		OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(organizationID)},
	}

	var pages []UsersConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get users by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}

	users := aucs.pageNodes(ctx, operator, pages)
	return users, nil
}

func (aucs AmsUserConnectionService) GetOnlyUnderOrgUsers(ctx context.Context, op *entity.Operator, orgID string) ([]*User, error) {
	filter := UserFilter{
		OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(op.OrgID)},
		SchoolID:       &UUIDExclusiveFilter{Operator: UUIDExclusiveOperator(OperatorTypeIsNull)},
		ClassID:        &UUIDExclusiveFilter{Operator: UUIDExclusiveOperator(OperatorTypeIsNull)},
	}

	var pages []UsersConnectionResponse
	err := pageQuery(ctx, op, filter, &pages)
	if err != nil {
		log.Error(ctx, "get only under org users failed",
			log.Err(err),
			log.Any("operator", op),
			log.Any("filter", filter))
		return nil, err
	}
	users := aucs.pageNodes(ctx, op, pages)
	return users, nil
}
