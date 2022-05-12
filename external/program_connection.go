package external

import (
	"context"
	"fmt"
	"strings"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cache/cache"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type ProgramFilter struct {
	ID             *UUIDFilter         `json:"id,omitempty" gqls:"id,omitempty"`
	Name           *StringFilter       `json:"name,omitempty" gqls:"name,omitempty"`
	Status         *StringFilter       `json:"status,omitempty" gqls:"status,omitempty"`
	System         *BooleanFilter      `json:"system,omitempty" gqls:"system,omitempty"`
	OrganizationID *UUIDFilter         `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	GradeID        *UUIDFilter         `json:"gradeId,omitempty" gqls:"gradeId,omitempty"`
	AgeRangeFrom   *AgeRangeTypeFilter `json:"ageRangeFrom,omitempty" gqls:"ageRangeFrom,omitempty"`
	AgeRangeTo     *AgeRangeTypeFilter `json:"ageRangeTo,omitempty" gqls:"ageRangeTo,omitempty"`
	SubjectID      *UUIDFilter         `json:"subjectId,omitempty" gqls:"subjectId,omitempty"`
	SchoolID       *UUIDFilter         `json:"schoolId,omitempty" gqls:"schoolId,omitempty"`
	ClassID        *UUIDFilter         `json:"classId,omitempty" gqls:"classId,omitempty"`
	AND            []ProgramFilter     `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []ProgramFilter     `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (ProgramFilter) FilterName() FilterType {
	return ProgramFilterType
}

func (ProgramFilter) ConnectionName() ConnectionType {
	return ProgramsConnectionType
}

type ProgramConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}
type ProgramsConnectionEdge struct {
	Cursor string                `json:"cursor" gqls:"cursor"`
	Node   ProgramConnectionNode `json:"node" gqls:"node"`
}

type ProgramsConnectionResponse struct {
	TotalCount int                      `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo       `json:"pageInfo" gqls:"pageInfo"`
	Edges      []ProgramsConnectionEdge `json:"edges" gqls:"edges"`
}

func (pcs ProgramsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}

type AmsProgramConnectionService struct {
	AmsProgramService
}

func (pcs AmsProgramConnectionService) pageNodes(ctx context.Context, operator *entity.Operator, pages []ProgramsConnectionResponse) []*Program {
	if len(pages) == 0 {
		log.Warn(ctx, "pageNodes is empty",
			log.Any("operator", operator))
		return []*Program{}
	}
	programs := make([]*Program, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, edge := range page.Edges {
			if _, ok := exists[edge.Node.ID]; ok {
				log.Warn(ctx, "pageNodes: category exist",
					log.Any("category", edge.Node),
					log.Any("operator", operator))
				continue
			}
			exists[edge.Node.ID] = true
			obj := &Program{
				ID:   edge.Node.ID,
				Name: edge.Node.Name,
				//GroupName:
				Status: APStatus(edge.Node.Status),
				System: edge.Node.System,
			}
			programs = append(programs, obj)
		}
	}
	return programs
}

func (pcs AmsProgramConnectionService) NewProgramFilter(ctx context.Context, operator *entity.Operator, options ...APOption) *ProgramFilter {
	condition := NewCondition(options...)
	var filter ProgramFilter
	if condition.Status.Valid && condition.Status.Status != Ignore {
		filter.Status = &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    condition.Status.Status.String(),
		}
	} else if !condition.Status.Valid {
		filter.Status = &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
		}
	}
	if condition.System.Valid {
		filter.System = &BooleanFilter{
			Operator: OperatorTypeEq,
			Value:    condition.System.Bool,
		}
	}
	return &filter
}

func (pcs AmsProgramConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Program, error) {
	filter := pcs.NewProgramFilter(ctx, operator, options...)

	filter.OR = []ProgramFilter{
		{OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(operator.OrgID)}},
		{System: &BooleanFilter{Operator: OperatorTypeEq, Value: true}},
	}

	var pages []ProgramsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get programs by ids failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}

	programs := pcs.pageNodes(ctx, operator, pages)
	return programs, nil
}

func (pcs AmsProgramConnectionService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
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
		fmt.Fprintf(sb, "q%d: programNode(id: $program_id_%d) {id name status system}\n", index, index)
	}
	sb.WriteString("}")

	request := NewRequest(sb.String(), RequestToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("program_id_%d", index), id)
	}

	data := map[string]*Program{}
	response := &GraphQLSubResponse{
		Data: &data,
	}

	_, err = GetAmsConnection().Run(ctx, request, response)
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
