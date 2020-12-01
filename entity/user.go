package entity

import (
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"

	"github.com/dgrijalva/jwt-go"
)

type User struct {
	UserID   string `gorm:"column:user_id"`
	UserName string `gorm:"column:user_name"`
	Email    string `gorm:"column:email"`
	Phone    string `gorm:"column:phone"`

	Secret string `gorm:"column:secret"`
	Salt   string `gorm:"salt"`

	Avatar   string `gorm:"avatar"`
	Gender   string `gorm:"gender"`
	Birthday int64  `gorm:"birthday"`

	CreateAt int64 `gorm:"column:create_at"`
	UpdateAt int64 `gorm:"column:update_at"`
	DeleteAt int64 `gorm:"column:delete_at"`
}

func (User) TableName() string {
	return "users"
}

func (user User) GetID() interface{} {
	return user.UserID
}

const ValidDays = 30

func (user User) Token() (string, error) {
	now := time.Now()
	claim := &jwt.StandardClaims{
		Audience:  "Kidsloop",
		Id:        user.UserID,
		ExpiresAt: now.Add(time.Hour * 24 * ValidDays).Unix(),
		IssuedAt:  now.Add(-30 * time.Second).Unix(),
		Issuer:    "Kidsloop_cn",
		NotBefore: 0,
		Subject:   "authorization",
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS512, claim).SignedString(config.Get().KidsloopCNLoginConfig.PrivateKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

func NewUserFromToken(token string) (*User, error) {
	var claim jwt.StandardClaims
	_, err := jwt.ParseWithClaims(token, &claim, func(t *jwt.Token) (interface{}, error) {
		return config.Get().KidsloopCNLoginConfig.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	return &User{
		UserID: claim.Id,
	}, nil
}

func (user User) VerifyPassword(password string) (bool, error) {
	// TODO
	return false, nil
}
