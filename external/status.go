package external

// APStatus academic profile status
type APStatus string
type JoinType string
type OperatorType string

const (
	Active   APStatus = "active"
	Inactive APStatus = "inactive"
)

const (
	IsTeaching JoinType = "teaching"
	IsStudy    JoinType = "studying"
)

const (
	OperatorTypeContains OperatorType = "contains"
	OperatorTypeEq       OperatorType = "eq"
	OperatorTypeNeq      OperatorType = "neq"
	OperatorTypeIsNull   OperatorType = "isNull"
)

func (o OperatorType) String() string {
	return string(o)
}

func (j JoinType) string() string {
	return string(j)
}

func (s APStatus) String() string {
	return string(s)
}

func (s APStatus) Valid() bool {
	switch s {
	case Active, Inactive:
		return true
	default:
		return false
	}
}

type NullAPStatus struct {
	Status APStatus
	Valid  bool
}
