package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
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
		c.AbortWithStatus(http.StatusUnauthorized)
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
