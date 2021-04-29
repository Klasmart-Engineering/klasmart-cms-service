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
	Kind da.IShortcodeKind
}

type BaseShortcodeModel struct{}

func (scm ShortcodeModel) Generate(ctx context.Context, tx *dbo.DBContext, orgID string, shortcode string) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixShortcodeMute, scm.Kind.Kind(), orgID)
	if err != nil {
		log.Error(ctx, "GenerateShortcode: NewLock failed",
			log.Err(err),
			log.String("org", orgID),
			log.String("shortcode", shortcode))
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()
	shortcodes, err := scm.search(ctx, tx, orgID)
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
		if !shortcodes[value] && value != utils.PaddingString(shortcode, constant.ShortcodeShowLength) {
			err = scm.cacheIt(ctx, orgID, value)
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

func (scm ShortcodeModel) search(ctx context.Context, tx *dbo.DBContext, orgID string) (map[string]bool, error) {

	dbShortcodes, err := da.GetShortcodeDA().Search(ctx, tx, scm.Kind, &da.ShortcodeCondition{
		OrgID: sql.NullString{String: orgID, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "search: Search failed",
			log.String("org", orgID))
		return nil, err
	}
	cachedShortcodes, f, err := da.GetShortcodeCacheDA().SearchWithPatten(ctx, scm.Kind.Kind(), orgID)
	if err != nil {
		log.Error(ctx, "search: SearchWithPatten failed",
			log.String("org", orgID))
		return nil, err
	}
	shortcodes := make(map[string]bool)
	for i := range dbShortcodes {
		shortcodes[dbShortcodes[i].Shortcode] = true
	}
	for i := range cachedShortcodes {
		shortcodes[f(cachedShortcodes[i])] = true
	}
	return shortcodes, nil
}

func (scm ShortcodeModel) isCached(ctx context.Context, orgID string, shortcode string) (bool, error) {
	return da.GetShortcodeCacheDA().Exists(ctx, scm.Kind.Kind(), orgID, shortcode)
}

func (scm ShortcodeModel) isOccupied(ctx context.Context, tx *dbo.DBContext, orgID string, ancestor string, shortcode string) (bool, error) {
	shortcodes, err := da.GetShortcodeDA().Search(ctx, tx, scm.Kind, &da.ShortcodeCondition{
		OrgID:         sql.NullString{String: orgID, Valid: true},
		NotAncestorID: sql.NullString{String: ancestor, Valid: true},
		Shortcode:     sql.NullString{String: shortcode, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "isOccupied: Search failed",
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

func (scm ShortcodeModel) cacheIt(ctx context.Context, orgID string, shortcode string) error {
	return da.GetShortcodeCacheDA().Save(ctx, scm.Kind.Kind(), orgID, shortcode)
}

func (scm ShortcodeModel) removeIt(ctx context.Context, orgID string, shortcode string) error {
	return da.GetShortcodeCacheDA().Remove(ctx, scm.Kind.Kind(), orgID, shortcode)
}

type MilestoneShortcodeModel struct{}

func (MilestoneShortcodeModel) Kind() entity.ShortcodeKind {
	return entity.KindMileStone
}

var (
	_milestoneShortcodeModel     *ShortcodeModel
	_milestoneShortcodeModelOnce sync.Once
)

func GetMilestoneShortcodeModel() *ShortcodeModel {
	_milestoneShortcodeModelOnce.Do(func() {
		_milestoneShortcodeModel = new(ShortcodeModel)
		_milestoneShortcodeModel.Kind = new(MilestoneShortcodeModel)
	})
	return _milestoneShortcodeModel
}

type LearningOutcomeShortcodeModel struct{}

func (LearningOutcomeShortcodeModel) Kind() entity.ShortcodeKind {
	return entity.KindOutcome
}

var (
	_learningOutcomeShortcodeModel     *ShortcodeModel
	_learningOutcomeShortcodeModelOnce sync.Once
)

func GetLearningOutcomeShortcodeModel() *ShortcodeModel {
	_milestoneShortcodeModelOnce.Do(func() {
		_learningOutcomeShortcodeModel = new(ShortcodeModel)
		_learningOutcomeShortcodeModel.Kind = new(LearningOutcomeShortcodeModel)
	})
	return _learningOutcomeShortcodeModel
}
