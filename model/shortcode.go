package model

import (
	"context"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type ShortcodeProvider interface {
	Current(ctx context.Context, op *entity.Operator) (int, error)
	Intersect(ctx context.Context, tx *dbo.DBContext, orgID string, shortcodeNum int) (map[string]bool, error)
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
			log.Error(ctx, "utils.NumToBHex error",
				log.Err(err),
				log.Int("index", index))
			return 0, "", err
		}

		shortcodes = append(shortcodes, code)
	}

	intersects, err := provider.Intersect(ctx, tx, op.OrgID, cursor)
	if err != nil {
		log.Error(ctx, "provider.Intersect error",
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
