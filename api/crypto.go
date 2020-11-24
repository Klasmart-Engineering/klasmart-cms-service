package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

func (s Server) h5pSignature(c *gin.Context) {
	operator := s.getOperator(c)
	urlStr := c.Query("url")
	res, err := utils.URLSignature(operator.UserID, urlStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	h5pPath := fmt.Sprintf("%v?badanamuId=%v&timestamp=%016x&randNum=%016x&signature=%v", urlStr, operator.UserID, res.Timestamp, res.RandNum, res.Signature)

	c.JSON(http.StatusOK, gin.H{
		"url": h5pPath,
	})
}
