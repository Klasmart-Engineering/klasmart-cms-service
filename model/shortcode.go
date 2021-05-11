package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ShortcodeModel struct {
	kind   entity.ShortcodeKind
	cursor int
}

func (scm *ShortcodeModel) Cursor(ctx context.Context, op *entity.Operator) int {
	return scm.cursor
}

func (scm *ShortcodeModel) Cache(ctx context.Context, op *entity.Operator, cursor int, shortcode string) error {
	err := da.GetShortcodeRedis(ctx).Cache(ctx, op, string(scm.kind), cursor, shortcode)
	if err != nil {
		log.Debug(ctx, "Cache: redis access failed",
			log.Any("op", op),
			log.Int("cursor", cursor),
			log.String("shortcode", shortcode))
		return err
	}
	scm.cursor = cursor
	return nil
}

func (scm *ShortcodeModel) IsCached(ctx context.Context, op *entity.Operator, shortcode string) (bool, error) {
	exists, err := da.GetShortcodeRedis(ctx).IsCached(ctx, op, string(scm.kind), shortcode)
	if err != nil {
		log.Debug(ctx, "IsCached: redis access failed",
			log.Any("op", op),
			log.String("shortcode", shortcode))
		return false, err
	}
	return exists, nil
}

func (scm *ShortcodeModel) Remove(ctx context.Context, op *entity.Operator, shortcode string) error {
	err := da.GetShortcodeRedis(ctx).Remove(ctx, op, string(scm.kind), shortcode)
	if err != nil {
		log.Error(ctx, "Remove: redis access failed",
			log.Err(err),
			log.Any("op", op),
			log.String("shortcode", shortcode))
		return err
	}
	return nil
}


var shortcodeFactory = make(map[entity.ShortcodeKind]*ShortcodeModel)

func GetShortcodeModel(ctx context.Context, op *entity.Operator, kind entity.ShortcodeKind) *ShortcodeModel {
	if _, ok := shortcodeFactory[kind]; !ok {
		scm := &ShortcodeModel{kind: kind}
		var err error
		scm.cursor, err = da.GetShortcodeRedis(ctx).Get(ctx, op, string(kind))
		if err != nil {
			log.Debug(ctx, "GetShortcodeModel: redis access failed",
				log.Any("op", op),
				log.String("kind", string(kind)))
			return nil
		}
		shortcodeFactory[kind] = scm
	}
	return shortcodeFactory[kind]
}
