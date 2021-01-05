package external

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
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
		}

	})
	return _amsClient
}

type AmsClient struct {
	*chlorine.Client
}

func (c AmsClient) Run(ctx context.Context, req *chlorine.Request, resp *chlorine.Response) (int, error) {
	log.Debug(ctx, "run query start", log.Any("request", req))

	var externalStopwatch *utils.Stopwatch
	stopwatches := ctx.Value(constant.ContextStopwatchKey)
	if stopwatches != nil {
		durationMap, ok := stopwatches.(map[string]*utils.Stopwatch)
		if ok {
			externalStopwatch := durationMap[string(constant.ExternalStopwatch)]
			if externalStopwatch != nil {
				externalStopwatch.Start()
			}
		}
	}

	statusCode, err := c.Client.Run(ctx, req, resp)
	if externalStopwatch != nil {
		externalStopwatch.Stop()
	}

	log.Debug(ctx, "run query end",
		log.Duration("duration", externalStopwatch.Duration()),
		log.Any("request", req),
		log.Any("response", resp))

	return statusCode, err
}
