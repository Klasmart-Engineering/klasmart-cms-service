package entity

import "github.com/dgrijalva/jwt-go"

type LiveTokenType string

const (
	LiveTokenTypePreview LiveTokenType = "preview"
	LiveTokenTypeLive    LiveTokenType = "live"
)

type MaterialType string

const (
	MaterialTypeH5P MaterialType = "Iframe"
)

const (
	LiveTokenEnvPath = "v1"
)

type LiveTokenInfo struct {
	Name       string          `json:"name,omitempty"`
	ScheduleID string          `json:"schedule_id,omitempty"`
	UserID     string          `json:"user_id,omitempty"`
	Type       string          `json:"type"`
	Teacher    bool            `json:"teacher"`
	RoomID     string          `json:"roomid"`
	Materials  []*LiveMaterial `json:"materials,omitempty"`
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

type LiveTokenView struct {
	Token string `json:"token"`
}
