package external

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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

func (pcs AmsProgramConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Program, error) {
	condition := NewCondition(options...)

	filter := ProgramFilter{
		Status: &StringFilter{Operator: StringOperator(OperatorTypeEq), Value: Active.String()},
		OR: []ProgramFilter{
			{OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(operator.OrgID)}},
			{System: &BooleanFilter{Operator: OperatorTypeEq, Value: true}},
		},
	}
	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	if condition.System.Valid {
		filter.System = &BooleanFilter{
			Operator: OperatorTypeEq,
			Value:    condition.System.Valid,
		}
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

	if len(pages) == 0 {
		log.Warn(ctx, "program is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*Program{}, nil
	}

	programs := make([]*Program, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, p := range pages {
		for _, v := range p.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "program exist",
					log.Any("program", v),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
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
