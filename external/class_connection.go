package external

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"text/template"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cache/cache"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type ClassFilter struct {
	ID                *UUIDFilter          `json:"id,omitempty" gqls:"id,omitempty"`
	Name              *StringFilter        `json:"name,omitempty" gqls:"name,omitempty"`
	Status            *StringFilter        `json:"status,omitempty" gqls:"status,omitempty"`
	OrganizationID    *UUIDFilter          `json:"organizationId,omitempty" gqls:"organizationId,omitempty"`
	AgeRangeValueFrom *AgeRangeValueFilter `json:"ageRangeValueFrom,omitempty" gqls:"ageRangeValueFrom,omitempty"`
	AgeRangeUnitFrom  *AgeRangeUnitFilter  `json:"ageRangeUnitFrom,omitempty" gqls:"ageRangeUnitFrom,omitempty"`
	AgeRangeValueTo   *AgeRangeValueFilter `json:"ageRangeValueTo,omitempty" gqls:"ageRangeValueTo,omitempty"`
	AgeRangeUnitTo    *AgeRangeUnitFilter  `json:"ageRangeUnitTo,omitempty" gqls:"ageRangeUnitTo,omitempty"`
	SchoolID          *UUIDExclusiveFilter `json:"schoolId,omitempty" gqls:"schoolId,omitempty"`
	GradeID           *UUIDFilter          `json:"gradeId,omitempty" gqls:"gradeId,omitempty"`
	SubjectID         *UUIDFilter          `json:"subjectId,omitempty" gqls:"subjectId,omitempty"`
	AcademicTermID    *UUIDFilter          `json:"academicTermId,omitempty" gqls:"academicTermId,omitempty"`
	StudentID         *UUIDFilter          `json:"studentId,omitempty" gqls:"studentId,omitempty"`
	TeacherID         *UUIDFilter          `json:"teacherId,omitempty" gqls:"teacherId,omitempty"`
	ProgramID         *UUIDFilter          `json:"programId,omitempty" gqls:"programId,omitempty"`
	AND               []ClassFilter        `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR                []ClassFilter        `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (ClassFilter) FilterName() FilterType {
	return ClassFilterType
}

func (ClassFilter) ConnectionName() ConnectionType {
	return ClassesConnectionType
}

type ClassTeachingFilter ClassFilter

func (ClassTeachingFilter) FilterName() FilterType {
	return ClassFilterType
}

func (ClassTeachingFilter) ConnectionName() ConnectionType {
	return ClassesTeachingConnectionType
}

type ClassStudyingFilter ClassFilter

func (ClassStudyingFilter) FilterName() FilterType {
	return ClassFilterType
}

func (ClassStudyingFilter) ConnectionName() ConnectionType {
	return ClassesStudyingConnectionType
}

type ClassConnectionNode struct {
	ID        string `json:"id" gqls:"id"`
	Name      string `json:"name" gqls:"name"`
	Status    string `json:"status" gqls:"status"`
	ShortCode string `json:"shortCode" gqls:"shortCode"`
}

type ClassConnectionEdge struct {
	Cursor string              `json:"cursor" gqls:"cursor"`
	Node   ClassConnectionNode `json:"node" gqls:"node"`
}

