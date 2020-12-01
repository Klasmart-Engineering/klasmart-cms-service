package model

import (
	"context"
	"crypto/hmac"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"golang.org/x/crypto/sha3"
)

type Bubble struct {
	codeKey  string
	counts   int
	interval time.Duration
	canceled bool
}

type OptFunc func(*Bubble)

func GetBubbleMachine(codeKey string, opts ...OptFunc) *Bubble {
	machine := &Bubble{
		codeKey:  codeKey,
		counts:   5,
		interval: time.Hour * 2,
	}

	for i := range opts {
		opts[i](machine)
	}

	return machine
}

func OptCounts(count int) OptFunc {
	return func(machine *Bubble) {
		machine.counts = count
	}
}

func OptInterval(interval time.Duration) OptFunc {
	return func(machine *Bubble) {
		machine.interval = interval
	}
}

func (machine *Bubble) Launch(ctx context.Context) (bubble string, err error) {
	err = machine.next(ctx)
	if err != nil {
		return "", err
	}
	baseSecret, err := NewOTPSecret(ctx)
	if err != nil {
		log.Error(ctx, "Launch: NewOTPSecret failed", log.Err(err))
		return
	}
	client, err := ro.GetRedis(ctx)
	if err != nil {
		log.Error(ctx, "Launch: GetRedis failed", log.Err(err))
		return
	}
	// generate temp redis key and OTP
	key := getHashKeyFromPlatformedString(machine.codeKey)
	if err = client.Set(key, string(baseSecret), time.Second*time.Duration(120*100)).Err(); err != nil {
		log.Error(ctx, "Launch: Set failed", log.String("code_hash_key", key), log.Err(err))
		return
	}

	bubble = baseSecret.GetCode(ctx, machine.codeKey)
	return
}

func (machine *Bubble) lockKey() string {
	return machine.codeKey + ":lock"
}

func (machine *Bubble) lock(ctx context.Context, expiration time.Duration) (err error) {
	client, err := ro.GetRedis(ctx)
	if err != nil {
		log.Error(ctx, "lock: GetRedis failed", log.Err(err))
		return err
	}
	if expiration < 0 {
		exTm := -expiration
		if exTm < time.Second*10 {
			err = fmt.Errorf("lock time should >=0 or < -time.Second*10  ")
			return
		}
		ok, e := client.SetNX(machine.lockKey(), 1, exTm).Result()
		if e != nil || !ok {
			err = fmt.Errorf("lock failed key=%v err=%v ok=%v ", machine.lockKey(), e, ok)
			return
		}

		go func(ctx context.Context) {
			for {
				if machine.canceled {
					return
				}

				time.Sleep(exTm - time.Second*2)
				_, _ = client.SetXX(machine.codeKey, 1, exTm).Result()
			}
		}(ctx)
		return
	} else {
		ok, e := client.SetNX(machine.lockKey(), 1, expiration).Result()
		if e != nil || !ok {
			err = fmt.Errorf("lock failed, key=%v err=%v ok=%v ", machine.lockKey(), e, ok)
			log.Error(ctx, "Lock: SetNX failed", log.Err(err))
			return
		}
		return
	}
	return
}

func (machine *Bubble) unlock(ctx context.Context) error {
	machine.canceled = true
	client, err := ro.GetRedis(ctx)
	if err != nil {
		log.Error(ctx, "unlock: GetRedis failed", log.Err(err))
		return err
	}
	_, err = client.Del(machine.lockKey()).Result()
	if err != nil {
		log.Error(ctx, "UnLock: Del failed", log.Err(err))
		return err
	}
	return nil
}

func (machine *Bubble) next(ctx context.Context) (err error) {
	defer func() {
		log.Info(ctx, "Next", log.Err(err))
	}()

	err = machine.lock(ctx, -time.Second*10)
	if err != nil {
		return
	}
	defer machine.unlock(ctx)

	client, err := ro.GetRedis(ctx)
	if err != nil {
		log.Error(ctx, "next: GetRedis failed", log.Err(err))
		return err
	}
	_, err = client.RPush(machine.codeKey, time.Now().Unix()).Result()
	if err != nil {
		return
	}

	sendTimeUnixList, err := client.LRange(machine.codeKey, 0, int64(machine.counts)).Result()
	if err != nil {
		return
	}

	if len(sendTimeUnixList) <= machine.counts {
		return
	}

	expiredCount := 0
	for _, tmUnixStr := range sendTimeUnixList {
		tmUnix, e := strconv.ParseInt(tmUnixStr, 10, 64)
		if e != nil {
			err = fmt.Errorf("parse sendTime err %v ", e)
			return
		}

		isExpired := tmUnix <= time.Now().Add(-machine.interval).Unix()
		if isExpired {
			expiredCount++
		} else {
			break
		}
	}
	if expiredCount == 0 {
		err = constant.ErrExceededLimit
		return
	}

	// remain [expiredCount,s.maxTimes-1]
	_, err = client.LTrim(machine.codeKey, int64(expiredCount), int64(machine.counts)).Result()
	if err != nil {
		return
	}

	return
}

const defaultSalt = "Kidsloop2@GetHashK3y"

func getHashKeyFromPlatformedString(platform string) (hash string) {
	key := sha3Sign(platform)
	return base64.StdEncoding.EncodeToString(str2Bs(key))
}
func sha3Sign(src string, otp ...string) (ret string) {
	var secret []byte
	if len(otp) == 0 {
		secret = []byte(defaultSalt)
	} else if len(otp) == 1 {
		secret = []byte(otp[0])
	} else {
		logger.Warnf("do not pass more then 1 otp")
	}

	h := hmac.New(sha3.New512, secret)
	_, _ = h.Write([]byte(src))
	retBytes := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(retBytes)
}

func genSalt(ctx context.Context, salt string) string {
	if salt == "" {
		otp, err := NewOTPSecret(ctx)
		if err != nil {
			salt = defaultSalt
		}
		salt = string(otp)
	}
	return salt
}

func str2Bs(str string) []byte {
	return *(*[]byte)(unsafe.Pointer(&str))
}

func bs2Str(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

func VerifyCode(ctx context.Context, codeKey string, code string) (bool, error) {
	client, err := ro.GetRedis(ctx)
	if err != nil {
		log.Error(ctx, "VerifyCode: GetRedis failed", log.String("code_key", codeKey), log.Err(err))
		return false, err
	}
	key := getHashKeyFromPlatformedString(codeKey)
	otpSecret, err := client.Get(key).Result()
	if err != nil {
		log.Error(ctx, "VerifyCode: Get failed", log.String("otp_secret", otpSecret), log.Err(err))
		return false, err
	}

	var authPassed bool
	baseSecret := OTPSecret(otpSecret)
	defer func() {
		if authPassed {
			log.Info(ctx, "VerifyCode: defer", log.Err(client.Del(key).Err()))
		}
	}()
	totp := baseSecret.getTOTPFromPool(codeKey)
	authPassed = totp.Verify(code)
	return authPassed, nil
}

func VerifySecretWithSalt(ctx context.Context, password string, secret string, salt string) bool {
	pwdHash := sha3Sign(strings.Replace(password, " ", "", -1), salt)
	if pwdHash == secret {
		return true
	}
	return false
}

func MakeSecretAndSalt(ctx context.Context, password string) (string, string) {
	salt := genSalt(ctx, "")
	secret := sha3Sign(strings.Replace(password, " ", "", -1), salt)
	return salt, secret
}
