package gql

import "errors"

var (
	ErrUnsupported = errors.New("unsupported")
	ErrUnMatch     = errors.New("match failed")
)

type OperatorType string

type UUID string
type UUIDOperator OperatorType
type StringOperator OperatorType
type NumberOrDateOperator OperatorType

const (
	OperatorTypeContains OperatorType = "contains"
	OperatorTypeEq       OperatorType = "eq"
	OperatorTypeNeq      OperatorType = "neq"
	OperatorTypeGt       OperatorType = "gt"
	OperatorTypeGte      OperatorType = "gte"
	OperatorTypeLt       OperatorType = "lt"
	OperatorTypeLte      OperatorType = "lte"
)

func (o OperatorType) String() string {
	return string(o)
}

type UUIDFilter struct {
	Operator OperatorType `gqls:"operator,noquoted"`
	Value    UUID         `gqls:"value"`
}

type UserFilter struct {
	UserID UUIDFilter `gqls:"userId"`
}

type StringFilter struct {
	Operator        OperatorType `gqls:"operator,noquoted"`
	Value           string       `gqls:"value"`
	CaseInsensitive bool         `gqls:"caseInsensitive"`
}

type BooleanFilter struct {
	Operator OperatorType `gqls:"operator,noquoted"`
	Value    bool         `gqls:"value"`
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

type ConnectionDirection string

const (
	FORWARD  ConnectionDirection = "FORWARD"
	BACKWARD ConnectionDirection = "BACKWARD"
)

const (
	PagerDirection string = "direction"
	PagerCursor    string = "cursor"
	PagerCount     string = "count"
)

type FilterOfType string

const (
	OrganizationsConnectionType FilterOfType = "organizationsConnection"
	ProgramsConnectionType      FilterOfType = "programsConnection"
	SubjectsConnectionType      FilterOfType = "subjectsConnection"
	CategoriesConnectionType    FilterOfType = "categoriesConnection"
	SubcategoriesConnectionType FilterOfType = "subcategoriesConnection"
	GradesConnectionType        FilterOfType = "gradesConnection"
	AgeRangesConnectionType     FilterOfType = "ageRangesConnection"
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

func (pager *ConnectionPageInfo) HasNext(direction ConnectionDirection) bool {
	if pager == nil {
		return true
	}
	if direction == FORWARD && pager.HasNextPage {
		return true
	}
	if direction == BACKWARD && pager.HasPreviousPage {
		return true
	}
	return false
}

func (pager *ConnectionPageInfo) Pager(direction ConnectionDirection, count int) map[string]interface{} {
	var cursor string
	if pager == nil {
		return map[string]interface{}{
			PagerDirection: FORWARD,
			PagerCursor:    "",
			PagerCount:     PageDefaultCount,
		}
	}
	if direction == FORWARD {
		cursor = pager.EndCursor
	}
	if direction == BACKWARD {
		cursor = pager.StartCursor
	}
	if count > PageDefaultCount {
		count = PageDefaultCount
	}
	return map[string]interface{}{
		PagerDirection: direction,
		PagerCursor:    cursor,
		PagerCount:     count,
	}
}

type ConnectionFilter interface {
	FilterType() FilterOfType
}

type ConnectionResponse interface {
	OrganizationsConnectionResponse |
		ProgramsConnectionResponse |
		SubjectsConnectionResponse |
		CategoriesConnectionResponse |
		SubcategoriesConnectionResponse |
		GradesConnectionResponse |
		AgesConnectionResponse

	GetPageInfo() *ConnectionPageInfo
}
