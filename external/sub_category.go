package external

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external/gql"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type SubCategoryServiceProvider interface {
	cache.IDataSource
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*SubCategory, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*SubCategory, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
	GetByCategory(ctx context.Context, operator *entity.Operator, categoryID string, options ...APOption) ([]*SubCategory, error)
	GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*SubCategory, error)
}

type SubCategory struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Status APStatus `json:"status"`
	System bool     `json:"system"`
}

func (n *SubCategory) StringID() string {
	return n.ID
}
func (n *SubCategory) RelatedIDs() []*cache.RelatedEntity {
	return nil
}

func GetSubCategoryServiceProvider() SubCategoryServiceProvider {
	return &AmsSubCategoryService{}
}

type AmsSubCategoryService struct{}

func (s AmsSubCategoryService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*SubCategory, error) {
	if len(ids) == 0 {
		return []*SubCategory{}, nil
	}

	uuids := make([]string, 0, len(ids))
	for _, id := range ids {
		if utils.IsValidUUID(id) {
			uuids = append(uuids, id)
		} else {
			log.Warn(ctx, "invalid uuid type", log.String("id", id))
		}
	}

	res := make([]*SubCategory, 0, len(uuids))
	err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), uuids, &res, operator)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s AmsSubCategoryService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$subcategory_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: subcategory(id: $subcategory_id_%d) {id name status system}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("subcategory_id_%d", index), id)
	}

	data := map[string]*SubCategory{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get subCategories by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get subCategories by ids failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Strings("ids", ids))
		return nil, response.Errors
	}

	subCategories := make([]cache.Object, 0, len(data))
	for index := range ids {
		subCategory := data[fmt.Sprintf("q%d", indexMapping[index])]
		if subCategory == nil {
			log.Error(ctx, "subCategory not found", log.String("id", ids[index]))
			return nil, constant.ErrRecordNotFound
		}
		subCategories = append(subCategories, subCategory)
	}

	log.Info(ctx, "get subCategories by ids success",
		log.Strings("ids", ids),
		log.Any("subCategories", subCategories))

	return subCategories, nil
}

func (s AmsSubCategoryService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*SubCategory, error) {
	subCategories, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]*SubCategory{}, err
	}

	dict := make(map[string]*SubCategory, len(subCategories))
	for _, subCategory := range subCategories {
		dict[subCategory.ID] = subCategory
	}

	return dict, nil
}

func (s AmsSubCategoryService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	subCategories, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]string{}, err
	}

	dict := make(map[string]string, len(subCategories))
	for _, subCategory := range subCategories {
		dict[subCategory.ID] = subCategory.Name
	}

	return dict, nil
}

