package model

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"time"
)

type ShortcodeModel struct {
	kind   entity.ShortcodeKind
	cursor int
	client *redis.Client
}

func (scm *ShortcodeModel) Cursor(ctx context.Context, op *entity.Operator) int {
	return scm.cursor
}

func (scm *ShortcodeModel) Cache(ctx context.Context, op *entity.Operator, cursor int, shortcode string) error {
	err := scm.client.Set(scm.cursorKey(ctx, op), cursor, -1).Err()
	if err != nil {
		log.Error(ctx, "Cache cursor failed",
			log.Err(err),
			log.Any("op", op),
			log.Int("cursor", cursor),
			log.String("shortcode", shortcode))
		return err
	}
	err = scm.client.Set(scm.shortcodeKey(ctx, op, shortcode), shortcode, time.Hour).Err()
	if err != nil {
		log.Error(ctx, "Cache shortcode failed",
			log.Err(err),
			log.Any("op", op),
			log.Int("cursor", cursor),
			log.String("shortcode", shortcode))
		return err
	}
	scm.cursor = cursor
	return nil
}

func (scm *ShortcodeModel) IsCached(ctx context.Context, op *entity.Operator, shortcode string) (bool, error) {
	result, err := scm.client.Exists(scm.shortcodeKey(ctx, op, shortcode)).Result()
	if err != nil {
		log.Error(ctx, "IsCached: redis access failed",
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

func (scm *ShortcodeModel) Remove(ctx context.Context, op *entity.Operator, shortcode string) error {
	err := scm.client.Del(scm.shortcodeKey(ctx, op, shortcode)).Err()
	if err != nil {
		log.Error(ctx, "Remove: redis access failed",
			log.Err(err),
			log.Any("op", op),
			log.String("shortcode", shortcode))
		return err
	}
	return nil
}

func (scm *ShortcodeModel) cursorKey(ctx context.Context, op *entity.Operator) string {
	return fmt.Sprintf("%s:%s:cursor:shortcode", op.OrgID, string(scm.kind))
}

func (scm *ShortcodeModel) shortcodeKey(ctx context.Context, op *entity.Operator, shortcode string) string {
	return fmt.Sprintf("%s:%s:shortcode:%s", op.OrgID, string(scm.kind), shortcode)
}

var shortcodeFactory = make(map[entity.ShortcodeKind]*ShortcodeModel)

func GetShortcodeModel(ctx context.Context, op *entity.Operator, kind entity.ShortcodeKind) *ShortcodeModel {
	if _, ok := shortcodeFactory[kind]; !ok {
		scm := &ShortcodeModel{kind: kind}
		scm.client = ro.MustGetRedis(ctx)
		cursor, err := scm.client.Get(scm.cursorKey(ctx, op)).Int()
		if err != nil {
			if err.Error() != "redis: nil" {
				log.Error(ctx, "GetShortcodeModel: redis access failed",
					log.Err(err),
					log.Any("op", op),
					log.String("kind", string(kind)))
				return nil
			}
			cursor = -1
		}
		scm.cursor = cursor
		shortcodeFactory[kind] = scm
	}
	return shortcodeFactory[kind]
}
