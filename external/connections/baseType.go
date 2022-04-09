package connections

import "errors"

var (
	ErrUnsupported = errors.New("unsupported")
	ErrMatchFailed = errors.New("match failed")
)

type OperatorType string

type UUID string
type UUIDOperator OperatorType
type StringOperator OperatorType

const (
	OperatorTypeContains OperatorType = "__contains__"
	OperatorTypeEq       OperatorType = "__eq__"
	OperatorTypeNeq      OperatorType = "__neq__"
)

func (o OperatorType) String() string {
	return string(o)
}

type UUIDFilter struct {
	Operator OperatorType `json:"__operator__"`
	Value    UUID         `json:"__value__"`
}

type UserFilter struct {
	UserID UUIDFilter `json:"__userId__"`
}

type StringFilter struct {
	Operator        OperatorType `json:"__operator__"`
	Value           string       `json:"__value__"`
	CaseInsensitive bool         `json:"__caseInsensitive__"`
}

type BooleanFilter struct {
	Operator OperatorType `json:"__operator__"`
	Value    bool         `json:"__value__"`
}

type AgeRangeUnit string

const (
	AgeRangeUnitYear  AgeRangeUnit = "__year__"
	AgeRangeUnitMonth AgeRangeUnit = "__month__"
)

type AgeRangeValue struct {
	Value int          `json:"__value__"`
	Unit  AgeRangeUnit `json:"__AgeRangeUnit__"`
}
type AgeRangeTypeFilter struct {
	Operator OperatorType  `json:"__operator__"`
	Value    AgeRangeValue `json:"__value__"`
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
	ProgramsConnectionType      FilterOfType = "programsConnection"
	OrganizationsConnectionType FilterOfType = "organizationsConnection"
	SubcategoriesConnectionType FilterOfType = "subcategoriesConnection"
)

const (
	PageDefaultCount = 50
)

type ConnectionPageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

func (pager *ConnectionPageInfo) HasNext(direction ConnectionDirection) bool {
	if pager == nil {
		return true
	}
	if pager.HasPreviousPage || pager.HasNextPage {
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
