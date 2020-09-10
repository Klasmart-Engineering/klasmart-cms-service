package utils

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"sync"
)

var jwtOnce sync.Once
var (
	prvKey []byte
	pubKey []byte
)

func InitJwtKey() {
	//c := config.Get()
	prvKey = []byte("c.Auth.PrivateKey")
	pubKey = []byte("c.Auth.PublicKey")
}

func CreateJWT(claims jwt.Claims) (signedToken string, err error) {
	jwtOnce.Do(InitJwtKey)
	defer func() {
		if err != nil {
			err = errors.New("failed to create jwt")
		}
	}()

	if err != nil {
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	key, err := jwt.ParseRSAPrivateKeyFromPEM(prvKey)
	if err != nil {
		return
	}

	return token.SignedString(key)
}
