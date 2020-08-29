package api

import (
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

func (s *Server) approve(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")
	if cid == "" {
		c.JSON(http.StatusBadRequest, "cid can't be empty string")
	}
	err := model.GetReviewerModel().Approve(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err {
	case model.ErrNoContent:
		log.Error(ctx, "approve", log.Any("op", op), log.String("cid", cid), log.Err(err))
		c.JSON(http.StatusNotFound, "content not found")
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		log.Error(ctx, "approve", log.Any("op", op), log.String("cid", cid), log.Err(err))
		// TODO: differentiate error types
		c.JSON(http.StatusInternalServerError, "Internal server error")
	}
}

func (s *Server) reject(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")
	if cid == "" {
		c.JSON(http.StatusBadRequest, "cid can't be empty string")
	}
	// extract reject reason
	err := model.GetReviewerModel().Reject(ctx, dbo.MustGetDB(ctx), cid, "", op)
	switch err {
	case model.ErrNoContent:
		log.Error(ctx, "reject", log.Any("op", op), log.String("cid", cid), log.Err(err))
		c.JSON(http.StatusNotFound, "content not found")
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		log.Error(ctx, "reject", log.Any("op", op), log.String("cid", cid), log.Err(err))
		c.JSON(http.StatusInternalServerError, "Internal server error")
	}
}