type ClassesConnectionResponse struct {
	TotalCount int                   `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo    `json:"pageInfo" gqls:"pageInfo"`
	Edges      []ClassConnectionEdge `json:"edges" gqls:"edges"`
}

func (ccs ClassesConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &ccs.PageInfo
}

type AmsClassConnectionService struct {
	AmsClassService
}

func (accs AmsClassConnectionService) GetByUserID(ctx context.Context, operator *entity.Operator, userID string, options ...APOption) ([]*Class, error) {
	classesMap, err := accs.GetByUserIDs(ctx, operator, []string{userID}, options...)
	if err != nil {
		log.Error(ctx, "GetByUserID: GetByUserIDs failed",
			log.Err(err),
			log.String("user", userID))
		return nil, err
	}
	return classesMap[userID], nil
}

func (accs AmsClassConnectionService) getByTeacherIDs(ctx context.Context, operator *entity.Operator, IDs []string, options ...APOption) (map[string][]*Class, error) {
	condition := NewCondition(options...)
	var filter ClassTeachingFilter
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

	result := make(map[string][]ClassesConnectionResponse)
	err := subPageQuery(ctx, operator, "userNode", filter, IDs, result)
	if err != nil {
		log.Error(ctx, "getByTeacherIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("user_ids", IDs))
		return nil, err
	}

	classesMap := make(map[string][]*Class)
	for k, pages := range result {
		if len(pages) == 0 {
			log.Warn(ctx, "getByTeacherIDs: class is empty",
				log.String("school", k),
				log.Any("filter", filter),
				log.Any("operator", operator))
		}
		classesMap[k] = accs.pageNodes(ctx, operator, pages)
	}
	return classesMap, nil
}

func (accs AmsClassConnectionService) getByStudentIDs(ctx context.Context, operator *entity.Operator, IDs []string, options ...APOption) (map[string][]*Class, error) {
	condition := NewCondition(options...)
	var filter ClassStudyingFilter
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

	result := make(map[string][]ClassesConnectionResponse)
	err := subPageQuery(ctx, operator, "userNode", filter, IDs, result)
	if err != nil {
		log.Error(ctx, "getByStudentIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("user_ids", IDs))
		return nil, err
	}

	classesMap := make(map[string][]*Class)
	for k, pages := range result {
		if len(pages) == 0 {
			log.Warn(ctx, "getByStudentIDs: class is empty",
				log.String("school", k),
				log.Any("filter", filter),
				log.Any("operator", operator))
		}
		classesMap[k] = accs.pageNodes(ctx, operator, pages)
	}
	return classesMap, nil
}

func (accs AmsClassConnectionService) GetByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string, options ...APOption) (map[string][]*Class, error) {
	IDs := utils.SliceDeduplicationExcludeEmpty(userIDs)
	var wg sync.WaitGroup
	wg.Add(2)
	var tErr, sErr error
	var teachClasses, studyClasses map[string][]*Class
	go func() {
		defer wg.Done()
		teachClasses, tErr = accs.getByTeacherIDs(ctx, operator, IDs, options...)
	}()
	go func() {
		defer wg.Done()
		studyClasses, sErr = accs.getByStudentIDs(ctx, operator, IDs, options...)
	}()
	wg.Wait()
	if tErr != nil {
		log.Error(ctx, "GetByUserIDs: subPageQuery failed",
			log.Err(tErr),
			log.Strings("user_ids", userIDs))
		return nil, tErr
	}
	if sErr != nil {
		log.Error(ctx, "GetByUserIDs: subPageQuery failed",
			log.Err(sErr),
			log.Strings("user_ids", userIDs))
		return nil, sErr
	}
	classesMap := make(map[string][]*Class)
	for _, k := range IDs {
		if teachClasses[k] == nil && studyClasses[k] != nil {
			classesMap[k] = studyClasses[k]
		}
		if teachClasses[k] != nil && studyClasses[k] == nil {
			classesMap[k] = teachClasses[k]
		}
		if teachClasses[k] != nil && studyClasses[k] != nil {
			classesMap[k] = teachClasses[k]
			classesMap[k] = append(classesMap[k], studyClasses[k]...)
		}
		if teachClasses[k] == nil && studyClasses[k] == nil {
			classesMap[k] = []*Class{}
		}
	}
	return classesMap, nil
}
func (accs AmsClassConnectionService) GetByOrganizationIDs(ctx context.Context, operator *entity.Operator, orgIDs []string, options ...APOption) (map[string][]*Class, error) {
	condition := NewCondition(options...)
	var filter ClassFilter
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

	if condition.Status.Valid && condition.Status.Status != Ignore {
		filter.Status.Value = condition.Status.Status.String()
	} else if !condition.Status.Valid {
		filter.Status.Value = Active.String()
	}

	result := make(map[string][]ClassesConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(orgIDs)
	err := subPageQuery(ctx, operator, "organizationNode", filter, IDs, result)
	if err != nil {
		log.Error(ctx, "GetByOrganizationIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("organization_ids", orgIDs))
		return nil, err
	}

	classesMap := make(map[string][]*Class)
	for k, pages := range result {
		if len(pages) == 0 {
			log.Warn(ctx, "GetByOrganizationIDs: class is empty",
				log.String("school", k),
				log.Any("filter", filter),
				log.Any("operator", operator))
		}
		classesMap[k] = accs.pageNodes(ctx, operator, pages)
	}
	return classesMap, nil
}

func (accs AmsClassConnectionService) pageNodes(ctx context.Context, operator *entity.Operator, pages []ClassesConnectionResponse) []*Class {
	if len(pages) == 0 {
		log.Warn(ctx, "pageNodes is empty",
			log.Any("operator", operator))
		return []*Class{}
	}
	classes := make([]*Class, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, edge := range page.Edges {
			if _, ok := exists[edge.Node.ID]; ok {
				log.Warn(ctx, "pageNodes: class exist",
					log.Any("class", edge.Node),
					log.Any("operator", operator))
				continue
			}
			exists[edge.Node.ID] = true
			class := Class{
				ID:     edge.Node.ID,
				Name:   edge.Node.Name,
				Status: APStatus(edge.Node.Status),
			}
			classes = append(classes, &class)
		}
	}
	return classes
}

func (accs AmsClassConnectionService) NewClassFilter(ctx context.Context, operator *entity.Operator, options ...APOption) *ClassFilter {
	condition := NewCondition(options...)
	var filter ClassFilter
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
	return &filter
}
func (accs AmsClassConnectionService) GetBySchoolIDs(ctx context.Context, operator *entity.Operator, schoolIDs []string, options ...APOption) (map[string][]*Class, error) {
	filter := accs.NewClassFilter(ctx, operator, options...)
	result := make(map[string][]ClassesConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(schoolIDs)
	err := subPageQuery(ctx, operator, "schoolNode", filter, IDs, result)
	if err != nil {
		log.Error(ctx, "GetBySchoolIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("school_ids", schoolIDs))
		return nil, err
	}

	classesMap := make(map[string][]*Class)
	for k, pages := range result {
		if len(pages) == 0 {
			log.Warn(ctx, "GetBySchoolIDs: class is empty",
				log.String("school", k),
				log.Any("filter", filter),
				log.Any("operator", operator))
		}
		classesMap[k] = accs.pageNodes(ctx, operator, pages)
	}
	return classesMap, nil
}

func (accs AmsClassConnectionService) GetOnlyUnderOrgClasses(ctx context.Context, operator *entity.Operator, orgID string) ([]*NullableClass, error) {

	filter := accs.NewClassFilter(ctx, operator)
	filter.OrganizationID = &UUIDFilter{Operator: UUIDOperator(OperatorTypeEq), Value: UUID(orgID)}
	filter.SchoolID = &UUIDExclusiveFilter{Operator: UUIDExclusiveOperator(OperatorTypeIsNull)}

	var pages []ClassesConnectionResponse
	err := pageQuery(ctx, operator, filter, &pages)
	if err != nil {
		log.Error(ctx, "get only under org class failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("filter", filter))
		return nil, err
	}
	if len(pages) == 0 {
		log.Debug(ctx, "class is empty",
			log.Any("operator", operator),
			log.Any("filter", filter))
		return []*NullableClass{}, nil
	}
	classes := make([]*NullableClass, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, v := range page.Edges {
			if _, ok := exists[v.Node.ID]; ok {
				log.Debug(ctx, "class exists",
					log.Any("class", v.Node),
					log.Any("operator", operator),
					log.Any("filter", filter))
				continue
			}
			exists[v.Node.ID] = true
			class := Class{
				ID:     v.Node.ID,
				Name:   v.Node.Name,
				Status: APStatus(v.Node.Status),
			}
			nullableClass := NullableClass{Class: class, StrID: class.ID, Valid: true}
			classes = append(classes, &nullableClass)
		}
	}

	return classes, nil
}

func (accs AmsClassConnectionService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	raw := `query{
		{{range $i, $e := .}}
		index_{{$i}}: classNode(id: "{{$e}}"){
			id
			name
			status
		  }
		{{end}}
	}`

	temp, err := template.New("Classes").Parse(raw)
	if err != nil {
		log.Error(ctx, "temp error", log.String("raw", raw), log.Err(err))
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	buf := bytes.Buffer{}
	err = temp.Execute(&buf, _ids)
	if err != nil {
		log.Error(ctx, "temp execute failed", log.String("raw", raw), log.Err(err))
		return nil, err
	}
	req := NewRequest(buf.String(), RequestToken(operator.Token))
	payload := make(map[string]*Class, len(ids))
	res := GraphQLSubResponse{
		Data: &payload,
	}

	_, err = GetAmsConnection().Run(ctx, req, &res)
	if err != nil {
		log.Error(ctx, "Run error", log.String("q", buf.String()), log.Any("res", res), log.Err(err))
		return nil, err
	}
	if len(res.Errors) > 0 {
		log.Error(ctx, "Res error", log.String("q", buf.String()), log.Any("res", res), log.Err(res.Errors))
		return nil, res.Errors
	}
	var classes []cache.Object
	for index := range ids {
		class := payload[fmt.Sprintf("index_%d", indexMapping[index])]
		if class == nil {
			classes = append(classes, &NullableClass{
				Valid: false,
				StrID: ids[index],
			})
		} else {
			classes = append(classes, &NullableClass{
				Class: *class,
				Valid: true,
				StrID: ids[index],
			})
		}
	}

	log.Info(ctx, "get classes by ids success",
		log.Strings("ids", ids),
		log.Any("classes", classes))

	return classes, nil
}
