package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
)

type IShortcodeDA interface {
	Search(ctx context.Context, tx *dbo.DBContext, table string, condition *ShortcodeCondition) ([]*entity.ShortcodeElement, error)
}

var _shortcodeOnce sync.Once

var _shortcodeDA *ShortcodeMySqlDA

func GetShortcodeDA() IShortcodeDA {
	_shortcodeOnce.Do(func() {
		_shortcodeDA = new(ShortcodeMySqlDA)
	})
	return _shortcodeDA
}
