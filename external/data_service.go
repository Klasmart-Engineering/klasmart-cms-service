package external

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	_dataServiceClient     *DataServiceClient
	_dataServiceClientOnce sync.Once
)

func GetDataServiceClient() *DataServiceClient {
	_dataServiceClientOnce.Do(func() {
		_dataServiceClient = &DataServiceClient{
			Client: chlorine.NewClient(config.Get().DataService.EndPoint),
		}

	})
	return _dataServiceClient
}

type DataServiceClient struct {
	*chlorine.Client
}

func (c DataServiceClient) Run(ctx context.Context, req *chlorine.Request, resp *chlorine.Response) (int, error) {
	externalStopwatch, foundStopwatch := utils.GetStopwatch(ctx, constant.ExternalStopwatch)
	if foundStopwatch {
		externalStopwatch.Start()
	}

	statusCode, err := c.Client.Run(ctx, req, resp)
	if err != nil {
		log.Error(
			ctx,
			"external/DataServiceClient.Run error",
			log.Err(err),
			log.Any("req", req),
			log.Any("statusCode", statusCode),
			log.Any("resp", resp),
		)
		err = &entity.ExternalError{
			Err:  err,
			Type: constant.InternalErrorTypeAms,
		}
	}

	if foundStopwatch {
		externalStopwatch.Stop()
	}

	return statusCode, err
}

var (
	dataServices     DataServices
	dataServicesOnce sync.Once
)

type DataServices struct {
	ScheduleReviewService ScheduleReviewServiceProvider
}

func GetDataServices() DataServices {
	dataServicesOnce.Do(func() {
		dataServices = DataServices{
			ScheduleReviewService: GetScheduleReviewServiceProvider(),
		}
	})

	return dataServices
}
