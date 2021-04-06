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

var num2char = "0123456789abcdefghijklmnopqrstuvwxyz"

func NumToBHex(ctx context.Context, num int, n int, l int) (string, error) {
	if n > len(num2char) || l > constant.ShortcodeMaxShowLength {
		log.Error(ctx, "NumToBHex: n is overflow",
			log.Int("num", num),
			log.Int("base", n),
			log.Int("length", l))
		return "", constant.ErrOverflow
	}
	result := [constant.ShortcodeMaxShowLength]uint8{num2char[0]}
	for i := 0; num != 0 && i < l; {
		yu := num % n
		result[l-1-i] = num2char[yu]
		num = num / n
		i++
	}
	if num != 0 {
		log.Error(ctx, "NumToBHex: num is overflow",
			log.Int("num", num),
			log.Int("base", n),
			log.Int("length", l))
		return "", constant.ErrOverflow
	}
	return strings.ToUpper(string(result[:l])), nil
}

func PaddingString(s string, l int) string {
	if l <= len(s) {
		return s
	}
	return strings.Repeat("0", l-len(s)) + s
}
