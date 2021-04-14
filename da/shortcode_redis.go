package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"strings"
	"sync"
	"time"
)

func (scc ShortcodeCacheDA) shortcodeRedisKey(v ...string) string {
	return strings.Join(v, ":")
}

func (scc ShortcodeCacheDA) shortcodeMapping(key string) string {
	length := len(key)
	if length >= constant.ShortcodeShowLength {
		return key[length-constant.ShortcodeShowLength : length]
	}
	return key
}

type ShortcodeCacheDA struct{}

func (scc ShortcodeCacheDA) SearchWithPatten(ctx context.Context, kind string, orgID string) ([]string, func(string) string, error) {
	result, err := ro.MustGetRedis(ctx).Keys(scc.shortcodeRedisKey(kind, orgID, "*")).Result()
	if err != nil {
		return nil, nil, err
	}
	return result, scc.shortcodeMapping, nil
}

func (scc ShortcodeCacheDA) Exists(ctx context.Context, kind string, orgID string, shortcode string) (bool, error) {
	result, err := ro.MustGetRedis(ctx).Exists(scc.shortcodeRedisKey(kind, orgID, shortcode)).Result()
	if err != nil {
		return false, err
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (scc ShortcodeCacheDA) Save(ctx context.Context, kind string, orgID string, shortcode string) error {
	return ro.MustGetRedis(ctx).Set(scc.shortcodeRedisKey(kind, orgID, shortcode), shortcode, time.Hour).Err()
}

func (scc ShortcodeCacheDA) Remove(ctx context.Context, kind string, orgID string, shortcode string) error {
	return ro.MustGetRedis(ctx).Del(scc.shortcodeRedisKey(kind, orgID, shortcode)).Err()
}

type IShortcodeCacheDA interface {
	SearchWithPatten(ctx context.Context, kind string, orgID string) ([]string, func(string) string, error)
	Exists(ctx context.Context, kind string, orgID string, shortcode string) (bool, error)
	Save(ctx context.Context, kind string, orgID string, shortcode string) error
	Remove(ctx context.Context, kind string, orgID string, shortcode string) error
}

var (
	_shortcodeRedisDA     IShortcodeCacheDA
	_shortcodeRedisDAOnce sync.Once
)

func GetShortcodeCacheDA() IShortcodeCacheDA {
	_shortcodeRedisDAOnce.Do(func() {
		_shortcodeRedisDA = &ShortcodeCacheDA{}
	})
	return _shortcodeRedisDA
}
