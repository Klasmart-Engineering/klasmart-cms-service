package api

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func ExtractSession(c *gin.Context) (string, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		log.Error(c.Request.Context(), "ExtractSession", log.Err(errors.New("no session")))
		// TODO: for mock
		//return "", constant.ErrUnAuthorized
		return "", nil
	}

	prefix := "Bearer "
	if strings.HasPrefix(token, prefix) {
		token = token[len(prefix):]
	}

	return token, nil
}

const Operator = "_op_"

func MustLogin(c *gin.Context) {
	token, err := ExtractSession(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	// TODO: get user info from token
	log.Info(c.Request.Context(), "MustLogin", log.String("token", token))
	op := &entity.Operator{
		UserID: "1",
		OrgID:  "1",
		Role:   "admin",
	}
	c.Set(Operator, op)
}

func GetOperator(c *gin.Context) *entity.Operator {
	op, exist := c.Get(Operator)
	if exist {
		return op.(*entity.Operator)
	}
	return &entity.Operator{}
}

func GetTimeLocation(c *gin.Context) *time.Location {
	tz := c.GetHeader("CloudFront-Viewer-Time-Zone")
	if tz == "" {
		log.Debug(c.Request.Context(), "GetTimeLocation: get header failed")
		return time.Local
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Debug(c.Request.Context(), "GetTimeLocation: load location failed", log.Err(err))
		return time.Local
	}
	return loc
}

func logger() gin.HandlerFunc {
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

		// login info is not exists before c.Next()

		// add response fields
		duration := time.Since(start)
		fields = append(fields,
			log.Any("operator", GetOperator(c)),
			log.String("session", c.GetHeader("Session")),
			log.Int("size", c.Writer.Size()),
			log.Int("status", c.Writer.Status()),
			log.Duration("duration", duration))

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
func recovery() gin.HandlerFunc {
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
					log.Stack("stack"),
					log.Any("operator", GetOperator(c)),
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
