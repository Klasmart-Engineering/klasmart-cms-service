package external

import (
	"context"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
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

	start := time.Now()
	statusCode, err := c.Client.Run(ctx, req, resp)
	duration := time.Since(start)

	log.Debug(ctx, "run query end",
		log.Duration("duration", duration),
		log.Any("request", req),
		log.Any("response", resp))

	durations := ctx.Value(constant.ContextDurationsKey)
	if durations != nil {
		durationMap, ok := durations.(map[string]int64)
		if ok {
			externalDuration := durationMap[string(constant.ExternalDuration)]
			externalDuration += duration.Milliseconds()
			durationMap[string(constant.ExternalDuration)] = externalDuration
			log.Debug(ctx, "set external duration success", log.Int64("externalDuration", externalDuration))
		}
	}

	return statusCode, err
}
