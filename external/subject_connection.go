package external

import (
	"context"
	"fmt"
	"strings"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cache/cache"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type SubjectFilter struct {
	ID             *UUIDFilter     `json:"id,omitempty" gqls:"id,omitempty"`
	Name           *StringFilter   `json:"name,omitempty" gqls:"name,omitempty"`
	Status         *StringFilter   `json:"status,omitempty" gqls:"status,omitempty"`
	System         *BooleanFilter  `json:"system,omitempty" gqls:"system,omitempty"`
	OrganizationID *UUIDFilter     `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	CategoryId     *UUIDFilter     `json:"categoryId,omitempty" gqls:"categoryId,omitempty"`
	ClassID        *UUIDFilter     `json:"classId,omitempty" gqls:"classId,omitempty"`
	ProgramID      *UUIDFilter     `json:"programId,omitempty" gqls:"programId,omitempty"`
	AND            []SubjectFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR             []SubjectFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (SubjectFilter) FilterName() FilterType {
	return SubjectFilterType
}

func (SubjectFilter) ConnectionName() ConnectionType {
	return SubjectsConnectionType
}

type SubjectConnectionNode struct {
	ID     string `json:"id" gqls:"id"`
	Name   string `json:"name" gqls:"name"`
	Status string `json:"status" gqls:"status"`
	System bool   `json:"system" gqls:"system"`
}

type SubjectsConnectionEdge struct {
	Cursor string                `json:"cursor" gqls:"cursor"`
	Node   SubjectConnectionNode `json:"node" gqls:"node"`
}

type SubjectsConnectionResponse struct {
	TotalCount int                      `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo       `json:"pageInfo" gqls:"pageInfo"`
	Edges      []SubjectsConnectionEdge `json:"edges" gqls:"edges"`
}

func (scr SubjectsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &scr.PageInfo
}

type AmsSubjectConnectionService struct {
	AmsSubjectService
}

func (scs AmsSubjectConnectionService) pageNodes(ctx context.Context, operator *entity.Operator, pages []SubjectsConnectionResponse) []*Subject {
	if len(pages) == 0 {
		log.Warn(ctx, "pageNodes is empty",
			log.Any("operator", operator))
		return []*Subject{}
	}
	subjects := make([]*Subject, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, edge := range page.Edges {
			if _, ok := exists[edge.Node.ID]; ok {
				log.Warn(ctx, "pageNodes: subcategory exist",
					log.Any("subcategory", edge.Node),
					log.Any("operator", operator))
				continue
			}
			exists[edge.Node.ID] = true
			obj := &Subject{
				ID:     edge.Node.ID,
				Name:   edge.Node.Name,
				Status: APStatus(edge.Node.Status),
				System: edge.Node.System,
			}
			subjects = append(subjects, obj)
		}
	}
	return subjects
}

func (scs AmsSubjectConnectionService) NewSubjectFilter(ctx context.Context, operator *entity.Operator, options ...APOption) *SubjectFilter {
	condition := NewCondition(options...)
	var filter SubjectFilter
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

func (scs AmsSubjectConnectionService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Subject, error) {
	filter := scs.NewSubjectFilter(ctx, operator, options...)
	filter.ProgramID = &UUIDFilter{
		Operator: UUIDOperator(OperatorTypeEq),
		Value:    UUID(programID),
	}

	var pages []SubjectsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get subject by program failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}

	subjects := scs.pageNodes(ctx, operator, pages)
	return subjects, nil
}
func (scs AmsSubjectConnectionService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Subject, error) {
	filter := scs.NewSubjectFilter(ctx, operator, options...)
	filter.OR = []SubjectFilter{
		{OrganizationID: &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(operator.OrgID)}},
		{System: &BooleanFilter{Operator: OperatorTypeEq, Value: true}},
	}
	var pages []SubjectsConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get subject by organization failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}

	subjects := scs.pageNodes(ctx, operator, pages)
	return subjects, nil
}

func (scs AmsSubjectConnectionService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)
	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$subject_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: subjectNode(id: $subject_id_%d) {id name status system}\n", index, index)
	}
	sb.WriteString("}")

	request := NewRequest(sb.String(), RequestToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("subject_id_%d", index), id)
	}

	data := map[string]*Subject{}
	response := &GraphQLSubResponse{
		Data: &data,
	}

	_, err = GetAmsConnection().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get subjects by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get subjects by ids failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Strings("ids", ids))
		return nil, response.Errors
	}

	subjects := make([]cache.Object, 0, len(data))
	for index := range ids {
		subject := data[fmt.Sprintf("q%d", indexMapping[index])]
		if subject == nil {
			log.Debug(ctx, "subject not found", log.String("id", ids[index]))
			continue
		}
		subjects = append(subjects, subject)
	}

	log.Info(ctx, "get subjects by ids success",
		log.Strings("ids", ids),
		log.Any("subjects", subjects))

	return subjects, nil
}
