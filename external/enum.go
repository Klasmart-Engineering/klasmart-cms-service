package external

type APStatus string

const (
	Active   APStatus = "active"
	Inactive APStatus = "inactive"
)

type NullAPStatus struct {
	Status APStatus
	Valid  bool
}
