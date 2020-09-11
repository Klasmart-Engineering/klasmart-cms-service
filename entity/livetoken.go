package entity

import "github.com/dgrijalva/jwt-go"

const (
	LiveTokenTypePreview = "preview"
	LiveTokenTypeLive    = "live"
)

const (
	MaterialTypeH5P = "Iframe"
)
const (
	LiveTokenEnvPath = "v1"
)

const ValidDays = 30

type LiveTokenInfo struct {
	Name       string          `json:"name,omitempty"`
	ScheduleID string          `json:"schedule_id,omitempty"`
	UserID     string          `json:"user_id,omitempty"`
	Type       string          `json:"type"`
	Teacher    bool            `json:"teacher"`
	Materials  []*LiveMaterial `json:"materials,omitempty"`
	//RoomID     string          `json:"roomid,omitempty"`
}

type LiveMaterial struct {
	Name     string `json:"name"`
	URL      string `json:"url,omitempty"`
	TypeName string `json:"__typename"`
}
type LiveTokenShort struct {
	ID   string
	Name string
}

type LiveTokenClaims struct {
	*jwt.StandardClaims
	LiveTokenInfo
}
