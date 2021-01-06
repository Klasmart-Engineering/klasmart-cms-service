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
	log.Debug(ctx, "execute query start", log.Any("request", req))

	externalStopwatch, foundStopwatch := utils.GetStopwatch(ctx, constant.ExternalStopwatch)
	if foundStopwatch {
		externalStopwatch.Start()
	}

	fields := []log.Field{
		log.Any("request", req),
		log.Any("response", resp),
	}

	statusCode, err := c.Client.Run(ctx, req, resp)

	if foundStopwatch {
		externalStopwatch.Stop()
		fields = append(fields, log.Duration("duration", externalStopwatch.Duration()))
	}

	log.Debug(ctx, "execute query end", fields...)

	return statusCode, err
}
