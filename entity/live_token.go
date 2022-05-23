package entity

import (
	"github.com/golang-jwt/jwt"
)

type LiveTokenType string

const (
	LiveTokenTypePreview LiveTokenType = "preview"
	LiveTokenTypeLive    LiveTokenType = "live"
)

func (t LiveTokenType) Valid() bool {
	switch t {
	case LiveTokenTypePreview, LiveTokenTypeLive:
		return true
	default:
		return false
	}
}

type MaterialType string

const (
	MaterialTypeH5P   MaterialType = "Iframe"
	MaterialTypeVideo MaterialType = "Video"
	MaterialTypeAudio MaterialType = "Audio"
	MaterialTypeImage MaterialType = "Image"
)

//Live (online class)
//Class (offline class)
//Study (homework)
//Task (task)
type LiveClassType string

const (
	LiveClassTypeInvalid LiveClassType = "invalid"
	LiveClassTypeLive    LiveClassType = "live"
	LiveClassTypeClass   LiveClassType = "class"
	LiveClassTypeStudy   LiveClassType = "study"
	LiveClassTypeTask    LiveClassType = "task"
)

type LiveTokenInfo struct {
	Name       string          `json:"name,omitempty"`
	ScheduleID string          `json:"schedule_id,omitempty"`
	IsReview   bool            `json:"is_review"`
	UserID     string          `json:"user_id,omitempty"`
	Type       LiveTokenType   `json:"type"`
	Teacher    bool            `json:"teacher"`
	RoomID     string          `json:"roomid"`
	Materials  []*LiveMaterial `json:"materials"`
	ClassType  LiveClassType   `json:"classtype"`
	OrgID      string          `json:"org_id"`

	StartAt int64 `json:"start_at"`
	EndAt   int64 `json:"end_at"`
}

type LiveMaterial struct {
	ID          string       `json:"id"`
	ContentData interface{}  `json:"-"`
	Name        string       `json:"name"`
	URL         string       `json:"url,omitempty"`
	TypeName    MaterialType `json:"__typename"`
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

type MaterialInput struct {
	ContentID  string
	ScheduleID string
	TokenType  LiveTokenType
}
