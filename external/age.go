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

type AgeServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Age, error)
	GetByProgram(ctx context.Context, operator *entity.Operator, programID string) ([]*Age, error)
}

type Age struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetAgeServiceProvider() AgeServiceProvider {
	return &AmsAgeService{}
}

type AmsAgeService struct{}

func (s AmsAgeService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Age, error) {
	if len(ids) == 0 {
		return []*Age{}, nil
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _ids {
		fmt.Fprintf(sb, "q%d: age(id: \"%s\") {id name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*Age{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get ages by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	ages := make([]*Age, 0, len(data))
	for index := range ids {
		age := data[fmt.Sprintf("q%d", indexMapping[index])]
		ages = append(ages, age)
	}

	log.Info(ctx, "get ages by ids success",
		log.Strings("ids", ids),
		log.Any("ages", ages))

	return ages, nil
}

func (s AmsAgeService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string) ([]*Age, error) {
	request := chlorine.NewRequest(`
	query($program_id: ID!) {
		program(id: $program_id) {
			ages {
				id
				name
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("program_id", programID)

	data := &struct {
		Program struct {
			Ages []*Age `json:"ages"`
		} `json:"program"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query ages by operator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("programID", programID))
		return nil, err
	}

	ages := data.Program.Ages

	log.Info(ctx, "get ages by program success",
		log.Any("operator", operator),
		log.String("programID", programID),
		log.Any("ages", ages))

	return ages, nil
}
