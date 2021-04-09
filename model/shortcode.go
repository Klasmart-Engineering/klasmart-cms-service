package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

func (ocm OutcomeModel) GenerateShortcode(ctx context.Context, tx *dbo.DBContext, orgID string, shortcode string) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixShortcodeMute, orgID)
	if err != nil {
		log.Error(ctx, "GenerateShortcode: NewLock failed",
			log.Err(err),
			log.String("org", orgID),
			log.String("shortcode", shortcode))
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()
	shortcodes, err := da.GetOutcomeDA().SearchShortcode(ctx, tx, orgID)
	if err != nil {
		log.Error(ctx, "GenerateShortcode: SearchShortcode failed",
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
			err = da.GetOutcomeDA().SaveShortcodeInRedis(ctx, orgID, value)
			if err != nil {
				log.Error(ctx, "GenerateShortcode: SaveShortcodeInRedis failed",
					log.String("org", orgID),
					log.String("new", value),
					log.String("old", shortcode))
				return "", err
			}
			log.Info(ctx, "GenerateShortcode: SaveShortcodeInRedis success",
				log.String("old", shortcode),
				log.String("new", value))
			return value, nil
		}
	}
	return "", constant.ErrExceededLimit
}
