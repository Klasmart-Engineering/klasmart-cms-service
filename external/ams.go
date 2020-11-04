package external

import (
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

var (
	_chlorineClient *chlorine.Client
	_chlorineOnce   sync.Once
)

func GetChlorine() *chlorine.Client {
	_chlorineOnce.Do(func() {
		_chlorineClient = chlorine.NewClient(config.Get().AMS.EndPoint)
	})
	return _chlorineClient
}
