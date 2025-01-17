package external

import (
	"context"
	"fmt"
	"strings"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cache/cache"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
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

type SchoolMembershipFilter struct {
	UserID   *UUIDFilter    `json:"userId,omitempty" gqls:"userId,omitempty"`
	SchoolID *UUIDFilter    `json:"schoolId,omitempty" gqls:"schoolId,omitempty"`
	Status   *StringFilter  `json:"status,omitempty" gqls:"status,omitempty"`
	RoleID   *UUIDFilter    `json:"roleId,omitempty" gqls:"roleId,omitempty"`
	AND      []SchoolFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR       []SchoolFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (SchoolMembershipFilter) FilterName() FilterType {
	return SchoolMembershipFilterType
}

func (SchoolMembershipFilter) ConnectionName() ConnectionType {
	return SchoolMembershipsConnectionType
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
	if len(pages) == 0 {
		log.Warn(ctx, "organization is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*School{}, nil
	}
	schools := make([]*School, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, v := range page.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "organization exists",
					log.Any("organization", v.Node),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
			sch := School{
				ID:             v.Node.ID,
				Name:           v.Node.Name,
				Status:         APStatus(v.Node.Status),
				OrganizationId: v.Node.OrganizationId,
			}
			schools = append(schools, &sch)
		}
	}
	return schools, nil
}

func (ascs AmsSchoolConnectionService) GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName, options ...APOption) ([]*School, error) {
	schools, err := ascs.GetByOperator(ctx, operator)
	if err != nil {
		log.Error(ctx, "GetByPermission: GetByOperator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("permission", string(permissionName)))
		return nil, err
	}
	permissions, err := GetPermissionServiceProvider().HasOrganizationPermissions(ctx, operator, []PermissionName{permissionName})
	if err != nil {
		log.Error(ctx, "GetByPermission: HasOrganizationPermissions failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("permission", string(permissionName)))
		return nil, err
	}
	if !permissions[permissionName] {
		log.Debug(ctx, "GetByPermission: Has no permission",
			log.Any("operator", operator),
			log.Any("schools", schools),
			log.String("permission", string(permissionName)))
		return []*School{}, err
	}
	return schools, nil
}

func (ascs AmsSchoolConnectionService) GetByOperator(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*School, error) {
	condition := NewCondition(options...)
	filter := SchoolFilter{
		OrganizationId: &UUIDFilter{
			Operator: UUIDOperator(OperatorTypeEq),
			Value:    UUID(operator.OrgID),
		},
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
	if len(pages) == 0 {
		log.Warn(ctx, "organization is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*School{}, nil
	}
	schools := make([]*School, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, v := range page.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "organization exists",
					log.Any("organization", v.Node),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
			sch := School{
				ID:             v.Node.ID,
				Name:           v.Node.Name,
				OrganizationId: v.Node.OrganizationId,
				Status:         APStatus(v.Node.Status),
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
	condition := NewCondition(options...)
	var filter SchoolMembershipFilter

	//if condition.Status.Valid && condition.Status.Status != Ignore {
	//	filter.Status.Operator = StringOperator(OperatorTypeEq)
	//	filter.Status.Value = condition.Status.Status.String()
	//} else if !condition.Status.Valid {
	//	filter.Status.Operator = StringOperator(OperatorTypeEq)
	//	filter.Status.Value = Active.String()
	//}

	result := make(map[string][]SchoolMembershipsConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(userIDs)
	err := subPageQuery(ctx, operator, "userNode", filter, IDs, result)
	if err != nil {
		log.Error(ctx, "GetByUsers: subPageQuery failed",
			log.Err(err),
			log.String("org_id", orgID),
			log.Strings("user_ids", userIDs))
		return nil, err
	}
	schoolsMap := make(map[string][]*School)
	for k, pages := range result {
		for _, page := range pages {
			if len(page.Edges) == 0 {
				log.Warn(ctx, "GetByUsers: school is empty",
					log.String("user", k),
					log.String("org_id", orgID),
					log.Strings("user_ids", userIDs))
				schoolsMap[k] = []*School{}
				continue
			}
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
					ID:             node.ID,
					Name:           node.Name,
					Status:         APStatus(node.Status),
					OrganizationId: node.OrganizationId,
				}
				schoolsMap[k] = append(schoolsMap[k], school)
			}
		}
	}
	return schoolsMap, nil
}

func (ascs AmsSchoolConnectionService) pageNodes(ctx context.Context, operator *entity.Operator, pages []SchoolsConnectionResponse) []*School {
	if len(pages) == 0 {
		log.Warn(ctx, "pageNodes is empty",
			log.Any("operator", operator))
		return []*School{}
	}
	schools := make([]*School, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, edge := range page.Edges {
			if _, ok := exists[edge.Node.ID]; ok {
				log.Warn(ctx, "pageNodes: school exist",
					log.Any("school", edge.Node),
					log.Any("operator", operator))
				continue
			}
			exists[edge.Node.ID] = true
			school := &School{
				ID:             edge.Node.ID,
				Name:           edge.Node.Name,
				Status:         APStatus(edge.Node.Status),
				OrganizationId: edge.Node.OrganizationId,
			}
			schools = append(schools, school)
		}
	}
	return schools
}

func (ascs AmsSchoolConnectionService) NewSchoolFilter(ctx context.Context, operator *entity.Operator, options ...APOption) *SchoolFilter {
	condition := NewCondition(options...)
	var filter SchoolFilter
	if condition.Status.Valid && condition.Status.Status != Ignore {
		filter.Status = &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    condition.Status.Status.String(),
		}
	} else if !condition.Status.Valid {
		filter.Status = &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		}
	}
	return &filter
}

func (ascs AmsSchoolConnectionService) GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string, options ...APOption) (map[string][]*School, error) {
	filter := ascs.NewSchoolFilter(ctx, operator, options...)
	result := make(map[string][]SchoolsConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(classIDs)
	err := subPageQuery(ctx, operator, "classNode", filter, IDs, result)
	if err != nil {
		log.Error(ctx, "GetByClasses: subPageQuery failed",
			log.Err(err),
			log.Strings("class_ids", classIDs))
		return nil, err
	}
	schoolsMap := make(map[string][]*School)
	for k, pages := range result {
		log.Warn(ctx, "GetByClasses: school is empty", log.String("class", k))
		schoolsMap[k] = ascs.pageNodes(ctx, operator, pages)
	}
	return schoolsMap, nil
}

func (ascs AmsSchoolConnectionService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$school_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: schoolNode(id: $school_id_%d) {id name status}\n", index, index)
	}
	sb.WriteString("}")

	request := NewRequest(sb.String(), RequestToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("school_id_%d", index), id)
	}

	data := map[string]*School{}

	response := &GraphQLSubResponse{
		Data: &data,
	}

	_, err = GetAmsConnection().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get schools by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	schools := make([]cache.Object, 0, len(data))
	for index := range ids {
		school := data[fmt.Sprintf("q%d", indexMapping[index])]
		schools = append(schools, &NullableSchool{
			Valid:  school != nil,
			School: school,
			StrID:  ids[index],
		})
	}

	log.Info(ctx, "get schools by ids success",
		log.Strings("ids", ids),
		log.Any("schools", schools))

	return schools, nil
}
