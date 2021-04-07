package da

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"time"
)

func (o OutcomeSqlDA) shortcodeRedisKeyPrefix(orgID string) string {
	return fmt.Sprintf("%s:%s:*", RedisKeyPrefixShortcode, orgID)
}

func (o OutcomeSqlDA) shortcodeRedisKey(orgID, shortcode string) string {
	return fmt.Sprintf("%s:%s:%s", RedisKeyPrefixShortcode, orgID, shortcode)
}

func (o OutcomeSqlDA) IsShortcodeExistInRedis(ctx context.Context, orgID string, shortcode string) (bool, error) {
	result, err := ro.MustGetRedis(ctx).Exists(o.shortcodeRedisKey(orgID, shortcode)).Result()
	if err != nil {
		log.Error(ctx, "isShortcodeExistInRedis: Exists failed",
			log.Err(err),
			log.String("shortcode", shortcode),
			log.String("org", orgID))
		return false, err
	}
	if result == 0 {
		return false, nil
	}
	return true, nil
}

func (o OutcomeSqlDA) IsShortcodeExistInDBWithOtherAncestor(ctx context.Context, tx *dbo.DBContext, orgID string, ancestorID string, shortcode string) (bool, error) {
	sql := fmt.Sprintf("select distinct shortcode from %s where organization_id = ? and ancestor_id != ? and delete_at=0 and length(shortcode) = %d order by shortcode",
		entity.Outcome{}.TableName(), constant.ShortcodeShowLength)

	var result []*struct {
		Shortcode string `gorm:"column:shortcode"`
	}
	err := tx.Raw(sql, orgID, ancestorID).Scan(&result).Error
	if err != nil {
		log.Error(ctx, "SearchShortcode: exec sql failed",
			log.Err(err),
			log.String("sql", sql))
		return false, err
	}
	for i := range result {
		if shortcode == result[i].Shortcode {
			return true, nil
		}
	}
	return false, nil
}

func (o OutcomeSqlDA) searchRedisShortcode(ctx context.Context, orgID string) ([]string, error) {
	shortcodes, err := ro.MustGetRedis(ctx).Keys(o.shortcodeRedisKeyPrefix(orgID)).Result()
	if err != nil {
		log.Error(ctx, "SearchTempShortcode: Keys failed",
			log.Err(err),
			log.String("org", orgID))
		return nil, err
	}
	return shortcodes, nil
}

func (o OutcomeSqlDA) searchDBShortcode(ctx context.Context, tx *dbo.DBContext, orgID string) (map[string]struct{}, error) {
	sql := fmt.Sprintf("select distinct shortcode from %s where organization_id = ? and delete_at=0 and length(shortcode) = %d order by shortcode",
		entity.Outcome{}.TableName(), constant.ShortcodeShowLength)

	var result []*struct {
		Shortcode string `gorm:"column:shortcode"`
	}
	err := tx.Raw(sql, orgID).Scan(&result).Error
	if err != nil {
		log.Error(ctx, "SearchShortcode: exec sql failed",
			log.Err(err),
			log.String("sql", sql))
		return nil, err
	}
	shortcodes := make(map[string]struct{})
	for i := range result {
		shortcodes[result[i].Shortcode] = struct{}{}
	}
	return shortcodes, nil
}

func (o OutcomeSqlDA) SearchShortcode(ctx context.Context, tx *dbo.DBContext, orgID string) (map[string]struct{}, error) {
	shortcodes, err := o.searchDBShortcode(ctx, tx, orgID)
	if err != nil {
		log.Error(ctx, "SearchShortcode: searchDBShortcode failed",
			log.Err(err),
			log.String("org", orgID))
		return nil, err
	}
	temp, err := o.searchRedisShortcode(ctx, orgID)
	if err != nil {
		log.Error(ctx, "SearchShortcode: searchRedisShortcode failed",
			log.Err(err),
			log.String("org", orgID))
		return nil, err
	}
	log.Info(ctx, "SearchShortcode: searchRedisShortcode success",
		log.Strings("shortcode", temp))
	for i := range temp {
		shortcodes[temp[i]] = struct{}{}
	}
	return shortcodes, nil
}

func (o OutcomeSqlDA) SaveShortcodeInRedis(ctx context.Context, orgID string, shortcode string) error {
	return ro.MustGetRedis(ctx).Set(o.shortcodeRedisKey(orgID, shortcode), shortcode, time.Hour).Err()
}
