package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type ShortcodeProvider interface {
	Current(ctx context.Context, op *entity.Operator) (int, error)
	Intersect(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, shortcodes []string) (map[string]bool, error)
	IsShortcodeExists(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, ancestor string, shortcode string) (bool, error)
	IsShortcodeCached(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, shortcode string) (bool, error)
	ShortcodeLength() int
}

type ShortcodeModel struct {
	kind     entity.ShortcodeKind
	cursor   int
	provider ShortcodeProvider
}

func (scm *ShortcodeModel) cache(ctx context.Context, op *entity.Operator, cursor int, shortcode string) error {
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

//
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

//
var shortcodeFactory = make(map[entity.ShortcodeKind]*ShortcodeModel)

func GetShortcodeModel(ctx context.Context, op *entity.Operator, kind entity.ShortcodeKind) *ShortcodeModel {
	if _, ok := shortcodeFactory[kind]; !ok {
		scm := &ShortcodeModel{kind: kind}
		switch kind {
		case entity.KindMileStone:
			scm.provider = GetMilestoneModel()
		case entity.KindOutcome:
			scm.provider = GetOutcomeModel()
		default:
			scm.provider = GetOutcomeModel()
		}
		var err error
		scm.cursor, err = scm.provider.Current(ctx, op)
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

func (scm *ShortcodeModel) generate(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, cursor int) (int, string, error) {
	var shortcode string
	if cursor >= constant.ShortcodeSpace {
		return 0, "", constant.ErrOverflow
	}

	shortcodes := make([]string, 0, constant.ShortcodeFindStep)
	for i := 0; i < constant.ShortcodeFindStep; i++ {
		index := cursor + i
		if index >= constant.ShortcodeSpace {
			break
		}
		code, err := utils.NumToBHex(ctx, index, constant.ShortcodeBaseCustom, scm.provider.ShortcodeLength())
		if err != nil {
			log.Debug(ctx, "Generate: NumToBHex failed",
				log.Any("op", op),
				log.Int("cursor", scm.cursor),
				log.Int("index", i))
			return 0, "", err
		}
		shortcodes = append(shortcodes, code)
	}

	intersects, err := scm.provider.Intersect(ctx, op, tx, shortcodes)
	if err != nil {
		log.Debug(ctx, "Generate: Intersect failed",
			log.Any("op", op),
			log.Int("cursor", scm.cursor))
		return 0, "", err
	}

	for i := 0; i < len(shortcodes); i++ {
		shortcode = shortcodes[i]
		if !intersects[shortcode] {
			return cursor + i, shortcode, nil
		}
	}

	log.Info(ctx, "generate:  recursion",
		log.Any("op", op),
		log.Int("cursor", cursor),
		log.Int("length", len(shortcodes)))

	return scm.generate(ctx, op, tx, cursor+len(shortcodes))
}

func (scm *ShortcodeModel) Generate(ctx context.Context, op *entity.Operator) (string, error) {
	var shortcode string
	var cursor int
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
		cursor, shortcode, err = scm.generate(ctx, op, tx, scm.cursor+1)
		if err != nil {
			log.Debug(ctx, "Generate: generate failed",
				log.Any("op", op),
				log.Int("cursor", scm.cursor))
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	err = scm.cache(ctx, op, cursor, shortcode)
	return shortcode, err
}
