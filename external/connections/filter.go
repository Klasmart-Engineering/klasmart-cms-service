package connections

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

//type DirectionArgs struct {
//	Count  int    `json:"__count__"`
//	Cursor string `json:"__cursor__"`
//}

type Pager map[string]interface{}

const (
	PagerDirection string = "direction"
	PagerCursor    string = "cursor"
	PagerCount     string = "count"
)

type ConnectionFilter interface {
	ProgramFilter | OrganizationFilter
	FilterType() FilterOfType
}

type FilterOfType string

const (
	ProgramsConnectionType      FilterOfType = "programsConnection"
	OrganizationsConnectionType FilterOfType = "organizationsConnection"
)
