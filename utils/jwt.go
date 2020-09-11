package utils

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"sync"
)

var jwtOnce sync.Once
var (
	prvKey []byte
	pubKey []byte
)

func InitJwtKey() {
	c := config.Get()
	prvKey = []byte(c.Auth.PrivateKey)
	pubKey = []byte(c.Auth.PublicKey)
}

func CreateJWT(ctx context.Context, claims jwt.Claims) (string, error) {
	jwtOnce.Do(InitJwtKey)
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	key, err := jwt.ParseRSAPrivateKeyFromPEM(prvKey)
	if err != nil {
		log.Error(ctx, "CreateJWT:create jwt error",
			log.Err(err),
			log.Any("claims", claims))
		return "", err
	}

	return token.SignedString(key)
}
