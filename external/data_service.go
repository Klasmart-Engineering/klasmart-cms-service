package external

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/KL-Engineering/common-log/log"

	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

var (
	_dataServiceClient     *DataServiceClient
	_dataServiceClientOnce sync.Once
)

func GetDataServiceClient() *DataServiceClient {
	_dataServiceClientOnce.Do(func() {
		_dataServiceClient = &DataServiceClient{
			Client: &http.Client{},
		}

	})
	return _dataServiceClient
}

type DataServiceClient struct {
	*http.Client
}

func (c DataServiceClient) Run(ctx context.Context, path string, req, resp interface{}) (int, error) {
	externalStopwatch, foundStopwatch := utils.GetStopwatch(ctx, constant.ExternalStopwatch)
	if foundStopwatch {
		externalStopwatch.Start()
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, constant.DataServiceHttpTimeout)
	defer cancel()

	reqBuffer, err := json.Marshal(req)
	if err != nil {
		log.Error(ctxWithTimeout, "json.Marshal error",
			log.Err(err),
			log.Any("req", req))
		return 0, err
	}

	url := config.Get().DataService.EndPoint + path
	request, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodPost, url, bytes.NewBuffer(reqBuffer))
	if err != nil {
		log.Error(ctxWithTimeout, "http.NewRequestWithContext error",
			log.Err(err),
			log.Any("req", req))
		return 0, err
	}

	request.Header.Set(constant.DataServiceAuthorizedHeaderKey, config.Get().DataService.AuthorizedKey)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Accept", "application; charset=utf-8")

	var result *http.Response
	var resultErr error
	httpDone := make(chan struct{})

	start := time.Now()

	go func() {
		result, resultErr = c.Client.Do(request)
		httpDone <- struct{}{}
	}()

	select {
	case <-httpDone:
		log.Info(ctxWithTimeout, "OK")
	case <-ctxWithTimeout.Done():
		duration := time.Since(start)
		log.Error(ctxWithTimeout, "do http failed",
			log.Duration("duration", duration),
			log.Err(resultErr),
			log.String("url", url),
			log.Any("req", req))
		return http.StatusRequestTimeout, ctx.Err()
	}

	duration := time.Since(start)
	if resultErr != nil {
		log.Error(ctxWithTimeout, "do http error",
			log.Duration("duration", duration),
			log.Err(resultErr),
			log.String("url", url),
			log.Any("req", req))
		return 0, resultErr
	}

	defer result.Body.Close()
	response, err := ioutil.ReadAll(result.Body)
	if err != nil {
		log.Error(ctxWithTimeout, "read response error",
			log.Err(err),
			log.Duration("duration", duration),
			log.String("url", url),
			log.Any("req", req),
			log.String("response", string(response)))
		return result.StatusCode, err
	}

	if result.StatusCode != http.StatusOK {
		log.Error(ctxWithTimeout, "data service failed",
			log.Duration("duration", duration),
			log.String("url", url),
			log.Any("req", req),
			log.String("response", string(response)))
		return result.StatusCode, constant.ErrDataServiceFailed
	}

	err = json.Unmarshal(response, resp)
	if err != nil {
		log.Error(ctxWithTimeout, "json.Unmarshal error",
			log.Err(err),
			log.Duration("duration", duration),
			log.String("url", url),
			log.Any("req", req),
			log.String("response", string(response)))
		return result.StatusCode, err
	}
	log.Debug(ctxWithTimeout, "do http success",
		log.Duration("duration", duration),
		log.Any("req", req),
		log.String("response", string(response)))

	if foundStopwatch {
		externalStopwatch.Stop()
	}

	return result.StatusCode, nil
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
