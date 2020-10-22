package api

import (
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @ID approveContentReview
// @Summary approve content
// @Tags content
// @Description approve content by id
// @Accept json
// @Produce json
// @Param content_id path string true "content id"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /contents/{content_id}/review/approve [put]
func (s *Server) approve(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")
	if cid == "" {
		c.JSON(http.StatusBadRequest, "cid can't be empty string")
		return
	}
	err := model.GetReviewerModel().Approve(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err {
	case model.ErrNoContent:
		log.Error(ctx, "approve", log.Any("op", op), log.String("cid", cid), log.Err(err))
		c.JSON(http.StatusNotFound, "content not found")
		return
	case nil:
		c.JSON(http.StatusOK, "ok")
		return
	default:
		log.Error(ctx, "approve", log.Any("op", op), log.String("cid", cid), log.Err(err))
		// TODO: differentiate error types
		c.JSON(http.StatusInternalServerError, "Internal server error")
		return
	}
}

type RejectReasonRequest struct {
	Reasons []string `json:"reject_reason"`
	Remark  string   `json:"remark"`
}

// @ID rejectContentReview
// @Summary reject content
// @Tags content
// @Description reject content by id
// @Accept json
// @Produce json
// @Param content_id path string true "content id"
// @Param RejectReasonRequest body RejectReasonRequest true "reject_reason"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /contents/{content_id}/review/reject [put]
func (s *Server) reject(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")
	if cid == "" {
		c.JSON(http.StatusBadRequest, "cid can't be empty string")
		return
	}
	var req RejectReasonRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, "can't bind data")
		return
	}
	// extract reject reason
	err = model.GetReviewerModel().Reject(ctx, dbo.MustGetDB(ctx), cid, req.Reasons, req.Remark, op)
	switch err {
	case model.ErrNoContent:
		log.Error(ctx, "reject", log.Any("op", op), log.String("cid", cid), log.Err(err))
		c.JSON(http.StatusNotFound, "content not found")
		return
	case nil:
		c.JSON(http.StatusOK, "ok")
		return
	default:
		log.Error(ctx, "reject", log.Any("op", op), log.String("cid", cid), log.Err(err))
		c.JSON(http.StatusInternalServerError, "Internal server error")
		return
	}
}
