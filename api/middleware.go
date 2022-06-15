package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func ExtractSession(c *gin.Context) (string, error) {
	token, err := c.Cookie("access")
	if token == "" || err != nil {
		log.Info(c.Request.Context(), "ExtractSession", log.String("session", "no session"), log.Err(err))
		return "", constant.ErrUnAuthorized
	}
	return token, nil
}

func (Server) mustAms(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		log.Info(c.Request.Context(), "mustAms", log.String("session", "no authorization"))
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}

	prefix := "Bearer "
	if strings.HasPrefix(token, prefix) {
		token = token[len(prefix):]
	}

	_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return config.Get().AMS.TokenVerifyKey, nil
	})
	if err != nil {
		log.Info(c.Request.Context(), "mustAms", log.String("token", token), log.Err(err))
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}
}

func (Server) mustDataService(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		log.Info(c.Request.Context(), "mustDataService", log.String("session", "no authorization"))
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}

	prefix := "Bearer "
	if strings.HasPrefix(token, prefix) {
		token = token[len(prefix):]
	}

	if token != config.Get().DataService.PublicAuthorizedKey {
		log.Info(c.Request.Context(), "mustDataService", log.String("token", token))
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}
}

const operatorKey = "_op_"

func (Server) mustLogin(c *gin.Context) {
	token, err := ExtractSession(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}

	claims := &struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		*jwt.StandardClaims
	}{}
	_, err = jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return config.Get().AMS.TokenVerifyKey, nil
	})
	if err != nil {
		log.Info(c.Request.Context(), "MustLogin", log.String("token", token), log.Err(err))
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}
	if c.Query(constant.URLOrganizationIDParameter) == "" {
		log.Info(c.Request.Context(), "MustLogin", log.String("OrgID", "no org_id"))
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorizedNoOrgID))
		return
	}
	op := &entity.Operator{
		UserID: claims.ID,
		OrgID:  c.Query(constant.URLOrganizationIDParameter),
		Token:  token,
	}
	c.Set(operatorKey, op)
}

func (Server) mustLoginWithoutOrgID(c *gin.Context) {
	token, err := ExtractSession(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}

	claims := &struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		*jwt.StandardClaims
	}{}
	_, err = jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) {
		return config.Get().AMS.TokenVerifyKey, nil
	})
	if err != nil {
		log.Info(c.Request.Context(), "MustLogin", log.String("token", token), log.Err(err))
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}
	op := &entity.Operator{
		UserID: claims.ID,
		OrgID:  c.Query(constant.URLOrganizationIDParameter),
		Token:  token,
	}
	c.Set(operatorKey, op)
}

func (Server) mustSTM(c *gin.Context) {
	ctx := c.Request.Context()
	token, err := ExtractSession(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}

	var claims jwt.StandardClaims
	_, err = jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (interface{}, error) {
		return config.Get().STMInternal.PublicKey, nil
	})
	if err != nil {
		log.Info(ctx, "grant internal privilege", log.String("token", token), log.Err(err))
		c.AbortWithStatusJSON(http.StatusUnauthorized, L(GeneralUnAuthorized))
		return
	}
	log.Info(ctx, "grant internal privilege", log.Any("claims", claims))
}
func (Server) getOperator(c *gin.Context) *entity.Operator {
	op, exist := c.Get(operatorKey)
	if exist {
		return op.(*entity.Operator)
	}
	return &entity.Operator{}
}

func (s Server) logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requstURL := c.Request.URL.String()

		fields := []log.Field{
			log.String("type", "logger"),
			log.String("method", c.Request.Method),
			log.String("url", requstURL),
			log.String("viewer_ip", c.GetHeader("X-Forwarded-For")),
			log.String("viewer_country", c.GetHeader("Cloudfront-Viewer-Country")),
			log.String("viewer_timezone", c.GetHeader("CloudFront-Viewer-Time-Zone")),
		}

		log.Info(c.Request.Context(), fmt.Sprintf("[START] %s %s", c.Request.Method, requstURL), fields...)

		// Process request
		c.Next()

		// add response fields
		duration := time.Since(start)
		fields = append(fields,
			log.Any("operator", s.getOperator(c)),
			log.Int("size", c.Writer.Size()),
			log.Int("status", c.Writer.Status()),
			log.Int64("duration", duration.Milliseconds()))

		// log stopwatch durations
		stopwatchMap, found := utils.GetStopwatches(c.Request.Context())
		if found {
			durations := make(map[string]int64, len(stopwatchMap))
			for key, stopwatch := range stopwatchMap {
				durations[key] = stopwatch.Duration().Milliseconds()
			}

			fields = append(fields, log.Any("durations", durations))
		}

		fn := log.Info
		if duration > constant.FunctionExpirationLimit {
			fn = log.Warn
		}

		if c.Writer.Status() == http.StatusInternalServerError {
			fn = log.Error
		}

		fn(c.Request.Context(), fmt.Sprintf("[END] %s %s (%d) in %s", c.Request.Method, requstURL, c.Writer.Status(), duration.String()), fields...)
	}
}

// recovery returns a middleware for a given writer that recovers from any panics and writes a 500 if there was one.
func (s Server) recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				log.Error(c.Request.Context(), "[Recovery] panic recovered",
					log.Any("err", err),
					log.Stack("stack"),
					log.Any("operator", s.getOperator(c)),
					log.String("type", "recovery"),
					log.String("method", c.Request.Method),
					log.String("url", c.Request.URL.String()),
					log.Int64("content_length", c.Request.ContentLength),
					log.String("viewer_ip", c.GetHeader("X-Forwarded-For")),
					log.String("viewer_country", c.GetHeader("Cloudfront-Viewer-Country")),
					log.String("viewer_timezone", c.GetHeader("CloudFront-Viewer-Time-Zone")),
				)

				// If the connection is dead, we can't write a status to it.
				if brokenPipe {
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next()
	}
}

func (s Server) contextStopwatch() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.SetupStopwatch(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func (s Server) getNewRelicMiddleware() gin.HandlerFunc {
	nrCfg := &config.Get().NewRelic
	nrApp, err := newrelic.NewApplication(
		newrelic.ConfigAppName(nrCfg.NewRelicAppName),
		newrelic.ConfigLicense(nrCfg.NewRelicLicenseKey),
		newrelic.ConfigDistributedTracerEnabled(nrCfg.NewRelicDistributedTracingEnabled),
		func(c *newrelic.Config) { c.Labels = nrCfg.NewRelicLabels }, // newrelic doesn't supply labels helper func
	)
	if err != nil {
		log.Panic(context.Background(), "failed to init new relic app", log.Any("new_relic_config", nrCfg))
	}
	return nrgin.Middleware(nrApp)
}

func (s Server) newRelicMiddlewareRectifier() gin.HandlerFunc {
	return func(c *gin.Context) {
		txn := newrelic.FromContext(c)
		if txn != nil {
			nrCtx := newrelic.NewContext(c.Request.Context(), txn)
			c.Request = c.Request.WithContext(nrCtx)
			log.Debug(nrCtx, "txn found")
		}
		c.Next()
	}
}
