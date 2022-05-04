package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
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
	schools, err := GetSchoolServiceProvider().GetByClasses(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "GetByClasses: get school failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("class_ids", classIDs))
		return nil, err
	}
	if len(schools) == 0 {
		log.Debug(ctx, "GetByClasses: school is empty",
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
		err := pageQuery(ctx, operator, filter, &pages)
		if err != nil {

		}
		orgErr = pageQuery(ctx, operator, filter, &pages)
		if err != nil {
			log.Error(ctx, "GetNameMapByOrganizationOrSchool: pageQuery organization failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("filter", filter))
			return
		}
		//var organizations []*Organization
		for _, page := range pages {
			for _, v := range page.Edges {
				nameMap[v.Node.ID] = v.Node.Name
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
		err := pageQuery(ctx, operator, filter, &pages)
		if err != nil {

		}
		schErr = pageQuery(ctx, operator, filter, &pages)
		if err != nil {
			log.Error(ctx, "GetNameMapByOrganizationOrSchool: pageQuery school failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("filter", filter))
			return
		}

		for _, page := range pages {
			for _, v := range page.Edges {
				nameMap[v.Node.ID] = v.Node.Name
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
	var organizations []*Organization
	for _, page := range pages {
		for _, v := range page.Edges {
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
