package utils

import (
	"context"
	"github.com/golang-jwt/jwt"
)

func CreateJWT(ctx context.Context, claims jwt.Claims, privateKey interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	return token.SignedString(privateKey)
}
