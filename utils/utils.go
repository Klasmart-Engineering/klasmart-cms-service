package utils

import (
	"context"
	"crypto/hmac"
	"encoding/base64"
	"errors"
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
	result := make([]uint8, l)
	index := l
	for i := 0; num != 0 && i < l; {
		yu := num % n
		index = l - 1 - i
		result[index] = num2char[yu]
		num = num / n
		i++
	}
	for i := 0; i < index; i++ {
		result[i] = num2char[0]
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

var indMap = map[byte]int{
	48: 0, // '0'
	49: 1, // '1'
	50: 2, // '0'
	51: 3, // '1'
	52: 4, // '0'
	53: 5, // '1'
	54: 6, // '0'
	55: 7, // '1'
	56: 8, // '0'
	57: 9, // '1'

	65: 10, // 'A'
	66: 11, // 'B'
	67: 12, // 'C'
	68: 13, // 'D'
	69: 14, // 'E'
	70: 15, // 'F'
	71: 16, // 'G'
	72: 17, // 'H'
	73: 18, // 'I'
	74: 19, // 'J'
	75: 20, // 'K'
	76: 21, // 'L'
	77: 22, // 'M'
	78: 23, // 'N'
	79: 24, // 'O'
	80: 25, // 'P'
	81: 26, // 'Q'
	82: 27, // 'R'
	83: 28, // 'S'
	84: 29, // 'T'
	85: 30, // 'U'
	86: 31, // 'V'
	87: 32, // 'W'
	88: 33, // 'X'
	89: 34, // 'Y'
	90: 35, // 'Z'
}

func BHexToNum(ctx context.Context, hexString string) (int, error) {
	length := len(hexString)
	ind := 0
	for i := 0; i < length; i++ {
		power := 1
		for j := 0; j < length-1-i; j++ {
			power *= constant.ShortcodeBaseCustom
		}
		if value, ok := indMap[hexString[i]]; ok {
			ind += value * power
		} else {
			log.Error(ctx, "BHexToNum: failed",
				log.String("hex", hexString),
				log.Int("index", i))
			return 0, errors.New("not exist")
		}
	}
	return ind, nil
}

func PaddingString(s string, l int) string {
	if l <= len(s) {
		return s
	}
	return strings.Repeat("0", l-len(s)) + s
}
