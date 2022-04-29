package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"net/http"
)

type Iterator interface {
	HasNext() bool
	Next(ctx context.Context, operator *entity.Operator, query string, variables map[string]interface{}, res chlorine.Response, f func(*chlorine.Response) (Iterator, error)) (interface{}, error)
}

type Direction string

var (
	Forward  Direction = "FORWARD"
	BackWard Direction = "BACKWARD"
)

type UUID string
type UUIDOperator OperatorType
type StringOperator OperatorType
type UUIDExclusiveOperator OperatorType

type UUIDFilter struct {
	Operator UUIDOperator `json:"operator" gqls:"operator,noquoted"`
	Value    UUID         `json:"value" gqls:"value"`
}

type UUIDExclusiveFilter struct {
	Operator UUIDExclusiveOperator `json:"operator" gqls:"operator,noquoted"`
	Value    *UUID                 `json:"value,omitempty" gqls:"value"`
}

type StringFilter struct {
	Operator        StringOperator `json:"operator" gqls:"operator,noquoted"`
	Value           string         `json:"value" gqls:"value"`
	CaseInsensitive bool           `json:"caseInsensitive" gqls:"caseInsensitive"`
}

type BooleanFilter struct {
	Operator OperatorType `json:"operator" gqls:"operator,noquoted"`
	Value    bool         `json:"value" gqls:"value"`
}

type Branding struct {
	IconImageURL string `json:"iconImageURL,omitempty" gqls:"iconImageURL,omitempty"`
	PrimaryColor string `json:"primaryColor,omitempty" gqls:"primaryColor,omitempty"`
}
type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

func (pageInfo *PageInfo) HasNext() bool {
	if pageInfo == nil {
		// the beginning
		return true
	}
	return pageInfo.HasNextPage
}

func (pageInfo *PageInfo) ForwardCursor() string {
	if pageInfo != nil && pageInfo.HasNextPage {
		return pageInfo.EndCursor
	}
	return ""
}

func (pageInfo *PageInfo) BackwardCursor() string {
	if pageInfo != nil && pageInfo.HasPreviousPage {
		return pageInfo.StartCursor
	}
	return ""
}

type UserConnectionEdge struct {
	Node UserConnectionNode `json:"node"`
}

type TeachersConnection struct {
	PageInfo PageInfo             `json:"pageInfo"`
	Edges    []UserConnectionEdge `json:"edges"`
}

func (tc *TeachersConnection) HasNext() bool {
	if tc == nil {
		return false
	}
	return tc.PageInfo.HasNext()
}

func (tc *TeachersConnection) Next(ctx context.Context, operator *entity.Operator, query string, variables map[string]interface{}, res chlorine.Response, f func(response *chlorine.Response) (Iterator, error)) (interface{}, error) {
	//var res chlorine.Response
	err := do(ctx, operator, query, variables, &res)
	if err != nil {
		return nil, err
	}
	it, err := f(&res)
	if err != nil {
		log.Error(ctx, "Next: f failed",
			log.Any("data", res))
		return nil, err
	}
	teacherConnection, ok := it.(*TeachersConnection)
	if !ok {
		err = constant.ErrAssertFailed
		log.Error(ctx, "Next: assert failed",
			log.Err(err),
			log.Any("data", res))
		return nil, err
	}
	log.Debug(ctx, "Next: success",
		log.String("query", query),
		log.Any("variables", variables),
		log.Any("classesConnection", teacherConnection))
	tc.PageInfo = teacherConnection.PageInfo
	tc.Edges = teacherConnection.Edges
	return tc.Edges, err
}

type StudentsConnection struct {
	PageInfo PageInfo             `json:"pageInfo"`
	Edges    []UserConnectionEdge `json:"edges"`
}

func (sc *StudentsConnection) HasNext() bool {
	if sc == nil {
		return false
	}
	return sc.PageInfo.HasNext()
}

