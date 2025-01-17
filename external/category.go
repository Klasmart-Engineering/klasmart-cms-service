package external

import (
	"context"
	"fmt"
	"strings"

	"github.com/KL-Engineering/kidsloop-cms-service/config"

	"github.com/KL-Engineering/kidsloop-cache/cache"

	"github.com/KL-Engineering/chlorine"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type CategoryServiceProvider interface {
	cache.IDataSource
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Category, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*Category, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
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

func (n *Category) StringID() string {
	return n.ID
}
func (n *Category) RelatedIDs() []*cache.RelatedEntity {
	return nil
}
func GetCategoryServiceProvider() CategoryServiceProvider {
	if config.Get().AMS.UseDeprecatedQuery {
		return &AmsCategoryService{}
	}
	return &AmsCategoryConnectionService{}
}

type AmsCategoryService struct{}

func (s AmsCategoryService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Category, error) {
	if len(ids) == 0 {
		return []*Category{}, nil
	}

	uuids := make([]string, 0, len(ids))
	for _, id := range ids {
		if utils.IsValidUUID(id) {
			uuids = append(uuids, id)
		} else {
			log.Warn(ctx, "invalid uuid type", log.String("id", id))
		}
	}

	res := make([]*Category, 0, len(uuids))
	err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), uuids, &res, operator)
	if err != nil {
		return nil, err
	}

	return res, nil
}
func (s AmsCategoryService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$category_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: categoryNode(id: $category_id_%d) {id name status system}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("category_id_%d", index), id)
	}

	data := map[string]*Category{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetAmsClient().Run(ctx, request, response)
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

	categories := make([]cache.Object, 0, len(data))
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

func (s AmsCategoryService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*Category, error) {
	categories, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]*Category{}, err
	}

	dict := make(map[string]*Category, len(categories))
	for _, category := range categories {
		dict[category.ID] = category
	}

	return dict, nil
}

func (s AmsCategoryService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	categories, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]string{}, err
	}

	dict := make(map[string]string, len(categories))
	for _, category := range categories {
		dict[category.ID] = category.Name
	}

	return dict, nil
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

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$subject_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: subjectNode(id: $subject_id_%d) {categories {id name status system}}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("subject_id_%d", index), id)
	}

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
func (s AmsCategoryService) Name() string {
	return "ams_category_service"
}
