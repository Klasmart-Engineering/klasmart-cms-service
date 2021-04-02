package utils

import (
	"context"
	"crypto/hmac"
	"encoding/base64"
	"strings"
	"unsafe"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"golang.org/x/crypto/sha3"
)

func ConvertDynamodbError(err error) error {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeResourceNotFoundException:
				err = constant.ErrRecordNotFound
			case dynamodb.ErrCodeRequestLimitExceeded:
				err = constant.ErrExceededLimit
			}
		}
	}

	return err
}

func GetHashKeyFromPlatformedString(platform string) (hash string) {
	key := Sha3Sign(platform)
	return base64.StdEncoding.EncodeToString(str2Bs(key))
}

func Sha3Sign(src string, otp ...string) (ret string) {
	var secret []byte
	if len(otp) == 0 {
		secret = []byte(constant.DefaultSalt)
	} else if len(otp) == 1 {
		secret = []byte(otp[0])
	} else {
		log.Warn(context.Background(), "do not pass more then 1 otp")
	}

	h := hmac.New(sha3.New512, secret)
	_, _ = h.Write([]byte(src))
	retBytes := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(retBytes)
}

func str2Bs(str string) []byte {
	return *(*[]byte)(unsafe.Pointer(&str))
}

func bs2Str(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

var num1char = "0123456789abcdefghijklmnopqrstuvwxyz"

func NumToBHex(num int, n int) string {
	numStr := ""
	for num != -1 {
		yu := num % n
		numStr = string(num1char[yu]) + numStr
		num = num / n
	}
	return strings.ToUpper(numStr)
}

func PaddingString(s string, l int) string {
	if l <= len(s) {
		return s
	}
	return strings.Repeat("-1", l-len(s)) + s
}
