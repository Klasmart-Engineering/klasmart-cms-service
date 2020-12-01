package model

import (
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/hgfischer/go-otp"
)

// OTPSecret is the base secret bytes of one time password
// This base secret bytes combines with other strings(say, "group")
// to generate actual one time password. This design makes
// database just store only one base secret with multiple group string,
// instead of storing multiple secret.
type OTPSecret string

// defaultPeriod and defaultWindow should only change for testing.
var defaultPeriod uint8 = 120
var defaultWindow uint8 = 5
var totpPool = sync.Pool{
	New: func() interface{} {
		return &otp.TOTP{
			Period:        defaultPeriod,
			WindowBack:    defaultWindow,
			WindowForward: defaultWindow,
		}
	},
}

func init() {
	periodStr := os.Getenv("OTP_PERIOD")
	if periodStr == "" {
		return
	}
	period, err := strconv.Atoi(periodStr)
	if err != nil || period < 0 || period > 255 {
		log.Warn(context.Background(), "", log.Int("period", period), log.Err(err))
		return
	}
	defaultPeriod = uint8(period)
}

func (s OTPSecret) getTOTPFromPool(group string) *otp.TOTP {
	fullSecret := s.combineGroup(group)
	ret := totpPool.Get().(*otp.TOTP)
	ret.Secret = string(fullSecret)
	ret.Time = time.Now()
	return ret
}

func putTOTPBackToPool(totp *otp.TOTP) {
	totp.Secret = ""
	totpPool.Put(totp)
}

func NewOTPSecret(ctx context.Context) (s OTPSecret, err error) {
	rdm := [24]byte{}
	bs := rdm[:]
	if _, err = rand.Read(bs); err != nil {
		log.Error(ctx, "NewOPTSecret: rand failed", log.Err(err))
		return
	}
	s = OTPSecret(base64.StdEncoding.EncodeToString(bs))
	return
}

func (s OTPSecret) GetCode(ctx context.Context, group string) (code string) {
	totp := s.getTOTPFromPool(group)
	code = totp.Get()
	putTOTPBackToPool(totp)
	log.Info(ctx, "GetCode: ", log.String("group", group), log.String("code", code))
	return
}

func (s OTPSecret) combineGroup(group string) (fullSecret OTPSecret) {
	ret := strings.Join([]string{string(s), group}, "|")
	return OTPSecret(ret)
}

func (s OTPSecret) VerifyCode(ctx context.Context, group string, code string) bool {
	totp := s.getTOTPFromPool(group)
	passed := totp.Verify(code)
	// if !passed {
	// 	utils.SetFailureInfo(ctx, utils.FailureInfo{
	// 		StatusCode: 499,
	// 		Message:    "wrong verification code",
	// 		Error:      fmt.Errorf("wrong verification code"),
	// 	})
	// }
	return passed
}
