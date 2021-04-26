package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"strings"
	"sync"
	"time"
)

func (scc ShortcodeRedisDA) shortcodeRedisKey(v ...string) string {
	return strings.Join(v, ":")
}

func (scc ShortcodeRedisDA) shortcodeMapping(key string) string {
	length := len(key)
	if length >= constant.ShortcodeShowLength {
		return key[length-constant.ShortcodeShowLength : length]
	}
	return key
}

type ShortcodeRedisDA struct{}

func (scc ShortcodeRedisDA) SearchWithPatten(ctx context.Context, kind entity.ShortcodeKind, orgID string) ([]string, func(string) string, error) {
	result, err := ro.MustGetRedis(ctx).Keys(scc.shortcodeRedisKey(RedisKeyPrefixShortcode, string(kind), orgID, "*")).Result()
	if err != nil {
		return nil, nil, err
	}
	return result, scc.shortcodeMapping, nil
}

func (scc ShortcodeRedisDA) Exists(ctx context.Context, kind entity.ShortcodeKind, orgID string, shortcode string) (bool, error) {
	result, err := ro.MustGetRedis(ctx).Exists(scc.shortcodeRedisKey(RedisKeyPrefixShortcode, string(kind), orgID, shortcode)).Result()
	if err != nil {
		return false, err
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (scc ShortcodeRedisDA) Save(ctx context.Context, kind entity.ShortcodeKind, orgID string, shortcode string) error {
	return ro.MustGetRedis(ctx).Set(scc.shortcodeRedisKey(RedisKeyPrefixShortcode, string(kind), orgID, shortcode), shortcode, time.Hour).Err()
}

func (scc ShortcodeRedisDA) Remove(ctx context.Context, kind entity.ShortcodeKind, orgID string, shortcode string) error {
	return ro.MustGetRedis(ctx).Del(scc.shortcodeRedisKey(RedisKeyPrefixShortcode, string(kind), orgID, shortcode)).Err()
}

type IShortcodeRedisDA interface {
	SearchWithPatten(ctx context.Context, kind entity.ShortcodeKind, orgID string) ([]string, func(string) string, error)
	Exists(ctx context.Context, kind entity.ShortcodeKind, orgID string, shortcode string) (bool, error)
	Save(ctx context.Context, kind entity.ShortcodeKind, orgID string, shortcode string) error
	Remove(ctx context.Context, kind entity.ShortcodeKind, orgID string, shortcode string) error
}

var (
	_shortcodeRedisDA     IShortcodeRedisDA
	_shortcodeRedisDAOnce sync.Once
)

func GetShortcodeCacheDA() IShortcodeRedisDA {
	_shortcodeRedisDAOnce.Do(func() {
		_shortcodeRedisDA = &ShortcodeRedisDA{}
	})
	return _shortcodeRedisDA
}
