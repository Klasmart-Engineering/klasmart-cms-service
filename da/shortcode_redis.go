package da

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"sync"
	"time"
)

type IShortcodeRedis interface {
	Get(ctx context.Context, op *entity.Operator, kind string) (int, error)
	Cache(ctx context.Context, op *entity.Operator, kind string, cursor int, shortcode string) error
	IsCached(ctx context.Context, op *entity.Operator, kind string, shortcode string) (bool, error)
	Remove(ctx context.Context, op *entity.Operator, kind string, shortcode string) error
}

type ShortcodeRedis struct {
	client *redis.Client
}

func (scr *ShortcodeRedis) Get(ctx context.Context, op *entity.Operator, kind string) (int, error) {
	cursor, err := scr.client.Get(scr.cursorKey(ctx, op, kind)).Int()
	if err != nil {
		if err.Error() != "redis: nil" {
			log.Error(ctx, "Get: redis access failed",
				log.Err(err),
				log.Any("op", op),
				log.String("kind", string(kind)))
			return 0, nil
		}
		cursor = -1
	}
	return cursor, nil
}

func (scr *ShortcodeRedis) Cache(ctx context.Context, op *entity.Operator, kind string, cursor int, shortcode string) error {
	err := scr.client.Set(scr.cursorKey(ctx, op, kind), cursor, -1).Err()
	if err != nil {
		log.Error(ctx, "Cache: Set cursor failed",
			log.Err(err),
			log.Any("op", op),
			log.Int("cursor", cursor),
			log.String("shortcode", shortcode))
		return err
	}
	err = scr.client.Set(scr.shortcodeKey(ctx, op, kind, shortcode), shortcode, time.Hour).Err()
	if err != nil {
		log.Error(ctx, "Cache: Set shortcode failed",
			log.Err(err),
			log.Any("op", op),
			log.Int("cursor", cursor),
			log.String("shortcode", shortcode))
		return err
	}
	return nil
}

func (scr *ShortcodeRedis) IsCached(ctx context.Context, op *entity.Operator, kind string, shortcode string) (bool, error) {
	result, err := scr.client.Exists(scr.shortcodeKey(ctx, op, kind, shortcode)).Result()
	if err != nil {
		log.Error(ctx, "IsCached: Exists failed",
			log.Err(err),
			log.Any("op", op),
			log.String("shortcode", shortcode))
		return false, err
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (scr *ShortcodeRedis) Remove(ctx context.Context, op *entity.Operator, kind string, shortcode string) error {
	err := scr.client.Del(scr.shortcodeKey(ctx, op, kind, shortcode)).Err()
	if err != nil {
		log.Error(ctx, "Remove: Del failed",
			log.Err(err),
			log.Any("op", op),
			log.String("shortcode", shortcode))
		return err
	}
	return nil
}
func (scr *ShortcodeRedis) cursorKey(ctx context.Context, op *entity.Operator, kind string) string {
	return fmt.Sprintf("%s:%s:cursor:shortcode", op.OrgID, kind)
}

func (scr *ShortcodeRedis) shortcodeKey(ctx context.Context, op *entity.Operator, kind string, shortcode string) string {
	return fmt.Sprintf("%s:%s:shortcode:%s", op.OrgID, kind, shortcode)
}

var (
	_shortcodeRedis     *ShortcodeRedis
	_shortcodeRedisOnce sync.Once
)

func GetShortcodeRedis(ctx context.Context) IShortcodeRedis {
	_shortcodeRedisOnce.Do(func() {
		_shortcodeRedis = &ShortcodeRedis{
			client: ro.MustGetRedis(ctx),
		}
	})
	return _shortcodeRedis
}
