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

type CategoryServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Category, error)
	GetByProgram(ctx context.Context, operator *entity.Operator, programID string) ([]*Category, error)
	GetByOrganization(ctx context.Context, operator *entity.Operator) ([]*Category, error)
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetCategoryServiceProvider() CategoryServiceProvider {
	return &AmsCategoryService{}
}

type AmsCategoryService struct{}

func (s AmsCategoryService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Category, error) {
	if len(ids) == 0 {
		return []*Category{}, nil
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _ids {
		fmt.Fprintf(sb, "q%d: category(id: \"%s\") {id name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*Category{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get categories by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	categories := make([]*Category, 0, len(data))
	for index := range ids {
		category := data[fmt.Sprintf("q%d", indexMapping[index])]
		if category == nil {
			log.Error(ctx, "category not found", log.String("id", ids[index]))
			return nil, constant.ErrRecordNotFound
		}
		categories = append(categories, category)
	}

	log.Info(ctx, "get categories by ids success",
		log.Strings("ids", ids),
		log.Any("categories", categories))

	return categories, nil
}

func (s AmsCategoryService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string) ([]*Category, error) {
	request := chlorine.NewRequest(`
	query($program_id: ID!) {
		program(id: $program_id) {
			subjects {
				categories {
					id
					name
				}
			}				
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("program_id", programID)

	data := &struct {
		Program struct {
			Subjects []struct {
				Categories []*Category `json:"categories"`
			} `json:"subjects"`
		} `json:"program"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query categories by operator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("programID", programID))
		return nil, err
	}

	categories := make([]*Category, 0)
	for _, subject := range data.Program.Subjects {
		categories = append(categories, subject.Categories...)
	}

	log.Info(ctx, "get categories by program success",
		log.Any("operator", operator),
		log.String("programID", programID),
		log.Any("categories", categories))

	return categories, nil
}

func (s AmsCategoryService) GetByOrganization(ctx context.Context, operator *entity.Operator) ([]*Category, error) {
	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			categories {
				id
				name
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("organization_id", operator.OrgID)

	data := &struct {
		Organization struct {
			Categories []*Category `json:"categories"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query categories by operator failed",
			log.Err(err),
			log.Any("operator", operator))
		return nil, err
	}

	categories := data.Organization.Categories

	log.Info(ctx, "get categories by operator success",
		log.Any("operator", operator),
		log.Any("categories", categories))

	return categories, nil
}
