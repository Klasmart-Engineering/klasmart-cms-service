package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type ShortcodeProvider interface {
	Current(ctx context.Context, op *entity.Operator) (int, error)
	Intersect(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, shortcodes []string) (map[string]bool, error)
	IsShortcodeExists(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, ancestor string, shortcode string) (bool, error)
	IsShortcodeCached(ctx context.Context, op *entity.Operator, shortcode string) (bool, error)
	RemoveShortcode(ctx context.Context, op *entity.Operator, shortcode string) error
	Cache(ctx context.Context, op *entity.Operator, cursor int, shortcode string) error
	ShortcodeLength() int
}

type ShortcodeModel struct {
}

var (
	_shortcodeModel     *ShortcodeModel
	_shortcodeModelOnce sync.Once
)

func GetShortcodeModel() *ShortcodeModel {
	_shortcodeModelOnce.Do(func() {
		_shortcodeModel = &ShortcodeModel{}
	})
	return _shortcodeModel
}

func (scm *ShortcodeModel) generate(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, cursor int, provider ShortcodeProvider) (int, string, error) {
	if cursor >= constant.ShortcodeSpace {
		return 0, "", constant.ErrOverflow
	}

	shortcodes := make([]string, 0, constant.ShortcodeFindStep)
	for index := cursor; index < cursor+constant.ShortcodeFindStep; index++ {
		if index >= constant.ShortcodeSpace {
			break
		}
		code, err := utils.NumToBHex(ctx, index, constant.ShortcodeBaseCustom, provider.ShortcodeLength())
		if err != nil {
			log.Debug(ctx, "Generate: NumToBHex failed",
				log.Any("op", op),
				log.Int("cursor", cursor),
				log.Int("index", index))
			return 0, "", err
		}
		shortcodes = append(shortcodes, code)
	}

	intersects, err := provider.Intersect(ctx, op, tx, shortcodes)
	if err != nil {
		log.Debug(ctx, "Generate: Intersect failed",
			log.Any("op", op),
			log.Int("cursor", cursor))
		return 0, "", err
	}

	for i, shortcode := range shortcodes {
		if !intersects[shortcode] {
			return cursor + i, shortcode, nil
		}
	}

	log.Info(ctx, "generate:  recursion",
		log.Any("op", op),
		log.Int("cursor", cursor),
		log.Int("length", len(shortcodes)))

	return scm.generate(ctx, op, tx, cursor+len(shortcodes), provider)
}

func (scm *ShortcodeModel) Generate(ctx context.Context, op *entity.Operator, provider ShortcodeProvider) (string, error) {
	var shortcode string
	cursor, err := provider.Current(ctx, op)
	if err != nil {
		log.Debug(ctx, "Generate: Current failed",
			log.Any("op", op),
			log.Int("cursor", cursor))
		return "", err
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
		cursor, shortcode, err = scm.generate(ctx, op, tx, cursor+1, provider)
		if err != nil {
			log.Debug(ctx, "Generate: generate failed",
				log.Any("op", op),
				log.Int("cursor", cursor))
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	err = provider.Cache(ctx, op, cursor, shortcode)
	return shortcode, err
}
