package external

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type GradeServiceProvider interface {
	cache.IDataSource
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Grade, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*Grade, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
	GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Grade, error)
	GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Grade, error)
}

type Grade struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Status APStatus `json:"status"`
	System bool     `json:"system"`
}

func (n *Grade) StringID() string {
	return n.ID
}
func (n *Grade) RelatedIDs() []*cache.RelatedEntity {
	return nil
}
func GetGradeServiceProvider() GradeServiceProvider {
	return &AmsGradeService{}
}

type AmsGradeService struct{}

func (s AmsGradeService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Grade, error) {
	if len(ids) == 0 {
		return []*Grade{}, nil
	}

	uuids := make([]string, 0, len(ids))
	for _, id := range ids {
		if utils.IsValidUUID(id) {
			uuids = append(uuids, id)
		} else {
			log.Warn(ctx, "invalid uuid type", log.String("id", id))
		}
	}

	res := make([]*Grade, 0, len(uuids))
	err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), uuids, &res, operator)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s AmsGradeService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$grade_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: grade(id: $grade_id_%d) {id name status system}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("grade_id_%d", index), id)
	}

	data := map[string]*Grade{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get grades by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get grades by ids failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Strings("ids", ids))
		return nil, response.Errors
	}

	grades := make([]cache.Object, 0, len(data))
	for index := range ids {
		grade := data[fmt.Sprintf("q%d", indexMapping[index])]
		if grade == nil {
			log.Error(ctx, "grade not found", log.String("id", ids[index]))
			return nil, constant.ErrRecordNotFound
		}
		grades = append(grades, grade)
	}

	log.Info(ctx, "get grades by ids success",
		log.Strings("ids", ids),
		log.Any("grades", grades))

	return grades, nil
}
func (s AmsGradeService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*Grade, error) {
	grades, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]*Grade{}, err
	}

	dict := make(map[string]*Grade, len(grades))
	for _, grade := range grades {
		dict[grade.ID] = grade
	}

	return dict, nil
}

func (s AmsGradeService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	grades, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]string{}, err
	}

	dict := make(map[string]string, len(grades))
	for _, grade := range grades {
		dict[grade.ID] = grade.Name
	}

	return dict, nil
}

func (s AmsGradeService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Grade, error) {
	condition := NewCondition(options...)

	request := chlorine.NewRequest(`
	query($program_id: ID!) {
		program(id: $program_id) {
			grades {
				id
				name
				status
				system
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("program_id", programID)

	data := &struct {
		Program struct {
			Grades []*Grade `json:"grades"`
		} `json:"program"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query grades by operator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("programID", programID),
			log.Any("condition", condition))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get grades by operator failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("programID", programID),
			log.Any("condition", condition))
		return nil, response.Errors
	}

	grades := make([]*Grade, 0, len(data.Program.Grades))
	for _, grade := range data.Program.Grades {
		if condition.Status.Valid {
			if condition.Status.Status != grade.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if grade.Status != Active {
				continue
			}
		}

		if condition.System.Valid && grade.System != condition.System.Bool {
			continue
		}

		grades = append(grades, grade)
	}

	log.Info(ctx, "get grades by program success",
		log.Any("operator", operator),
		log.String("programID", programID),
		log.Any("condition", condition),
		log.Any("grades", grades))

	return grades, nil
}

func (s AmsGradeService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Grade, error) {
	condition := NewCondition(options...)

	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			grades {
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
			Grades []*Grade `json:"grades"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query grades by operator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("condition", condition))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "query grades by operator failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Any("condition", condition))
		return nil, response.Errors
	}

	grades := make([]*Grade, 0, len(data.Organization.Grades))
	for _, grade := range data.Organization.Grades {
		if condition.Status.Valid {
			if condition.Status.Status != grade.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if grade.Status != Active {
				continue
			}
		}

		if condition.System.Valid && grade.System != condition.System.Bool {
			continue
		}

		grades = append(grades, grade)
	}

	log.Info(ctx, "get grades by operator success",
		log.Any("operator", operator),
		log.Any("condition", condition),
		log.Any("grades", grades))

	return grades, nil
}
func (s AmsGradeService) Name() string {
	return "ams_grade_service"
}
