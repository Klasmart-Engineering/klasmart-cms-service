package external

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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

func (OrganizationFilter) FilterName() FilterType {
	return OrganizationFilterType
}

func (OrganizationFilter) ConnectionName() ConnectionType {
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

func (aocs AmsOrganizationConnectionService) GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string, options ...APOption) (map[string]*Organization, error) {
	if len(classIDs) == 0 {
		log.Warn(ctx, "GetByClasses: class is empty",
			log.Any("operator", operator),
			log.Strings("class_ids", classIDs))
		return map[string]*Organization{}, nil
	}
	schools, err := GetSchoolServiceProvider().GetByClasses(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "GetByClasses: get school failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("class_ids", classIDs))
		return nil, err
	}
	if len(schools) == 0 {
		log.Warn(ctx, "GetByClasses: school is empty",
			log.Any("operator", operator),
			log.Strings("class_ids", classIDs))
		return map[string]*Organization{}, nil
	}
	condition := NewCondition(options...)
	filter := OrganizationFilter{
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
	}

	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}

	clsOrgIDMap := make(map[string]string)
	for k, schs := range schools {
		for _, sch := range schs {
			if org, ok := clsOrgIDMap[k]; ok && org != sch.OrganizationId {
				log.Error(ctx, "GetByClasses: one class maps different org",
					log.Any("operator", operator),
					log.Any("class_org_map", clsOrgIDMap),
					log.Any("map_key", k),
					log.Any("school", sch),
					log.Strings("class_ids", classIDs))
				return nil, constant.ErrDuplicateRecord
			}
			clsOrgIDMap[k] = sch.OrganizationId

			filter.OR = append(filter.OR, OrganizationFilter{
				ID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(sch.OrganizationId)},
			})
		}
	}

	var pages []OrganizationsConnectionResponse
	err = pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "GetByClasses: pageQuery failed",
			log.Err(err),
			log.Strings("class_ids", classIDs),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}

	if len(pages) == 0 {
		log.Warn(ctx, "GetByClasses: organization is empty",
			log.Any("school", schools),
			log.Any("filter", filter),
			log.Any("operator", operator),
			log.Strings("class_ids", classIDs))
		return map[string]*Organization{}, nil
	}

	orgMap := make(map[string]*Organization)
	for _, page := range pages {
		for _, v := range page.Edges {
			org := Organization{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
			}
			if _, ok := orgMap[v.Node.ID]; !ok {
				orgMap[v.Node.ID] = &org
			}
		}
	}

	classOrganizationMap := make(map[string]*Organization)
	for k, v := range clsOrgIDMap {
		if org, ok := orgMap[v]; ok {
			classOrganizationMap[k] = org
		}
	}
	return classOrganizationMap, nil
}
func (aocs AmsOrganizationConnectionService) GetNameByOrganizationOrSchool(ctx context.Context, operator *entity.Operator, id []string) ([]string, error) {
	nameMap, err := aocs.GetNameMapByOrganizationOrSchool(ctx, operator, id)
	if err != nil {
		log.Error(ctx, "GetNameByOrganizationOrSchool: get name map  failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("ids", id))
		return nil, err
	}
	names := make([]string, len(id))
	for i, v := range id {
		names[i] = nameMap[v]
	}
	return names, nil
}
func (aocs AmsOrganizationConnectionService) GetNameMapByOrganizationOrSchool(ctx context.Context, operator *entity.Operator, id []string) (map[string]string, error) {
	IDs := utils.SliceDeduplicationExcludeEmpty(id)
	if len(IDs) == 0 {
		log.Debug(ctx, "GetNameMapByOrganizationOrSchool: id is empty",
			log.Any("operator", operator),
			log.Strings("ids", IDs))
		return nil, nil
	}

	nameMap := make(map[string]string)
	var mapLock sync.RWMutex
	var orgErr, schErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		filter := OrganizationFilter{}
		for _, oid := range IDs {
			filter.OR = append(filter.OR, OrganizationFilter{
				ID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(oid)},
			})
		}
		var pages []OrganizationsConnectionResponse
		orgErr = pageQuery(ctx, operator, filter, &pages)
		if orgErr != nil {
			log.Error(ctx, "GetNameMapByOrganizationOrSchool: pageQuery organization failed",
				log.Err(orgErr),
				log.Any("operator", operator),
				log.Any("filter", filter))
			return
		}
		if len(pages) == 0 {
			log.Warn(ctx, "organization is empty",
				log.Any("operator", operator),
				log.Any("filter", filter))
			return
		}
		//var organizations []*Organization
		for _, page := range pages {
			for _, v := range page.Edges {
				mapLock.RLock()
				nameMap[v.Node.ID] = v.Node.Name
				mapLock.RUnlock()
			}
		}
	}()
	go func() {
		defer wg.Done()
		filter := SchoolFilter{}
		for _, oid := range IDs {
			filter.OR = append(filter.OR, SchoolFilter{
				SchoolID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(oid)},
			})
		}
		var pages []SchoolsConnectionResponse
		schErr = pageQuery(ctx, operator, filter, &pages)
		if schErr != nil {
			log.Error(ctx, "GetNameMapByOrganizationOrSchool: pageQuery school failed",
				log.Err(schErr),
				log.Any("operator", operator),
				log.Any("filter", filter))
			return
		}
		if len(pages) == 0 {
			log.Warn(ctx, "school is empty",
				log.Any("operator", operator),
				log.Any("filter", filter))
			return
		}
		for _, page := range pages {
			for _, v := range page.Edges {
				mapLock.RLock()
				nameMap[v.Node.ID] = v.Node.Name
				mapLock.RUnlock()
			}
		}
	}()
	wg.Wait()
	if orgErr != nil {
		log.Error(ctx, "GetNameMapByOrganizationOrSchool: pageQuery organization failed",
			log.Err(orgErr),
			log.Any("operator", operator),
			log.Strings("ids", id))
		return nil, orgErr
	}
	if schErr != nil {
		log.Error(ctx, "GetNameMapByOrganizationOrSchool: pageQuery school failed",
			log.Err(orgErr),
			log.Any("operator", operator),
			log.Strings("ids", id))
		return nil, schErr
	}
	return nameMap, nil
}
func (aocs AmsOrganizationConnectionService) GetByUserID(ctx context.Context, operator *entity.Operator, id string, options ...APOption) ([]*Organization, error) {
	condition := NewCondition(options...)
	filter := OrganizationFilter{
		UserID: &UUIDFilter{
			Operator: UUIDOperator(OperatorTypeEq),
			Value:    UUID(id),
		},
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		},
	}

	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	var pages []OrganizationsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get org by user failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	if len(pages) == 0 {
		log.Warn(ctx, "org is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*Organization{}, nil
	}
	organizations := make([]*Organization, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, v := range page.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "organization exists",
					log.Any("org", v.Node),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
			org := Organization{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
			}
			organizations = append(organizations, &org)
		}
	}
	return organizations, nil
}

func (aocs AmsOrganizationConnectionService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)
	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$org_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: organizationNode(id: $org_id_%d) {id name status}\n", index, index)
	}
	sb.WriteString("}")

	request := NewRequest(sb.String(), RequestToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("org_id_%d", index), id)
	}

	data := map[string]*Organization{}
	response := &GraphQLSubResponse{
		Data: &data,
	}

	_, err = GetAmsConnection().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get organizations by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	nullableOrganizations := make([]cache.Object, len(ids))
	for index := range ids {
		organization := data[fmt.Sprintf("q%d", indexMapping[index])]
		if organization == nil {
			nullableOrganizations[index] = &NullableOrganization{
				StrID: ids[index],
				Valid: false,
			}
			continue
		}

		nullableOrganizations[index] = &NullableOrganization{
			Organization: *organization,
			StrID:        organization.ID,
			Valid:        true,
		}
	}

	log.Info(ctx, "get orgs by ids success",
		log.Strings("ids", ids),
		log.Any("orgs", nullableOrganizations))

	return nullableOrganizations, nil
}