func (s AmsSubCategoryService) getWithCategory(ctx context.Context, operator *entity.Operator, id string, condition *APCondition) ([]*SubCategory, error) {
	filter := gql.SubcategoryFilter{
		CategoryID: &gql.UUIDFilter{
			Operator: gql.OperatorTypeEq,
			Value:    gql.UUID(id),
		},
		Status: &gql.StringFilter{
			Operator: gql.OperatorTypeEq,
			Value:    Active.String(),
		},
	}
	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	if condition.System.Valid {
		filter.System = &gql.BooleanFilter{
			Operator: gql.OperatorTypeEq,
			Value:    condition.System.Valid,
		}
	}

	var subCategories []*SubCategory
	var pages []gql.SubcategoriesConnectionResponse
	err := gql.Query(ctx, operator, filter.FilterType(), filter, &pages)
	if err != nil {
		log.Error(ctx, "get subcategory by category failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	for _, p := range pages {
		for _, v := range p.Edges {
			obj := &SubCategory{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			subCategories = append(subCategories, obj)
		}
	}
	return subCategories, nil
}
func (s AmsSubCategoryService) getByCategory(ctx context.Context, operator *entity.Operator, id string, condition *APCondition) ([]*SubCategory, error) {
	request := chlorine.NewRequest(`
	query($category_id: ID!) {
		category(id: $category_id) {
			subcategories {
				id
				name
				status
				system
			}
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("category_id", id)

	data := &struct {
		Category struct {
			SubCategories []*SubCategory `json:"subcategories"`
		} `json:"category"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query sub categories by operator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("categoryID", id),
			log.Any("condition", condition))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get sub categories by operator failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("categoryID", id),
			log.Any("condition", condition))
		return nil, response.Errors
	}

	subCategories := make([]*SubCategory, 0, len(data.Category.SubCategories))
	for _, subCategory := range data.Category.SubCategories {
		if condition.Status.Valid {
			if condition.Status.Status != subCategory.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if subCategory.Status != Active {
				continue
			}
		}

		if condition.System.Valid && subCategory.System != condition.System.Bool {
			continue
		}

		subCategories = append(subCategories, subCategory)
	}

	log.Info(ctx, "get sub categories by program success",
		log.Any("operator", operator),
		log.String("categoryID", id),
		log.Any("condition", condition),
		log.Any("subCategories", subCategories))

	return subCategories, nil
}
func (s AmsSubCategoryService) GetByCategory(ctx context.Context, operator *entity.Operator, categoryID string, options ...APOption) ([]*SubCategory, error) {
	condition := NewCondition(options...)

	if config.Get().AMS.ReplaceWithConnection {
		return s.getWithCategory(ctx, operator, categoryID, condition)
	}
	return s.getByCategory(ctx, operator, categoryID, condition)
}

func (s AmsSubCategoryService) getWithOrganization(ctx context.Context, operator *entity.Operator, id string, condition *APCondition) ([]*SubCategory, error) {
	filter := gql.SubcategoryFilter{
		OrganizationID: &gql.UUIDFilter{
			Operator: gql.OperatorTypeEq,
			Value:    gql.UUID(id),
		},
		Status: &gql.StringFilter{
			Operator: gql.OperatorTypeEq,
			Value:    Active.String(),
		},
	}
	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	if condition.System.Valid {
		filter.System = &gql.BooleanFilter{
			Operator: gql.OperatorTypeEq,
			Value:    condition.System.Valid,
		}
	}

	var subCategories []*SubCategory
	var pages []gql.SubcategoriesConnectionResponse
	err := gql.Query(ctx, operator, filter.FilterType(), filter, &pages)
	if err != nil {
		log.Error(ctx, "get subcategory by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	for _, p := range pages {
		for _, v := range p.Edges {
			obj := &SubCategory{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			subCategories = append(subCategories, obj)
		}
	}
	return subCategories, nil

}
func (s AmsSubCategoryService) getByOrganization(ctx context.Context, operator *entity.Operator, id string, condition *APCondition) ([]*SubCategory, error) {
	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			subcategories {
				id
				name
				status
				system
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("organization_id", id)

	data := &struct {
		Organization struct {
			SubCategories []*SubCategory `json:"subcategories"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query sub categories by operator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("condition", condition))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "query sub categories by operator failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Any("condition", condition))
		return nil, response.Errors
	}

	subCategories := make([]*SubCategory, 0, len(data.Organization.SubCategories))
	for _, subCategory := range data.Organization.SubCategories {
		if condition.Status.Valid {
			if condition.Status.Status != subCategory.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if subCategory.Status != Active {
				continue
			}
		}

		if condition.System.Valid && subCategory.System != condition.System.Bool {
			continue
		}

		subCategories = append(subCategories, subCategory)
	}

	log.Info(ctx, "get sub categories by operator success",
		log.Any("operator", operator),
		log.Any("condition", condition),
		log.Any("subcategories", subCategories))

	return subCategories, nil
}
func (s AmsSubCategoryService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*SubCategory, error) {
	condition := NewCondition(options...)

	if config.Get().AMS.ReplaceWithConnection {
		return s.getWithOrganization(ctx, operator, operator.OrgID, condition)
	}
	return s.getByOrganization(ctx, operator, operator.OrgID, condition)
}

func (s AmsSubCategoryService) Name() string {
	return "ams_subcategory_service"
}
