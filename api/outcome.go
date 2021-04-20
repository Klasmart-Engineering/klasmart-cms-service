package api

import (
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @ID createLearningOutcomes
// @Summary createLearningOutcome
// @Tags learning_outcomes
// @Description Create learning outcomes
// @Accept json
// @Produce json
// @Param outcome body model.OutcomeCreateView true "create outcome"
// @Success 200 {object} model.OutcomeCreateResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes [post]
func (s *Server) createOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data model.OutcomeCreateView
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "createOutcome: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.CreateLearningOutcome)
	if err != nil {
		log.Warn(ctx, "createOutcome: HasOrganizationPermission failed", log.Any("op", op), log.Any("data", data), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "createOutcome: no permission", log.Any("op", op), log.String("perm", string(external.CreateLearningOutcome)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	outcome, err := data.ToOutcome()
	if err != nil {
		log.Warn(ctx, "createOutcome: outcome failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err = model.GetOutcomeModel().CreateLearningOutcome(ctx, op, outcome)
	data.OutcomeID = outcome.ID
	switch err {
	case nil:
		c.JSON(http.StatusOK, model.NewCreateResponse(ctx, op, &data, outcome))
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(AssessMsgExistShortcode))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID getLearningOutcomesById
// @Summary getLearningOutcome
// @Tags learning_outcomes
// @Description learning outcomes info
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Success 200 {object} model.OutcomeView
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes/{outcome_id} [get]
func (s *Server) getOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	outcomeID := c.Param("id")
	outcome, err := model.GetOutcomeModel().GetLearningOutcomeByID(ctx, op, dbo.MustGetDB(ctx), outcomeID)
	switch err {
	//case model.ErrInvalidResourceID:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//case entity.ErrInvalidContentType:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		views, err := model.FillOutcomeViews(ctx, op, []*entity.Outcome{outcome})
		if err != nil {
			log.Error(ctx, "getOutcome: FillOutcomeViews failed", log.Any("op", op), log.Any("outcome", outcome))
			return
		}
		c.JSON(http.StatusOK, views[0])
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID updateLearningOutcomes
// @Summary update learning outcome
// @Tags learning_outcomes
// @Description update learning outcomes by id
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Param outcome body model.OutcomeCreateView true "learning outcome"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 409 {object} ConflictResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes/{outcome_id} [put]
func (s *Server) updateOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	outcomeID := c.Param("id")
	var data model.OutcomeCreateView
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "updateOutcome: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	outcome, err := data.ToOutcomeWithID(outcomeID)
	if err != nil {
		log.Warn(ctx, "updateOutcome: outcome failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	// permission check has to delegated to business lay for recognizing org's permission or author's permission
	err = model.GetOutcomeModel().UpdateLearningOutcome(ctx, op, outcome)
	switch err {
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusNotAcceptable, L(GeneralUnknown))
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(AssessMsgExistShortcode))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID deleteLearningOutcome
// @Summary delete learning outcome
// @Tags learning_outcomes
// @Description delete learning outcomes by id
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 406 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes/{outcome_id} [delete]
func (s *Server) deleteOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	outcomeID := c.Param("id")
	if outcomeID == "" {
		log.Warn(ctx, "deleteOutcome: outcomeID is null", log.String("outcome_id", outcomeID))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err := model.GetOutcomeModel().DeleteLearningOutcome(ctx, op, outcomeID)
	lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
	if ok {
		c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
		return
	}
	switch err {
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID searchLearningOutcomes
// @Summary search learning outcome
// @Tags learning_outcomes
// @Description search learning outcomes
// @Accept json
// @Produce json
// @Param outcome_name query string false "search by name"
// @Param description query string false "search by description"
// @Param keywords query string false "search by keywords"
// @Param shortcode query string false "search by shortcode"
// @Param author_name query string false "search by author_name"
// @Param set_name query string false "search by set_name"
// @Param search_key query string false "search by search_key"
// @Param assumed query integer false "search by assumed: 1 true, 0 false, -1 all"
// @Param publish_status query string false "search by publish_status" Enums(draft, pending, published, rejected)
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at, updated_at, -updated_at)
// @Success 200 {object} model.OutcomeSearchResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes [get]
// search public outcomes as a general user
func (s *Server) queryOutcomes(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var condition entity.OutcomeCondition
	err := c.ShouldBindQuery(&condition)
	if err != nil {
		log.Warn(ctx, "queryOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ViewPublishedLearningOutcome)
	if err != nil {
		log.Error(ctx, "queryOutcomes: HasOrganizationPermission failed", log.Any("op", op),
			log.String("perm", string(external.ViewPublishedLearningOutcome)), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "queryOutcomes: HasOrganizationPermission failed", log.Any("op", op),
			log.String("perm", string(external.ViewPublishedLearningOutcome)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	total, outcomes, err := model.GetOutcomeModel().SearchLearningOutcome(ctx, op, dbo.MustGetDB(ctx), &condition)
	switch err {
	//case model.ErrInvalidResourceID:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//case model.ErrResourceNotFound:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//case entity.ErrInvalidContentType:
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		response, err := model.NewSearchResponse(ctx, op, total, outcomes)
		if err != nil {
			log.Error(ctx, "queryOutcomes: NewSearchResponse failed",
				log.Any("op", op),
				log.Any("outcome", outcomes))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, response)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID lockLearningOutcomes
// @Summary lock learning outcome
// @Tags learning_outcomes
// @Description edit published learning outcomes
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Success 200 {string} model.OutcomeLockResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 406 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes/{outcome_id}/lock [put]
func (s *Server) lockOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	outcomeID := c.Param("id")
	if outcomeID == "" {
		log.Warn(ctx, "lockOutcome: outcomeID is null", log.String("outcome_id", outcomeID))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.EditPublishedLearningOutcome)
	if err != nil {
		log.Error(ctx, "lockOutcome: HasOrganizationPermission failed", log.String("outcome_id", outcomeID), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "lockOutcome: HasOrganizationPermission failed", log.Any("op", op),
			log.String("perm", string(external.EditPublishedLearningOutcome)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	newID, err := model.GetOutcomeModel().LockLearningOutcome(ctx, op, dbo.MustGetDB(ctx), outcomeID)
	lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
	if ok {
		c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
		return
	}
	switch err {
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, model.OutcomeLockResponse{OutcomeID: newID})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID publishLearningOutcomes
// @Summary publish learning outcome
// @Tags learning_outcomes
// @Description submit publish learning outcomes
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Param PublishOutcomeRequest body model.PublishOutcomeReq false "publish scope"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes/{outcome_id}/publish [put]
func (s *Server) publishOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	outcomeID := c.Param("id")
	if outcomeID == "" {
		log.Warn(ctx, "publishOutcome: outcomeID is null", log.String("outcome_id", outcomeID))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	var req model.PublishOutcomeReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warn(ctx, "publishOutcome: ShouldBindJSON failed", log.String("outcome_id", outcomeID),
			log.Any("req", req),
			log.Any("op", op))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	req.Scope = op.OrgID
	err = model.GetOutcomeModel().PublishLearningOutcome(ctx, op, outcomeID, req.Scope)

	switch err {
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidContentStatusToPublish:
		c.JSON(http.StatusNotAcceptable, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID approveLearningOutcomes
// @Summary approve learning outcome
// @Tags learning_outcomes
// @Description approve learning outcomes
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes/{outcome_id}/approve [put]
func (s *Server) approveOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	outcomeID := c.Param("id")
	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ApprovePendingLearningOutcome)
	if err != nil {
		log.Error(ctx, "approveOutcome: HasOrganizationPermission failed", log.String("id", outcomeID), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "approveOutcome: no permission",
			log.Any("op", op), log.String("id", outcomeID),
			log.String("perm", string(external.ApprovePendingLearningOutcome)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	err = model.GetOutcomeModel().ApproveLearningOutcome(ctx, op, outcomeID)
	switch err {
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusNotAcceptable, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID rejectLearningOutcomes
// @Summary reject learning outcome
// @Tags learning_outcomes
// @Description reject learning outcomes
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Param OutcomeRejectReq body model.OutcomeRejectReq true "reject reason"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes/{outcome_id}/reject [put]
func (s *Server) rejectOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	outcomeID := c.Param("id")
	var reason model.OutcomeRejectReq
	err := c.ShouldBindJSON(&reason)
	if err != nil {
		log.Warn(ctx, "rejectOutcome: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.RejectPendingLearningOutcome)
	if err != nil {
		log.Error(ctx, "rejectOutcome: HasOrganizationPermission failed", log.String("id", outcomeID), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "rejectOutcome: no permission",
			log.Any("op", op), log.String("id", outcomeID),
			log.String("perm", string(external.RejectPendingLearningOutcome)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	err = model.GetOutcomeModel().RejectLearningOutcome(ctx, op, dbo.MustGetDB(ctx), outcomeID, reason.RejectReason)
	switch err {
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusNotAcceptable, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID approveLearningOutcomesBulk
// @Summary bulk approve learning outcome
// @Tags learning_outcomes
// @Description approve learning outcomes
// @Accept json
// @Produce json
// @Param id_list body model.OutcomeIDList true "outcome id list"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /bulk_approve/learning_outcomes [put]
func (s *Server) bulkApproveOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data model.OutcomeIDList
	err := c.ShouldBindJSON(&data)
	if err != nil || len(data.OutcomeIDs) == 0 {
		log.Warn(ctx, "bulkApproveOutcome: ShouldBind failed", log.Any("req", data), log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ApprovePendingLearningOutcome)
	if err != nil {
		log.Error(ctx, "approveOutcome: HasOrganizationPermission failed", log.Strings("ids", data.OutcomeIDs), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "approveOutcome: no permission",
			log.Any("op", op), log.Strings("ids", data.OutcomeIDs),
			log.String("perm", string(external.ApprovePendingLearningOutcome)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	err = model.GetOutcomeModel().BulkApproveLearningOutcome(ctx, op, utils.SliceDeduplication(data.OutcomeIDs))
	switch err {
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID rejectLearningOutcomesBulk
// @Summary bulk reject learning outcome
// @Tags learning_outcomes
// @Description reject learning outcomes
// @Accept json
// @Produce json
// @Param bulk_reject_list body model.OutcomeBulkRejectRequest true "outcome id list"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /bulk_reject/learning_outcomes [put]
func (s *Server) bulkRejectOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data model.OutcomeBulkRejectRequest
	err := c.ShouldBindJSON(&data)
	if err != nil || len(data.OutcomeIDs) == 0 {
		log.Warn(ctx, "bulkRejectOutcome: ShouldBind failed", log.Any("req", data), log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.RejectPendingLearningOutcome)
	if err != nil {
		log.Error(ctx, "rejectOutcome: HasOrganizationPermission failed", log.Strings("ids", data.OutcomeIDs), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "rejectOutcome: no permission",
			log.Any("op", op), log.Strings("ids", data.OutcomeIDs),
			log.String("perm", string(external.RejectPendingLearningOutcome)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	err = model.GetOutcomeModel().BulkRejectLearningOutcome(ctx, op, utils.SliceDeduplication(data.OutcomeIDs), data.RejectReason)
	switch err {
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID publishLearningOutcomesBulk
// @Summary publish bulk learning outcome
// @Tags learning_outcomes
// @Description submit publish learning outcomes
// @Accept json
// @Produce json
// @Param id_list body model.OutcomeIDList true "outcome id list"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /bulk_publish/learning_outcomes [put]
func (s *Server) bulkPublishOutcomes(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data struct {
		OutcomeIDs []string `json:"outcome_ids"`
	}
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "bulkPublishOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = model.GetOutcomeModel().BulkPubLearningOutcome(ctx, op, dbo.MustGetDB(ctx), data.OutcomeIDs, "")
	switch err {
	case model.ErrInvalidResourceID:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID deleteOutcomeBulk
// @Summary bulk delete learning outcome
// @Tags learning_outcomes
// @Description bulk delete learning outcomes
// @Accept json
// @Produce json
// @Param id_list body model.OutcomeIDList true "outcome id list"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /bulk/learning_outcomes [delete]
func (s *Server) bulkDeleteOutcomes(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data struct {
		OutcomeIDs []string `json:"outcome_ids"`
	}
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "bulkPublishOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err = model.GetOutcomeModel().BulkDelLearningOutcome(ctx, op, dbo.MustGetDB(ctx), data.OutcomeIDs)
	_, ok := err.(*model.ErrContentAlreadyLocked)
	if ok {
		c.JSON(http.StatusNotAcceptable, L(AssessMsgLockedLo))
		return
	}
	switch err {
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID searchPrivateLearningOutcomes
// @Summary search private learning outcome
// @Tags learning_outcomes
// @Description search private learning outcomes
// @Accept json
// @Produce json
// @Param outcome_name query string false "search by name"
// @Param description query string false "search by description"
// @Param keywords query string false "search by keywords"
// @Param shortcode query string false "search by shortcode"
// @Param author_name query string false "search by author_name"
// @Param search_key query string false "search by search_key"
// @Param assumed query integer false "search by assumed: 1 true, 0 false, -1 all"
// @Param publish_status query string false "search by publish_status" Enums(draft, pending, published, rejected)
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at, updated_at, -updated_at)
// @Success 200 {object} model.OutcomeSearchResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /private_learning_outcomes [get]
// search private outcomes as an author user
func (s *Server) queryPrivateOutcomes(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var condition entity.OutcomeCondition
	err := c.ShouldBindQuery(&condition)
	if err != nil {
		log.Warn(ctx, "queryPrivateOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	total, outcomes, err := model.GetOutcomeModel().SearchPrivateOutcomes(ctx, op, dbo.MustGetDB(ctx), &condition)
	switch err {
	case model.ErrInvalidResourceID:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		response, err := model.NewSearchResponse(ctx, op, total, outcomes)
		if err != nil {
			log.Error(ctx, "queryPrivateOutcomes: NewSearchResponse failed",
				log.Any("op", op),
				log.Any("outcome", outcomes))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, response)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID searchPendingLearningOutcomes
// @Summary search pending learning outcome
// @Tags learning_outcomes
// @Description search pending learning outcomes
// @Accept json
// @Produce json
// @Param outcome_name query string false "search by name"
// @Param description query string false "search by description"
// @Param keywords query string false "search by keywords"
// @Param shortcode query string false "search by shortcode"
// @Param author_name query string false "search by author_name"
// @Param search_key query string false "search by search_key"
// @Param assumed query integer false "search by assumed: 1 true, 0 false, -1 all"
// @Param publish_status query string false "search by publish_status" Enums(draft, pending, published, rejected)
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at, updated_at, -updated_at)
// @Success 200 {object} model.OutcomeSearchResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /pending_learning_outcomes [get]
// search pending outcomes as an admin user
func (s *Server) queryPendingOutcomes(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var condition entity.OutcomeCondition
	err := c.ShouldBindQuery(&condition)
	if err != nil {
		log.Warn(ctx, "queryPendingOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	total, outcomes, err := model.GetOutcomeModel().SearchPendingOutcomes(ctx, op, dbo.MustGetDB(ctx), &condition)
	switch err {
	case model.ErrBadRequest:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidResourceID:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case nil:
		response, err := model.NewSearchResponse(ctx, op, total, outcomes)
		if err != nil {
			log.Error(ctx, "queryPendingOutcomes: NewSearchResponse failed",
				log.Any("op", op),
				log.Any("outcome", outcomes))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, response)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

type ShortcodeRequest struct {
	Kind entity.ShortcodeKind `json:"kind" form:"kind"`
}
type ShortcodeResponse struct {
	Shortcode string `json:"shortcode" form:"shortcode"`
}

// @ID generateShortcode
// @Summary generate Shortcode
// @Tags learning_outcomes
// @Description generate shortcode
// @Accept json
// @Produce json
// @Param kind body ShortcodeRequest false "learning outcome"
// @Success 200 {object} ShortcodeResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 409 {object} ConflictResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /shortcode [post]
func (s *Server) generateShortcode(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var data ShortcodeRequest
	err := c.ShouldBindJSON(&data)
	if err != nil && err.Error() != "EOF" {
		log.Warn(ctx, "generateShortcode: ShouldBindJSON failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if err != nil && err.Error() == "EOF" {
		data.Kind = entity.KindOutcome
	}
	if data.Kind != entity.KindOutcome && data.Kind != entity.KindMileStone {
		log.Warn(ctx, "generateShortcode: kind not allowed", log.Any("shortcode_kind", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.CreateLearningOutcome)
	if err != nil {
		log.Warn(ctx, "generateShortcode: HasOrganizationPermission failed", log.Any("op", op), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "generateShortcode: no permission", log.Any("op", op), log.String("perm", string(external.CreateLearningOutcome)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	shortcode, err := model.GetShortcodeModel().Generate(ctx, dbo.MustGetDB(ctx), data.Kind, op.OrgID, "")
	switch err {
	case nil:
		c.JSON(http.StatusOK, &ShortcodeResponse{Shortcode: shortcode})
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(AssessMsgExistShortcode))
	case constant.ErrExceededLimit:
		c.JSON(http.StatusConflict, L(AssessMsgExistShortcode))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
