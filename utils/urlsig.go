package utils

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"math/rand"
	rand2 "crypto/rand"
	"time"
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
func readPrivateKeyDerBase64() (*rsa.PrivateKey, error){
	privateKeyDerBase64 := []byte(config.Get().CryptoConfig.PrivateKey)

	privateKeyDer, err := base64.StdEncoding.DecodeString(string(privateKeyDerBase64))
	if err != nil{
		return nil, err
	}
	return x509.ParsePKCS1PrivateKey(privateKeyDer)
}

func URLSignature(id string, url string)(*SignatureResult, error){
	privateKey, err := readPrivateKeyDerBase64()
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