package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

func (s *Server) approve(c *gin.Context) {
	ctx := c.Request.Context()
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
	}
	cid := c.Param("content_id")
	if cid == "" {
		c.JSON(http.StatusBadRequest, "cid can't be empty string")
	}
	err := model.GetReviewerModel().Approve(ctx, dbo.MustGetDB(ctx), cid, op)
	if err != nil {
		// TODO: differentiate error types
		c.JSON(http.StatusInternalServerError, "Internal server error")
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func (s *Server) reject(c *gin.Context) {
	ctx := c.Request.Context()
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
	}
	cid := c.Param("content_id")
	if cid == "" {
		c.JSON(http.StatusBadRequest, "cid can't be empty string")
	}
	// extract reject reason
	err := model.GetReviewerModel().Reject(ctx, dbo.MustGetDB(ctx), cid, "", op)
	if err != nil {
		// TODO: differentiate error types
		c.JSON(http.StatusInternalServerError, "Internal server error")
		return
	}
	c.JSON(http.StatusOK, "ok")
}
