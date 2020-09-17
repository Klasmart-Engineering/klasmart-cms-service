package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

func (s *Server) getScheduleLiveToken(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	scheduleID := c.Param("id")
	token, err := model.GetLiveTokenModel().MakeLiveToken(ctx, op, scheduleID)
	switch err {
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"token": token,
		})
	case constant.ErrRecordNotFound:
		log.Info(ctx, "schedule not found", log.Err(err), log.String("scheduleID", scheduleID))
		c.JSON(http.StatusNotFound, L(Unknown))
	default:
		log.Error(ctx, "make schedule live token error", log.Err(err), log.String("scheduleID", scheduleID))
		c.JSON(http.StatusInternalServerError, L(Unknown))
	}
}

func (s *Server) getContentLiveToken(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	contentID := c.Param("content_id")
	token, err := model.GetLiveTokenModel().MakeLivePreviewToken(ctx, op, contentID)
	switch err {
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"token": token,
		})
	case constant.ErrRecordNotFound:
		log.Info(ctx, "content not found", log.Err(err), log.String("contentID", contentID))
		c.JSON(http.StatusNotFound, L(Unknown))

	default:
		log.Error(ctx, "make content live token error", log.Err(err), log.String("contentID", contentID))
		c.JSON(http.StatusInternalServerError, L(Unknown))
	}
}