func (sc *StudentsConnection) Next(ctx context.Context, operator *entity.Operator, query string, variables map[string]interface{}, res chlorine.Response, f func(response *chlorine.Response) (Iterator, error)) (interface{}, error) {
	//var res chlorine.Response
	err := do(ctx, operator, query, variables, &res)
	if err != nil {
		return nil, err
	}
	it, err := f(&res)
	if err != nil {
		log.Error(ctx, "Next: f failed",
			log.Any("data", res))
		return nil, err
	}
	studentsConnection, ok := it.(*StudentsConnection)
	if !ok {
		err = constant.ErrAssertFailed
		log.Error(ctx, "Next: assert failed",
			log.Err(err),
			log.Any("data", res))
		return nil, err
	}
	log.Debug(ctx, "Next: success",
		log.String("query", query),
		log.Any("variables", variables),
		log.Any("classesConnection", studentsConnection))
	sc.PageInfo = studentsConnection.PageInfo
	sc.Edges = studentsConnection.Edges
	return sc.Edges, err
}

type ClassesConnectionNode struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Status             APStatus           `json:"status"`
	StudentsConnection StudentsConnection `json:"studentsConnection"`
	TeachersConnection TeachersConnection `json:"teachersConnection"`
}

type ClassesConnectionEdge struct {
	Node ClassesConnectionNode `json:"node"`
}

type ClassesConnection struct {
	PageInfo *PageInfo               `json:"pageInfo"`
	Edges    []ClassesConnectionEdge `json:"edges"`
}

func (cc *ClassesConnection) HasNext() bool {
	return cc.PageInfo.HasNext()
}

func (cc *ClassesConnection) Next(ctx context.Context, operator *entity.Operator, query string, variables map[string]interface{}, res chlorine.Response, f func(response *chlorine.Response) (Iterator, error)) (interface{}, error) {
	//var res chlorine.Response
	err := do(ctx, operator, query, variables, &res)
	if err != nil {
		return nil, err
	}

	it, err := f(&res)
	if err != nil {
		log.Error(ctx, "Next: f failed",
			log.Any("data", res))
		return nil, err
	}
	classesConnection, ok := it.(*ClassesConnection)
	if !ok {
		err = constant.ErrAssertFailed
		log.Error(ctx, "Next: assert failed",
			log.Err(err),
			log.Any("data", res))
		return nil, err
	}

	log.Debug(ctx, "Next: success",
		log.String("query", query),
		log.Any("variables", variables),
		log.Any("classesConnection", classesConnection))

	cc.PageInfo = classesConnection.PageInfo
	cc.Edges = classesConnection.Edges
	return cc.Edges, err
}

func do(ctx context.Context, operator *entity.Operator, query string, variables map[string]interface{}, response *chlorine.Response) error {
	request := chlorine.NewRequest(query, chlorine.ReqToken(operator.Token))
	for k, v := range variables {
		request.Var(k, v)
	}

	statusCode, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "do: run failed",
			log.Err(err),
			log.String("query", query),
			log.Any("variables", variables))
		return err
	}
	if statusCode != http.StatusOK {
		err = &entity.ExternalError{
			Err:  constant.ErrAmsHttpFailed,
			Type: constant.InternalErrorTypeAms,
		}
		log.Warn(ctx, "do: run failed",
			log.Int("status_code", statusCode),
			log.String("query", query),
			log.Any("variables", variables))
		return err
	}

	log.Debug(ctx, "do success",
		log.String("query", query),
		log.Any("variables", variables),
		log.Any("response", response))

	return nil
}

type AgeRangeValueFilter struct {
	Operator OperatorType `gqls:"operator,noquoted"`
	Value    int          `gqls:"value"`
}

type AgeRangeUnitFilter struct {
	Operator OperatorType `gqls:"operator,noquoted"`
	Value    int          `gqls:"value"`
}

type AgeRangeUnit string

const (
	AgeRangeUnitYear  AgeRangeUnit = "year"
	AgeRangeUnitMonth AgeRangeUnit = "month"
)

type AgeRangeValue struct {
	Value int          `gqls:"value"`
	Unit  AgeRangeUnit `gqls:"AgeRangeUnit"`
}
type AgeRangeTypeFilter struct {
	Operator OperatorType  `gqls:"operator,noquoted"`
	Value    AgeRangeValue `gqls:"value"`
}

