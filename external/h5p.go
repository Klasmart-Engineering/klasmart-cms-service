package external

import (
	"context"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"

	"github.com/KL-Engineering/chlorine"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
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
	if err != nil {
		log.Error(
			ctx,
			"external/H5PClient.Run error",
			log.Err(err),
			log.Any("req", req),
			log.Any("statusCode", statusCode),
			log.Any("resp", resp),
		)
		err = &entity.ExternalError{
			Err:  err,
			Type: constant.InternalErrorTypeAssessmentApi,
		}
	}

	if foundStopwatch {
		externalStopwatch.Stop()
	}

	return statusCode, err
}
