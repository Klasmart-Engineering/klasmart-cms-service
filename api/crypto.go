package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
)

func (s Server) h5pSignature(c *gin.Context) {
	operator := GetOperator(c)
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

