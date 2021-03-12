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

type GradeServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Grade, error)
	GetByProgram(ctx context.Context, operator *entity.Operator, programID string) ([]*Grade, error)
}

type Grade struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetGradeServiceProvider() GradeServiceProvider {
	return &AmsGradeService{}
}

type AmsGradeService struct{}

func (s AmsGradeService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Grade, error) {
	if len(ids) == 0 {
		return []*Grade{}, nil
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _ids {
		fmt.Fprintf(sb, "q%d: grade(id: \"%s\") {id name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*Grade{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get grades by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	grades := make([]*Grade, 0, len(data))
	for index := range ids {
		grade := data[fmt.Sprintf("q%d", indexMapping[index])]
		grades = append(grades, grade)
	}

	log.Info(ctx, "get grades by ids success",
		log.Strings("ids", ids),
		log.Any("grades", grades))

	return grades, nil
}

func (s AmsGradeService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string) ([]*Grade, error) {
	request := chlorine.NewRequest(`
	query($program_id: ID!) {
		program(id: $program_id) {
			grades {
				id
				name
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
			log.String("programID", programID))
		return nil, err
	}

	grades := data.Program.Grades

	log.Info(ctx, "get grades by program success",
		log.Any("operator", operator),
		log.String("programID", programID),
		log.Any("grades", grades))

	return grades, nil
}
