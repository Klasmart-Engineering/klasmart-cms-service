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
	GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Category, error)
	GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Category, error)
	GetBySubjects(ctx context.Context, operator *entity.Operator, subjectIDs []string, options ...APOption) ([]*Category, error)
}

type Category struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Status APStatus `json:"status"`
	System bool     `json:"system"`
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
		fmt.Fprintf(sb, "q%d: category(id: \"%s\") {id name status system}\n", index, id)
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

	if len(response.Errors) > 0 {
		log.Error(ctx, "get categories by ids failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Strings("ids", ids))
		return nil, response.Errors
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

func (s AmsCategoryService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Category, error) {
	condition := NewCondition(options...)

	request := chlorine.NewRequest(`
	query($program_id: ID!) {
		program(id: $program_id) {
			subjects {
				categories {
					id
					name
					status
					system
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
			log.String("programID", programID),
			log.Any("condition", condition))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "query categories by operator failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("programID", programID),
			log.Any("condition", condition))
		return nil, response.Errors
	}

	categories := make([]*Category, 0, len(data.Program.Subjects))
	for _, subject := range data.Program.Subjects {
		for _, category := range subject.Categories {
			if condition.Status.Valid {
				if condition.Status.Status != category.Status {
					continue
				}
			} else {
				// only status = "Active" data is returned by default
				if category.Status != Active {
					continue
				}
			}

			if condition.System.Valid && category.System != condition.System.Bool {
				continue
			}

			categories = append(categories, category)
		}
	}

	log.Info(ctx, "get categories by program success",
		log.Any("operator", operator),
		log.String("programID", programID),
		log.Any("condition", condition),
		log.Any("categories", categories))

	return categories, nil
}

func (s AmsCategoryService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Category, error) {
	condition := NewCondition(options...)

	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			categories {
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
			log.Any("operator", operator),
			log.Any("condition", condition))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "query categories by operator failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Any("condition", condition))
		return nil, response.Errors
	}

	categories := make([]*Category, 0, len(data.Organization.Categories))
	for _, category := range data.Organization.Categories {
		if condition.Status.Valid {
			if condition.Status.Status != category.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if category.Status != Active {
				continue
			}
		}

		if condition.System.Valid && category.System != condition.System.Bool {
			continue
		}

		categories = append(categories, category)
	}

	log.Info(ctx, "get categories by operator success",
		log.Any("operator", operator),
		log.Any("condition", condition),
		log.Any("categories", categories))

	return categories, nil
}

func (s AmsCategoryService) GetBySubjects(ctx context.Context, operator *entity.Operator, subjectIDs []string, options ...APOption) ([]*Category, error) {
	if len(subjectIDs) == 0 {
		return []*Category{}, nil
	}

	condition := NewCondition(options...)

	_ids, indexMapping := utils.SliceDeduplicationMap(subjectIDs)

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _ids {
		fmt.Fprintf(sb, "q%d: subject(id: \"%s\") {categories {id name status system}}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	data := map[string]struct {
		Categories []*Category `json:"categories"`
	}{}

	response := &chlorine.Response{
		Data: &data,
	}
	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get categories by ids failed",
			log.Err(err),
			log.Strings("subjectIDs", _ids))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get categories by ids failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Strings("subjectIDs", _ids))
		return nil, response.Errors
	}
	fmt.Println(indexMapping)
	for key, item := range data {
		fmt.Println(key, ":", item.Categories)
	}
	categoryMap := make(map[string]*Category)
	result := make([]*Category, 0)
	for index := range subjectIDs {
		categories := data[fmt.Sprintf("q%d", indexMapping[index])]
		if len(categories.Categories) == 0 {
			continue
		}
		for _, category := range categories.Categories {
			if _, ok := categoryMap[category.ID]; ok {
				continue
			}

			if condition.Status.Valid {
				if condition.Status.Status != category.Status {
					continue
				}
			} else {
				// only status = "Active" data is returned by default
				if category.Status != Active {
					continue
				}
			}

			if condition.System.Valid && category.System != condition.System.Bool {
				continue
			}
			categoryMap[category.ID] = category
			result = append(result, category)
		}
	}
	log.Info(ctx, "get categories by subjectIDs success",
		log.Strings("subjectIDs", _ids),
		log.Any("categories", result))

	return result, nil
}
