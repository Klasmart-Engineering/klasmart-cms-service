package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type ShortcodeModel struct {
}

func (scm ShortcodeModel) Generate(ctx context.Context, tx *dbo.DBContext, kind string, orgID string, shortcode string) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixShortcodeMute, kind, orgID)
	if err != nil {
		log.Error(ctx, "GenerateShortcode: NewLock failed",
			log.Err(err),
			log.String("org", orgID),
			log.String("shortcode", shortcode))
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()
	shortcodes, err := scm.search(ctx, tx, kind, orgID)
	if err != nil {
		log.Error(ctx, "GenerateShortcode: search failed",
			log.Err(err),
			log.String("org", orgID),
			log.String("shortcode", shortcode))
		return "", err
	}
	for i := 0; i < constant.ShortcodeSpace; i++ {
		value, err := utils.NumToBHex(ctx, i, constant.ShortcodeBaseCustom, constant.ShortcodeShowLength)
		if err != nil {
			return "", err
		}
		if _, ok := shortcodes[value]; !ok && value != utils.PaddingString(shortcode, constant.ShortcodeShowLength) {
			err = scm.cacheIt(ctx, kind, orgID, value)
			if err != nil {
				log.Error(ctx, "GenerateShortcode: cacheIt failed",
					log.String("org", orgID),
					log.String("new", value),
					log.String("old", shortcode))
				return "", err
			}
			log.Info(ctx, "GenerateShortcode: cacheIt success",
				log.String("old", shortcode),
				log.String("new", value))
			return value, nil
		}
	}
	return "", constant.ErrExceededLimit
}

func (scm ShortcodeModel) search(ctx context.Context, tx *dbo.DBContext, kind string, orgID string) (map[string]struct{}, error) {
	var table string
	if kind == constant.ShortcodeOutcomeKind {
		table = entity.OutcomeTable
	}
	if kind == constant.ShortcodeMilestoneKind {
		table = entity.MilestoneTable
	}

	dbShortcodes, err := da.GetShortcodeDA().Search(ctx, tx, table, &da.ShortcodeCondition{
		OrgID: sql.NullString{String: orgID, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "search: Search failed",
			log.String("kind", kind),
			log.String("org", orgID))
		return nil, err
	}
	cachedShortcodes, f, err := da.GetShortcodeCacheDA().SearchWithPatten(ctx, kind, orgID)
	if err != nil {
		log.Error(ctx, "search: SearchWithPatten failed",
			log.String("kind", kind),
			log.String("org", orgID))
		return nil, err
	}
	shortcodes := make(map[string]struct{})
	for i := range dbShortcodes {
		shortcodes[dbShortcodes[i].Shortcode] = struct{}{}
	}
	for i := range cachedShortcodes {
		shortcodes[f(cachedShortcodes[i])] = struct{}{}
	}
	return shortcodes, nil
}

func (scm ShortcodeModel) isCached(ctx context.Context, kind string, orgID string, shortcode string) (bool, error) {
	return da.GetShortcodeCacheDA().Exists(ctx, kind, orgID, shortcode)
}

func (scm ShortcodeModel) isOccupied(ctx context.Context, tx *dbo.DBContext, table string, orgID string, ancestor string, shortcode string) (bool, error) {
	shortcodes, err := da.GetShortcodeDA().Search(ctx, tx, table, &da.ShortcodeCondition{
		OrgID:         sql.NullString{String: orgID, Valid: true},
		NotAncestorID: sql.NullString{String: ancestor, Valid: true},
		Shortcode:     sql.NullString{String: shortcode, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "isOccupied: Search failed",
			log.String("kind", table),
			log.String("org", orgID),
			log.String("shortcode", shortcode))
		return false, err
	}
	for i := range shortcodes {
		if shortcode == shortcodes[i].Shortcode {
			return true, nil
		}
	}
	return false, nil
}

func (scm ShortcodeModel) cacheIt(ctx context.Context, kind string, orgID string, shortcode string) error {
	return da.GetShortcodeCacheDA().Save(ctx, kind, orgID, shortcode)
}

func (scm ShortcodeModel) removeIt(ctx context.Context, kind string, orgID string, shortcode string) error {
	return da.GetShortcodeCacheDA().Remove(ctx, kind, orgID, shortcode)
}

var (
	_shortcodeModel     *ShortcodeModel
	_shortcodeModelOnce sync.Once
)

func GetShortcodeModel() *ShortcodeModel {
	_shortcodeModelOnce.Do(func() {
		_shortcodeModel = new(ShortcodeModel)
	})
	return _shortcodeModel
}
