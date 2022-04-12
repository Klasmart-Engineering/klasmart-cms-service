package external

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"regexp"
	"sync"
)

var (
	_amsConnection     *AmsConnection
	_amsConnectionOnce sync.Once
)

func GetAmsConnection() *AmsConnection {
	_amsConnectionOnce.Do(func() {
		_amsConnection = &AmsConnection{
			Client: NewClient(config.Get().AMS.EndPoint),
			reg:    regexp.MustCompile("access=\\S+"),
		}

	})
	return _amsConnection
}

type AmsConnection struct {
	Client *GraphGLClient
	reg    *regexp.Regexp
}

func (c AmsConnection) Run(ctx context.Context, req *GraphQLRequest, resp interface{}) (int, error) {
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
		// using authorized key
		if amsAuthorizedKey := config.Get().AMS.AuthorizedKey; amsAuthorizedKey != "" {
			req.SetHeader(constant.AMSAuthorizedHeaderKey, fmt.Sprintf("Bearer %s", amsAuthorizedKey))
			log.Debug(ctx, "using ams authorizedKey",
				log.String("ams_authorized_key", amsAuthorizedKey),
				log.String("header constant key", constant.AMSAuthorizedHeaderKey))
			if len(config.Get().AMS.AuthorizedKey) > 5 {
				log.Debug(ctx, "using ams authorizedKey", log.String("header value", config.Get().AMS.AuthorizedKey[0:5]))
			}
		} else {
			log.Warn(ctx, "Found access graphql without cookie and user_service_api_key is empty")
		}
	}
	statusCode, err := c.Client.Run(ctx, req, resp)
	if err != nil {
		log.Error(
			ctx,
			"ams connection run failed",
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
