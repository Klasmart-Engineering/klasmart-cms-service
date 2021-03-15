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

type ProgramServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Program, error)
	GetByOrganization(ctx context.Context, operator *entity.Operator) ([]*Program, error)
}

type Program struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	GroupName string `json:"group_name"`
}

func GetProgramServiceProvider() ProgramServiceProvider {
	return &AmsProgramService{}
}

type AmsProgramService struct{}

func (s AmsProgramService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Program, error) {
	if len(ids) == 0 {
		return []*Program{}, nil
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _ids {
		fmt.Fprintf(sb, "q%d: program(id: \"%s\") {id name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*Program{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get programs by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	programs := make([]*Program, 0, len(data))
	for index := range ids {
		program := data[fmt.Sprintf("q%d", indexMapping[index])]
		if program == nil {
			log.Error(ctx, "program not found", log.String("id", ids[index]))
			return nil, constant.ErrRecordNotFound
		}
		programs = append(programs, program)
	}

	log.Info(ctx, "get programs by ids success",
		log.Strings("ids", ids),
		log.Any("programs", programs))

	return programs, nil
}

func (s AmsProgramService) GetByOrganization(ctx context.Context, operator *entity.Operator) ([]*Program, error) {
	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			programs {
				id
				name
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("organization_id", operator.OrgID)

	data := &struct {
		Organization struct {
			Programs []*Program `json:"programs"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query programs by operator failed",
			log.Err(err),
			log.Any("operator", operator))
		return nil, err
	}

	programs := data.Organization.Programs

	log.Info(ctx, "get programs by operator success",
		log.Any("operator", operator),
		log.Any("programs", programs))

	return programs, nil
}
