package external

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type SubCategoryServiceProvider interface {
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

func GetSubCategoryServiceProvider() SubCategoryServiceProvider {
	return &AmsSubCategoryService{}
}

type AmsSubCategoryService struct{}

func (s AmsSubCategoryService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*SubCategory, error) {
	if len(ids) == 0 {
		return []*SubCategory{}, nil
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

	_, err := GetAmsClient().Run(ctx, request, response)
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

	subCategories := make([]*SubCategory, 0, len(data))
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

func (s AmsSubCategoryService) GetByCategory(ctx context.Context, operator *entity.Operator, categoryID string, options ...APOption) ([]*SubCategory, error) {
	condition := NewCondition(options...)

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
	request.Var("category_id", categoryID)

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
			log.String("categoryID", categoryID),
			log.Any("condition", condition))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get sub categories by operator failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("categoryID", categoryID),
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
		log.String("categoryID", categoryID),
		log.Any("condition", condition),
		log.Any("subCategories", subCategories))

	return subCategories, nil
}

func (s AmsSubCategoryService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*SubCategory, error) {
	condition := NewCondition(options...)

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
	request.Var("organization_id", operator.OrgID)

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
