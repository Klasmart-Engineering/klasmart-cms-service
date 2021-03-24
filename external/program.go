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
	Query(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Program, error)
}

type Program struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	GroupName string   `json:"group_name"`
	Status    APStatus `json:"status"`
	System    bool     `json:"system"`
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
		fmt.Fprintf(sb, "q%d: program(id: \"%s\") {id name status system}\n", index, id)
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

	if len(response.Errors) > 0 {
		log.Error(ctx, "get programs by ids failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Strings("ids", ids))
		return nil, response.Errors
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
	return s.Query(ctx, operator, WithOrganization(operator.OrgID), WithStatus(Active))
}

func (s AmsProgramService) Query(ctx context.Context, operator *entity.Operator, conditions ...APOption) ([]*Program, error) {
	condition := NewCondition(conditions...)
	if !condition.OrganizationID.Valid || condition.OrganizationID.String == "" {
		log.Debug(ctx, "query program without organization id", log.Any("options", conditions))
		return []*Program{}, nil
	}

	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			programs {
				id
				name
				status
				system
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("organization_id", condition.OrganizationID.String)

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
		log.Error(ctx, "query programs failed",
			log.Err(err),
			log.Any("operator", operator))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "query programs failed",
			log.Err(response.Errors),
			log.Any("operator", operator))
		return nil, response.Errors
	}

	programs := make([]*Program, 0, len(data.Organization.Programs))
	for _, program := range data.Organization.Programs {
		if condition.Status.Valid && condition.Status.Status != program.Status {
			continue
		}

		if condition.System.Valid && program.System != condition.System.Bool {
			continue
		}

		programs = append(programs, program)
	}

	log.Info(ctx, "query programs success",
		log.Any("operator", operator),
		log.Any("condition", condition),
		log.Any("programs", programs))

	return programs, nil
}
