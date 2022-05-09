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

type AmsGradeConnectionService struct {
	AmsGradeService
}

type GradeFilter struct {
	ID             *UUIDFilter    `json:"id,omitempty" gqls:"id,omitempty"`
	Name           *StringFilter  `json:"name,omitempty" gqls:"name,omitempty"`
	Status         *StringFilter  `json:"status,omitempty" gqls:"status,omitempty"`
	System         *BooleanFilter `json:"system,omitempty" gqls:"system,omitempty"`
	OrganizationID *UUIDFilter    `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	CategoryID     *UUIDFilter    `json:"categoryId,omitempty" gqls:"categoryId,omitempty"`
	ClassID        *UUIDFilter    `json:"classId,omitempty" gqls:"classId,omitempty"`
	ProgramID      *UUIDFilter    `json:"programId,omitempty" gqls:"programId,omitempty"`
	FromGradeID    *UUIDFilter    `json:"fromGradeId,omitempty" gqls:"fromGradeId,omitempty"`
	ToGradeID      *UUIDFilter    `json:"toGradeId,omitempty" gqls:"toGradeId,omitempty"`
	AND            []GradeFilter  `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []GradeFilter  `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (GradeFilter) FilterName() FilterType {
	return GradeFilterType
}

func (GradeFilter) ConnectionName() ConnectionType {
	return GradesConnectionType
}

type GradeSummaryNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}
type GradeConnectionNode struct {
	ID        string           `json:"id" gqls:"id"`
	Name      string           `json:"name" gqls:"name"`
	Status    string           `json:"status" gqls:"status"`
	System    bool             `json:"system" gqls:"system"`
	FromGrade GradeSummaryNode `json:"fromGrade" gqls:"fromGrade"`
	ToGrade   GradeSummaryNode `json:"toGrade" gqls:"toGrade"`
}

type GradesConnectionEdge struct {
	Cursor string              `json:"cursor" gqls:"cursor"`
	Node   GradeConnectionNode `json:"node" gqls:"node"`
}

type GradesConnectionResponse struct {
	TotalCount int                    `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo     `json:"pageInfo" gqls:"pageInfo"`
	Edges      []GradesConnectionEdge `json:"edges" gqls:"edges"`
}

func (scs GradesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &scs.PageInfo
}
func (gcs AmsGradeConnectionService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Grade, error) {
	condition := NewCondition(options...)

	filter := GradeFilter{
		ProgramID: &UUIDFilter{
			Operator: UUIDOperator(OperatorTypeEq),
			Value:    UUID(programID),
		},
		Status: &StringFilter{
			Operator: StringOperator(OperatorTypeEq),
			Value:    Active.String(),
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

	var pages []GradesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get grade by program failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	if len(pages) == 0 {
		log.Warn(ctx, "grade is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*Grade{}, nil
	}

	grades := make([]*Grade, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, p := range pages {
		for _, v := range p.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "grade exists",
					log.Any("grade", v.Node),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
			obj := &Grade{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			grades = append(grades, obj)
		}
	}
	return grades, nil
}
func (gcs AmsGradeConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Grade, error) {
	condition := NewCondition(options...)
	filter := GradeFilter{
		Status: &StringFilter{Operator: StringOperator(OperatorTypeEq), Value: Active.String()},
		OR: []GradeFilter{
			{OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(operator.OrgID)}},
			{System: &BooleanFilter{Operator: OperatorTypeEq, Value: true}},
		},
	}
	if condition.Status.Valid {
		filter.Status.Value = condition.Status.Status.String()
	}
	if condition.System.Valid {
		filter.System = &BooleanFilter{Operator: OperatorTypeEq, Value: condition.System.Valid}
	}

	var pages []GradesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get grade by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	if len(pages) == 0 {
		log.Warn(ctx, "grade is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*Grade{}, nil
	}

	grades := make([]*Grade, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, p := range pages {
		for _, v := range p.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Warn(ctx, "grade exists",
					log.Any("grade", v.Node),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
			obj := &Grade{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
				System: v.Node.System,
			}
			grades = append(grades, obj)
		}
	}
	return grades, nil
}

func (gcs AmsGradeConnectionService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
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
		fmt.Fprintf(sb, "q%d: gradeNode(id: $grade_id_%d) {id name status system}\n", index, index)
	}
	sb.WriteString("}")

	request := NewRequest(sb.String(), RequestToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("grade_id_%d", index), id)
	}

	data := map[string]*Grade{}
	response := &GraphQLSubResponse{
		Data: &data,
	}

	_, err = GetAmsConnection().Run(ctx, request, response)
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
