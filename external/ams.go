package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"regexp"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	_amsClient     *AmsClient
	_amsClientOnce sync.Once
)

func GetAmsClient() *AmsClient {
	_amsClientOnce.Do(func() {
		_amsClient = &AmsClient{
			Client: chlorine.NewClient(config.Get().AMS.EndPoint),
			reg:    regexp.MustCompile("access=\\S+"),
		}

	})
	return _amsClient
}

type AmsClient struct {
	*chlorine.Client
	reg *regexp.Regexp
}

func (c AmsClient) Run(ctx context.Context, req *chlorine.Request, resp *chlorine.Response) (int, error) {
	externalStopwatch, foundStopwatch := utils.GetStopwatch(ctx, constant.ExternalStopwatch)
	if foundStopwatch {
		externalStopwatch.Start()
	}

	cookie := req.Header.Get(constant.CookieKey)
	if !c.reg.MatchString(cookie) {
		log.Warn(ctx,
			"Found access graphql without cookie",
			log.String(constant.CookieKey, cookie),
			log.Any("req", req))
	}
	statusCode, err := c.Client.Run(ctx, req, resp)

	if foundStopwatch {
		externalStopwatch.Stop()
	}

	return statusCode, err
}
