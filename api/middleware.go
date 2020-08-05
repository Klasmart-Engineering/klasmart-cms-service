package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"net/http"
	"strings"
)

func ExtractSession(c *gin.Context) (string, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		err := constant.ErrUnauthorized
		log.Error(c.Request.Context(), "ExtractSession", log.Err(err))
		return "", err
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
		UserID: "No.1",
		Role:   "admin",
	}
	c.Set(Operator, op)
}

func GetOperator(c *gin.Context) (*entity.Operator, bool) {
	op, exist := c.Get(Operator)
	if exist {
		return op.(*entity.Operator), true
	}
	return nil, false
}
