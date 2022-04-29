package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type SchoolFilter struct {
	SchoolID       *UUIDFilter    `json:"schoolId,omitempty" gqls:"schoolId,omitempty"`
	Name           *StringFilter  `json:"name,omitempty" gqls:"name,omitempty"`
	Shortcode      *StringFilter  `json:"shortcode,omitempty" gqls:"shortcode,omitempty"`
	Status         *StringFilter  `json:"status,omitempty" gqls:"status,omitempty"`
	OrganizationId *UUIDFilter    `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	UserID         *UUIDFilter    `json:"userId,omitempty" gqls:"userId,omitempty"`
	ProgramId      *UUIDFilter    `json:"programId,omitempty" gqls:"programId,omitempty"`
	AND            []SchoolFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []SchoolFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (SchoolFilter) FilterName() FilterType {
	return SchoolFilterType
}

func (SchoolFilter) ConnectionName() ConnectionType {
	return SchoolsConnectionType
}

type AmsSchoolConnectionService struct {
	AmsSchoolService
}

type SchoolConnectionNode struct {
	ID             string `json:"id" gqls:"id"`
	Name           string `json:"name" gqls:"name"`
	ShortCode      string `json:"shortCode" gqls:"shortCode"`
	Status         string `json:"status" gqls:"status"`
	OrganizationId string `json:"organizationId" gqls:"organizationId"`
}

type SchoolsConnectionEdge struct {
	Cursor string               `json:"cursor" gqls:"cursor"`
	Node   SchoolConnectionNode `json:"node" gqls:"node"`
}

type SchoolsConnectionResponse struct {
	TotalCount int                     `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo      `json:"pageInfo" gqls:"pageInfo"`
	Edges      []SchoolsConnectionEdge `json:"edges" gqls:"edges"`
}

func (pcs SchoolsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}

func (ascs AmsSchoolConnectionService) GetByOrganizationID(ctx context.Context, operator *entity.Operator, organizationID string, options ...APOption) ([]*School, error) {
	condition := NewCondition(options...)
	filter := SchoolFilter{
		OrganizationId: &UUIDFilter{
			Operator: UUIDOperator(OperatorTypeEq),
			Value:    UUID(organizationID),
		},
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
	}

	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	var pages []SchoolsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get school by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	var schools []*School
	for _, page := range pages {
		for _, v := range page.Edges {
			sch := School{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
			}
			schools = append(schools, &sch)
		}
	}
	return schools, nil
}
func (ascs AmsSchoolConnectionService) GetByOperator(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*School, error) {
	condition := NewCondition(options...)
	filter := SchoolFilter{
		UserID: &UUIDFilter{
			Operator: UUIDOperator(OperatorTypeEq),
			Value:    UUID(operator.UserID),
		},
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
	}

	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	var pages []SchoolsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get school by user failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	var schools []*School
	for _, page := range pages {
		for _, v := range page.Edges {
			sch := School{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
			}
			schools = append(schools, &sch)
		}
	}
	return schools, nil
}

type SchoolMembershipConnectionNode struct {
	UserID        string               `json:"userId" gqls:"userId"`
	SchoolID      string               `json:"schoolId" gqls:"schoolId"`
	Status        string               `json:"status" gqls:"status"`
	JoinTimestamp string               `json:"joinTimestamp" gqls:"joinTimestamp"`
	User          UserConnectionNode   `json:"user" gqls:"user"`
	School        SchoolConnectionNode `json:"school" gqls:"school"`
}
type SchoolMembershipsConnectionEdge struct {
	Cursor string                         `json:"cursor" gqls:"cursor"`
	Node   SchoolMembershipConnectionNode `json:"node" gqls:"node"`
}
type SchoolMembershipsConnectionResponse struct {
	TotalCount int                               `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo                `json:"pageInfo" gqls:"pageInfo"`
	Edges      []SchoolMembershipsConnectionEdge `json:"edges" gqls:"edges"`
}

func (pcs SchoolMembershipsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}

func (ascs AmsSchoolConnectionService) GetByUsers(ctx context.Context, operator *entity.Operator, orgID string, userIDs []string, options ...APOption) (map[string][]*School, error) {
	result := make(map[string][]SchoolMembershipsConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(userIDs)
	err := subPageQuery(ctx, operator, "userNode", "schoolMembershipsConnection", IDs, result)
	if err != nil {
		return nil, err
	}
	condition := NewCondition(options...)
	schoolsMap := make(map[string][]*School)
	for k, pages := range result {
		for _, page := range pages {
			for _, edge := range page.Edges {
				node := edge.Node.School
				if node.OrganizationId != orgID {
					continue
				}

				if condition.Status.Valid && condition.Status.Status != APStatus(node.Status) {
					continue
				} else if !condition.Status.Valid && APStatus(node.Status) != Active {
					// only status = "Active" data is returned by default
					continue
				}
				school := &School{
					ID:     node.ID,
					Name:   node.Name,
					Status: APStatus(node.Status),
				}
				schoolsMap[k] = append(schoolsMap[k], school)
			}
		}
	}
	return schoolsMap, nil
}
