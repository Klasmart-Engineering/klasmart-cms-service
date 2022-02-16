package external

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
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
			reg:    regexp.MustCompile("access=\\S+"),
		}

	})
	return _amsClient
}

type AmsClient struct {
	*chlorine.Client
	reg *regexp.Regexp
}

func (c AmsClient) Run(ctx context.Context, req *chlorine.Request, resp *chlorine.Response) (int, error) {
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
	statusCode, err := c.Client.Run(ctx, req, resp)
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

var (
	amsServices     AmsServices
	amsServicesOnce sync.Once
)

type AmsServices struct {
	ProgramService ProgramServiceProvider
	TeacherService TeacherServiceProvider
	UserService    UserServiceProvider
}

func GetAmsServices() AmsServices {
	amsServicesOnce.Do(func() {
		amsServices = AmsServices{
			ProgramService: GetProgramServiceProvider(),
			TeacherService: GetTeacherServiceProvider(),
			UserService:    GetUserServiceProvider(),
		}
	})

	return amsServices
}
