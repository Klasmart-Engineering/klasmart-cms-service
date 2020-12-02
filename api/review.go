package api

import (
	"context"
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

type ApproveReasonBulkRequest struct {
	IDs []string `json:"ids"`
}

// @ID approveContentReviewBulk
// @Summary approve content bulk
// @Tags content
// @Description approve content bulk
// @Accept json
// @Produce json
// @Param RejectReasonRequest body RejectReasonBulkRequest true "reject_reason"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /contents/review/approve [put]
func (s *Server) approveBulk(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var req ApproveReasonBulkRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, "can't bind data")
		return
	}
	err = s.checkApproveAuthByCIDs(ctx, req.IDs, external.ApprovePendingContent271, op)
	if err == model.ErrNoAuth {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	// extract reject reason
	err = model.GetReviewerModel().ApproveBulk(ctx, dbo.MustGetDB(ctx), req.IDs, op)
	switch err {
	case model.ErrNoContent:
		log.Error(ctx, "approve", log.Any("op", op), log.Strings("cids", req.IDs), log.Err(err))
		c.JSON(http.StatusNotFound, "content not found")
		return
	case nil:
		c.JSON(http.StatusOK, "ok")
		return
	default:
		log.Error(ctx, "approve", log.Any("op", op), log.Strings("cids", req.IDs), log.Err(err))
		c.JSON(http.StatusInternalServerError, "Internal server error")
		return
	}
}

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
	op := s.getOperator(c)
	cid := c.Param("content_id")
	if cid == "" {
		c.JSON(http.StatusBadRequest, "cid can't be empty string")
		return
	}
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ApprovePendingContent271)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		content, err := model.GetContentModel().GetContentByID(ctx, dbo.MustGetDB(ctx), cid, op)
		if err != nil {
			log.Error(ctx, "approve", log.Any("op", op), log.String("cid", cid), log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}

		hasPermission, err = external.GetPermissionServiceProvider().HasSchoolPermission(ctx, op, content.PublishScope, external.ApprovePendingContent271)
		if err != nil {
			log.Error(ctx, "approve", log.Any("op", op), log.String("cid", cid), log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		if !hasPermission {
			c.JSON(http.StatusForbidden, L(GeneralUnknown))
			return
		}
	}
	err = model.GetReviewerModel().Approve(ctx, dbo.MustGetDB(ctx), cid, op)
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
	op := s.getOperator(c)
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
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.RejectPendingContent272)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		content, err := model.GetContentModel().GetContentByID(ctx, dbo.MustGetDB(ctx), cid, op)
		if err != nil {
			log.Error(ctx, "reject", log.Any("op", op), log.String("cid", cid), log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}

		hasPermission, err = external.GetPermissionServiceProvider().HasSchoolPermission(ctx, op, content.PublishScope, external.ApprovePendingContent271)
		if err != nil {
			log.Error(ctx, "reject", log.Any("op", op), log.String("cid", cid), log.Err(err))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		if !hasPermission {
			c.JSON(http.StatusForbidden, L(GeneralUnknown))
			return
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
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

type RejectReasonBulkRequest struct {
	IDs     []string `json:"ids"`
	Reasons []string `json:"reject_reason"`
	Remark  string   `json:"remark"`
}

// @ID rejectContentReviewBulk
// @Summary reject content bulk
// @Tags content
// @Description reject content bulk
// @Accept json
// @Produce json
// @Param RejectReasonRequest body RejectReasonBulkRequest true "reject_reason"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /contents/review/reject [put]
func (s *Server) rejectBulk(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var req RejectReasonBulkRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, "can't bind data")
		return
	}
	err = s.checkApproveAuthByCIDs(ctx, req.IDs, external.RejectPendingContent272, op)
	if err == model.ErrNoAuth {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	// extract reject reason
	err = model.GetReviewerModel().RejectBulk(ctx, dbo.MustGetDB(ctx), req.IDs, req.Reasons, req.Remark, op)
	switch err {
	case model.ErrNoContent:
		log.Error(ctx, "reject", log.Any("op", op), log.Strings("cids", req.IDs), log.Err(err))
		c.JSON(http.StatusNotFound, "content not found")
		return
	case nil:
		c.JSON(http.StatusOK, "ok")
		return
	default:
		log.Error(ctx, "reject", log.Any("op", op), log.Strings("cids", req.IDs), log.Err(err))
		c.JSON(http.StatusInternalServerError, "Internal server error")
		return
	}
}

func (s *Server) checkApproveAuthByCIDs(ctx context.Context, cids []string, permission external.PermissionName, op *entity.Operator) error {
	//Search content by ids
	cids = utils.SliceDeduplication(cids)
	contentList, err := model.GetContentModel().GetContentByIdList(ctx, dbo.MustGetDB(ctx), cids, op)
	if err != nil {
		return err
	}
	if len(cids) != len(contentList) {
		return model.ErrNoContent
	}

	//Collect all publish scope
	publishScopes := make([]string, len(contentList))
	for i := range contentList {
		publishScopes[i] = contentList[i].PublishScope
	}
	//remove duplicate publish scopes
	publishScopes = utils.SliceDeduplication(publishScopes)

	for i := range publishScopes {
		hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permission)
		if err != nil {
			return err
		}
		//if has Permission continue
		if hasPermission {
			continue
		}

		//not org, maybe it is a school permission
		hasPermission, err = external.GetPermissionServiceProvider().HasSchoolPermission(ctx, op, publishScopes[i], permission)
		if err != nil {
			log.Error(ctx, "approve", log.Any("op", op), log.Strings("cids", cids), log.String("publishScope", publishScopes[i]), log.Err(err))
			return err
		}
		if !hasPermission {
			return model.ErrNoAuth
		}

	}

	return nil
}
