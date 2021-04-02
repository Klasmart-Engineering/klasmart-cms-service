package external

// APStatus academic profile status
type APStatus string

const (
	Active   APStatus = "active"
	Inactive APStatus = "inactive"
)

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
