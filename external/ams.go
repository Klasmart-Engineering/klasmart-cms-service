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
		stopwatchMap, ok := stopwatches.(map[string]*utils.Stopwatch)
		if ok {
			externalStopwatch := stopwatchMap[string(constant.ExternalStopwatch)]
			if externalStopwatch != nil {
				externalStopwatch.Start()
			}
		}
	}

	fields := []log.Field{
		log.Any("request", req),
		log.Any("response", resp),
	}

	statusCode, err := c.Client.Run(ctx, req, resp)
	if externalStopwatch != nil {
		externalStopwatch.Stop()
		fields = append(fields, log.Duration("duration", externalStopwatch.Duration()))
	}

	log.Debug(ctx, "run query end", fields...)

	return statusCode, err
}
