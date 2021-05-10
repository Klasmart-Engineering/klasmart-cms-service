package model

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/ro"
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
		counts:   constant.BounceMax,
		interval: time.Hour * constant.BounceInterval,
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
		return
	}
	baseSecret, e := NewOTPSecret(ctx)
	if e != nil {
		err = e
		log.Error(ctx, "Launch: NewOTPSecret failed", log.Err(err))
		return
	}
	client, e := ro.GetRedis(ctx)
	if e != nil {
		err = e
		log.Error(ctx, "Launch: GetRedis failed", log.Err(err))
		return
	}
	// generate temp redis key and OTP
	key := utils.GetHashKeyFromPlatformedString(machine.codeKey)
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

func (machine *Bubble) next(ctx context.Context) (err error) {
	locker, e := mutex.NewLock(ctx, da.RedisKeyPrefixVerifyCodeLock, machine.lockKey())
	if e != nil {
		err = e
		log.Error(ctx, "next: NewLock failed", log.Err(err))
		return
	}
	locker.Lock()
	defer locker.Unlock()

	client, e := ro.GetRedis(ctx)
	if e != nil {
		err = e
		log.Error(ctx, "next: GetRedis failed", log.Err(err))
		return
	}
	_, err = client.RPush(machine.codeKey, time.Now().Unix()).Result()
	if err != nil {
		log.Error(ctx, "next: RPush failed", log.Err(err))
		return
	}

	sendTimeUnixList, e := client.LRange(machine.codeKey, 0, int64(machine.counts)).Result()
	if e != nil {
		err = e
		log.Error(ctx, "next: LRange failed", log.Err(err))
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
			log.Error(ctx, "next: ParseInt failed", log.Err(err))
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
		log.Error(ctx, "next: exceed", log.Err(err))
		return err
	}

	// remain [expiredCount,s.maxTimes-1]
	_, err = client.LTrim(machine.codeKey, int64(expiredCount), int64(machine.counts)).Result()
	if err != nil {
		log.Error(ctx, "next: LTrim failed", log.Err(err))
		return
	}

	return
}

func genSalt(ctx context.Context, salt string) string {
	if salt == "" {
		otp, err := NewOTPSecret(ctx)
		if err != nil {
			salt = constant.DefaultSalt
		}
		salt = string(otp)
	}
	return salt
}
