package utils

import (
	"context"
	"crypto"
	rand2 "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"io/ioutil"
	"math/rand"
	"time"
)

var (
	ErrInvalidPrivateKeyFile = errors.New("invalid private key file")
)
type SignatureResult struct {
	Signature string `json:"signature"`
	RandNum int64 `json:"rand_num"`
	Timestamp int64 `json:"timestamp"`
}
func buildMessage(url string, badanamuId string, timestamp, randNum int64) string {
	message := fmt.Sprintf("%v?badanamuId=%v&timestamp=%016x&randNum=%016x", url, badanamuId, timestamp, randNum)
	return message
}
func SHA256Hash(msg string) []byte {
	hash := sha256.New()
	hash.Write([]byte(msg))
	msgHash := hash.Sum(nil)
	return msgHash
}
//func readPrivateKeyDerBase64() (*rsa.PrivateKey, error){
//	privateKeyDerBase64 := []byte(config.Get().CryptoConfig.PrivateKey)
//
//	privateKeyDer, err := base64.StdEncoding.DecodeString(string(privateKeyDerBase64))
//	if err != nil{
//		return nil, err
//	}
//	return x509.ParsePKCS1PrivateKey(privateKeyDer)
//}

func readPrivateKeyPEM() (*rsa.PrivateKey, error){
	privateKeyPEM, err := ioutil.ReadFile(config.Get().CryptoConfig.PrivateKeyPath)
	if err != nil{
		return nil, err
	}
	block, _ := pem.Decode(privateKeyPEM)
	if block.Type != "RSA PRIVATE KEY" {
		return nil, ErrInvalidPrivateKeyFile
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}


func URLSignature(id string, url string)(*SignatureResult, error){
	privateKey, err := readPrivateKeyPEM()
	if err != nil{
		return nil, err
	}

	now := time.Now().Unix()
	randNum := rand.Int63()
	msg := buildMessage(url, id, now, randNum)

	msgHash := SHA256Hash(msg)
	sig, err := rsa.SignPKCS1v15(rand2.Reader, privateKey, crypto.SHA256, msgHash)
	if err != nil{
		return nil, err
	}

	return &SignatureResult{
		Signature: hex.EncodeToString(sig),
		RandNum:   randNum,
		Timestamp: now,
	}, nil
}

type H5pClaims struct {
	jwt.StandardClaims
	ContentId string `json:"contentId"`
}

func GenerateH5pJWT(ctx context.Context, sub, contentID string) (string ,error){
	stdClaims := getStdClaims("H5P", time.Hour*2, contentID, sub)
	claims := &H5pClaims{
		StandardClaims: stdClaims,
		ContentId: contentID,
	}

	// test not expired token
	signedToken, err := createJWT(claims)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
func createJWT(claims jwt.Claims) (signedToken string, err error) {
	privateKey, err := readPrivateKeyPEM()
	if err != nil {
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	if err != nil {
		return
	}

	return token.SignedString(privateKey)
}


func getStdClaims(aud string, expire time.Duration, id, subject string) jwt.StandardClaims {
	now := time.Now()
	expiresAt := now.Add(expire)
	// TODO: audience should search from database on ourAppID

	return jwt.StandardClaims{
		Audience:  aud,
		ExpiresAt: expiresAt.Unix(),
		Id:        id,
		IssuedAt:  now.Unix(),
		Issuer:    "kl2-h5p",
		NotBefore: 0,
		Subject:   subject,
	}
}
