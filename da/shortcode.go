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

var _shortcodeDA *ShortcodeSqlDA

func GetShortcodeDA() IShortcodeDA {
	_shortcodeOnce.Do(func() {
		_shortcodeDA = new(ShortcodeSqlDA)
	})
	return _shortcodeDA
}
