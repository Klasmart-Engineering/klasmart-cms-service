package external

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external/gql"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type ProgramServiceProvider interface {
	cache.IDataSource
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Program, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*Program, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
	GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Program, error)
}

type Program struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	GroupName string   `json:"group_name"`
	Status    APStatus `json:"status"`
	System    bool     `json:"system"`
}

func (n *Program) StringID() string {
	return n.ID
}
func (n *Program) RelatedIDs() []*cache.RelatedEntity {
	return nil
}
func GetProgramServiceProvider() ProgramServiceProvider {
	return &AmsProgramService{}
}

type AmsProgramService struct{}

func (s AmsProgramService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Program, error) {
	if len(ids) == 0 {
		return []*Program{}, nil
	}

	uuids := make([]string, 0, len(ids))
	for _, id := range ids {
		if utils.IsValidUUID(id) {
			uuids = append(uuids, id)
		} else {
			log.Warn(ctx, "invalid uuid type", log.String("id", id))
		}
	}

	res := make([]*Program, 0, len(uuids))
	err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), uuids, &res, operator)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s AmsProgramService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		fmt.Println("options:", options)
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$program_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: program(id: $program_id_%d) {id name status system}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("program_id_%d", index), id)

	}

	data := map[string]*Program{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetAmsClient().Run(ctx, request, response)
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

	programs := make([]cache.Object, 0, len(data))
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
func (s AmsProgramService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*Program, error) {
	programs, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]*Program{}, err
	}

	dict := make(map[string]*Program, len(programs))
	for _, program := range programs {
		dict[program.ID] = program
	}

	return dict, nil
}

func (s AmsProgramService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	programs, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]string{}, err
	}

	dict := make(map[string]string, len(programs))
	for _, program := range programs {
		dict[program.ID] = program.Name
	}

	return dict, nil
}

func (s AmsProgramService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Program, error) {
	condition := NewCondition(options...)
	if constant.ReplaceWithConnection {
		filter := gql.ProgramFilter{
			OrganizationID: &gql.UUIDFilter{
				Operator: gql.OperatorTypeEq,
				Value:    gql.UUID(operator.OrgID),
			},
			Status: &gql.StringFilter{
				Operator: gql.OperatorTypeEq,
				Value:    Active.String(),
			},
		}
		if condition.Status.Valid {
			filter.Status.Value = condition.Status.Status.String()
		}
		if condition.System.Valid {
			filter.System = &gql.BooleanFilter{
				Operator: gql.OperatorTypeEq,
				Value:    condition.System.Valid,
			}
		}

		var programs []*Program
		var pages []gql.ProgramsConnectionResponse
		err := gql.Query(ctx, operator, filter.FilterType(), filter, &pages)
		if err != nil {
			log.Error(ctx, "get programs by ids failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("filter", filter))
			return nil, err
		}
		for _, p := range pages {
			for _, v := range p.Edges {
				obj := &Program{
					ID:   v.Node.ID,
					Name: v.Node.Name,
					//GroupName:
					Status: APStatus(v.Node.Status),
					System: v.Node.System,
				}
				programs = append(programs, obj)
			}
		}
		return programs, nil
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
		log.Error(ctx, "query programs failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("condition", condition))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "query programs failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Any("condition", condition))
		return nil, response.Errors
	}

	programs := make([]*Program, 0, len(data.Organization.Programs))
	for _, program := range data.Organization.Programs {
		if condition.Status.Valid {
			if condition.Status.Status != program.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if program.Status != Active {
				continue
			}
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

func (s AmsProgramService) Name() string {
	return "ams_program_service"
}
