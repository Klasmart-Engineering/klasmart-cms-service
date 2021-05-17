package external

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	_h5pClient     *H5PClient
	_h5pClientOnce sync.Once
)

func GetH5PClient() *H5PClient {
	_h5pClientOnce.Do(func() {
		_h5pClient = &H5PClient{
			Client: chlorine.NewClient(config.Get().H5P.EndPoint),
		}

	})
	return _h5pClient
}

type H5PClient struct {
	*chlorine.Client
}

func (c H5PClient) Run(ctx context.Context, req *chlorine.Request, resp *chlorine.Response) (int, error) {
	externalStopwatch, foundStopwatch := utils.GetStopwatch(ctx, constant.ExternalStopwatch)
	if foundStopwatch {
		externalStopwatch.Start()
	}

	statusCode, err := c.Client.Run(ctx, req, resp)

	if foundStopwatch {
		externalStopwatch.Stop()
	}

	return statusCode, err
}
