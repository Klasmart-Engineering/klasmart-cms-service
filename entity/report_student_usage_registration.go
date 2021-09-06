package entity

type StudentUsageRegistration struct {
	Registered int `json:"registered"`
	Created    int `json:"created"`
	NoCreate   int `json:"no_create"`
}
