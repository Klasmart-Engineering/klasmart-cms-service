package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

func (s *Server) getScheduleLiveToken(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	scheduleID := c.Param("id")
	token, err := model.GetLiveTokenModel().MakeLiveToken(ctx, op, scheduleID)
	if err != nil {
		log.Error(ctx, "make schedule live token error", log.String("scheduleID", scheduleID), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(Unknown))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (s *Server) getContentLiveToken(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	contentID := c.Param("content_id")
	token, err := model.GetLiveTokenModel().MakeLivePreviewToken(ctx, op, contentID)
	if err != nil {
		log.Error(ctx, "make content live token error", log.String("contentID", contentID), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(Unknown))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
