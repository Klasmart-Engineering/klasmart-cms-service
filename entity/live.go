package entity

type LiveToken struct {
	ScheduleID  string
	TokenType   string
	StudentName string
	DomainX     string
	EnvPath     string
}
type LiveContentInfo struct {
	Name      string          `json:"name,omitempty"`
	RoomID    string          `json:"roomid,omitempty"`
	UserID    string          `json:"userid,omitempty"`
	Materials []*LiveMaterial `json:"materials,omitempty"`
}

type LiveMaterial struct {
	Name     string `json:"name"`
	URL      string `json:"url,omitempty"`
	TypeName string `json:"__typename"`
}

type H5pFileInfo struct {
	Name     string
	FileURL  string
	TypeName string
}
