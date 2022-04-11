package gqp

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	_amsProvider     *AmsProvider
	_amsProviderOnce sync.Once
)

func GetAmsProvider() *AmsProvider {
	_amsProviderOnce.Do(func() {
		_amsProvider = &AmsProvider{
			Client: NewClient(config.Get().AMS.EndPoint),
			reg:    regexp.MustCompile("access=\\S+"),
		}

	})
	return _amsProvider
}

type AmsProvider struct {
	Client *GraphGLClient
	reg    *regexp.Regexp
}

func Run[ResType ConnectionResponse](ctx context.Context, c *AmsProvider, req *GraphQLRequest, resp *GraphQLResponse[ResType]) (int, error) {
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
			log.Debug(ctx, "using ams authorizedKey", log.String("header key", constant.AMSAuthorizedHeaderKey))
			if len(config.Get().AMS.AuthorizedKey) > 5 {
				log.Debug(ctx, "using ams authorizedKey", log.String("header value", config.Get().AMS.AuthorizedKey[0:5]))
			}
		} else {
			log.Warn(ctx, "Found access graphql without cookie and user_service_api_key is empty")
		}
	}
	//statusCode, err := c.Client.Run(ctx, req, resp)
	statusCode, err := GraphQLRun(ctx, c.Client, req, resp)
	if err != nil {
		log.Error(
			ctx,
			"external/AmsClient.Run error",
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
