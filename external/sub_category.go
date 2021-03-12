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
	GetByProgramAndCategory(ctx context.Context, operator *entity.Operator, programID, categoryID string) ([]*SubCategory, error)
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
		fmt.Fprintf(sb, "q%d: subCategory(id: \"%s\") {id name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*SubCategory{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get subCategorys by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	subCategorys := make([]*SubCategory, 0, len(data))
	for index := range ids {
		subCategory := data[fmt.Sprintf("q%d", indexMapping[index])]
		subCategorys = append(subCategorys, subCategory)
	}

	log.Info(ctx, "get subCategorys by ids success",
		log.Strings("ids", ids),
		log.Any("subCategorys", subCategorys))

	return subCategorys, nil
}

func (s AmsSubCategoryService) GetByProgramAndCategory(ctx context.Context, operator *entity.Operator, programID, categoryID string) ([]*SubCategory, error) {
	request := chlorine.NewRequest(`
	query($program_id: ID!) {
		program(id: $program_id) {
			subCategorys {
				id
				name
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("program_id", programID)

	data := &struct {
		Program struct {
			SubCategorys []*SubCategory `json:"subCategorys"`
		} `json:"program"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query subCategorys by operator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("programID", programID))
		return nil, err
	}

	subCategorys := data.Program.SubCategorys

	log.Info(ctx, "get subCategorys by program success",
		log.Any("operator", operator),
		log.String("programID", programID),
		log.Any("subCategorys", subCategorys))

	return subCategorys, nil
}
