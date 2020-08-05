package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"net/http"
	"strings"
)

func ExtractSession(c *gin.Context) (string, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		logger.WithContext(c.Request.Context()).Warn("ExtractSession: no session")
		return "", errors.New("ErrUnauthorized")
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
	fmt.Println(token)
	op := &entity.Operator{
		UserID: 1,
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
