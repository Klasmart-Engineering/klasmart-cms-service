package external

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type SubCategoryServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*SubCategory, error)
	GetByCategory(ctx context.Context, operator *entity.Operator, categoryID string) ([]*SubCategory, error)
}

type SubCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
	sb.WriteString("query {")
	for index, id := range _ids {
		fmt.Fprintf(sb, "q%d: subcategory(id: \"%s\") {id name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

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

	subCategories := make([]*SubCategory, 0, len(data))
	for index := range ids {
		subCategory := data[fmt.Sprintf("q%d", indexMapping[index])]
		subCategories = append(subCategories, subCategory)
	}

	log.Info(ctx, "get subCategories by ids success",
		log.Strings("ids", ids),
		log.Any("subCategories", subCategories))

	return subCategories, nil
}

func (s AmsSubCategoryService) GetByCategory(ctx context.Context, operator *entity.Operator, categoryID string) ([]*SubCategory, error) {
	request := chlorine.NewRequest(`
	query($category_id: ID!) {
		category(id: $category_id) {
			subcategories {
				id
				name
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
		log.Error(ctx, "query subCategories by operator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("categoryID", categoryID))
		return nil, err
	}

	subCategories := data.Category.SubCategories

	log.Info(ctx, "get subCategories by program success",
		log.Any("operator", operator),
		log.String("categoryID", categoryID),
		log.Any("subCategories", subCategories))

	return subCategories, nil
}