const (
	PagerDirection string = "direction"
	DirectionArgs  string = "directionArgs"
	PagerCursor    string = "cursor"
	PagerCount     string = "count"
)

type FilterType string

const (
	OrganizationFilterType FilterType = "OrganizationFilter"
	SchoolFilterType       FilterType = "SchoolFilter"
	UserFilterType         FilterType = "UserFilter"
	ClassFilterType        FilterType = "ClassFilter"
	RoleFilterType         FilterType = "RoleFilter"
	PermissionFilterType   FilterType = "PermissionFilter"
	ProgramFilterType      FilterType = "ProgramFilter"
	SubjectFilterType      FilterType = "SubjectFilter"
	CategoryFilterType     FilterType = "CategoryFilter"
	SubcategoryFilterType  FilterType = "SubcategoryFilter"
	GradeFilterType        FilterType = "GradeFilter"
	AgeRangeFilterType     FilterType = "AgeRangeFilter"
)

type ConnectionType string

const (
	OrganizationsConnectionType ConnectionType = "organizationsConnection"
	SchoolsConnectionType       ConnectionType = "schoolsConnection"
	UsersConnectionType         ConnectionType = "usersConnection"
	ClassesConnectionType       ConnectionType = "classesConnection"
	RolesConnectionType         ConnectionType = "rolesConnection"
	PermissionsConnectionType   ConnectionType = "permissionsConnection"
	ProgramsConnectionType      ConnectionType = "programsConnection"
	SubjectsConnectionType      ConnectionType = "subjectsConnection"
	CategoriesConnectionType    ConnectionType = "categoriesConnection"
	SubcategoriesConnectionType ConnectionType = "subcategoriesConnection"
	GradesConnectionType        ConnectionType = "gradesConnection"
	AgeRangesConnectionType     ConnectionType = "ageRangesConnection"
)

const (
	PageDefaultCount = 50
)

type ConnectionPageInfo struct {
	HasNextPage     bool   `json:"hasNextPage" gqls:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage" gqls:"hasPreviousPage"`
	StartCursor     string `json:"startCursor" gqls:"startCursor"`
	EndCursor       string `json:"endCursor" gqls:"endCursor"`
}

func (pager *ConnectionPageInfo) HasNext(direction Direction) bool {
	if pager == nil {
		return true
	}
	if direction == Forward && pager.HasNextPage {
		return true
	}
	if direction == BackWard && pager.HasPreviousPage {
		return true
	}
	return false
}

func (pager *ConnectionPageInfo) Pager(direction Direction, count int) map[string]interface{} {
	var cursor string
	if pager == nil {
		return map[string]interface{}{
			PagerDirection: Forward,
			DirectionArgs: map[string]interface{}{
				PagerCursor: "",
				PagerCount:  PageDefaultCount,
			},
		}
	}
	if direction == Forward {
		cursor = pager.EndCursor
	}
	if direction == BackWard {
		cursor = pager.StartCursor
	}
	if count > PageDefaultCount {
		count = PageDefaultCount
	}
	return map[string]interface{}{
		PagerDirection: direction,
		DirectionArgs: map[string]interface{}{
			PagerCursor: cursor,
			PagerCount:  count,
		},
	}
}

type ConnectionFilter interface {
	FilterName() FilterType
	ConnectionName() ConnectionType
}

type ConnectionResponse interface {
	OrganizationsConnectionResponse |
	ProgramsConnectionResponse |
	SubjectsConnectionResponse |
	CategoriesConnectionResponse |
	SubcategoriesConnectionResponse |
	GradesConnectionResponse |
	AgesConnectionResponse |
	ClassesConnectionResponse |
	SchoolsConnectionResponse |
	UsersConnectionResponse |
	SchoolMembershipsConnectionResponse

	GetPageInfo() *ConnectionPageInfo
}

type TemplateArguments struct {
	//Direction      bool
	DirectionArgs  bool
	Filter         bool
	Sort           bool
	FilterName     string
	ConnectionName string
	ResultString   string
}
