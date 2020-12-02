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
// @Summary createOutcome
// @Tags learning_outcomes
// @Description Create learning outcomes
// @Accept json
// @Produce json
// @Param outcome body OutcomeCreateView true "create outcome"
// @Success 200 {object} OutcomeCreateResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes [post]
func (s *Server) createOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data OutcomeCreateView
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
	outcome, err := data.outcome()
	if err != nil {
		log.Warn(ctx, "createOutcome: outcome failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err = model.GetOutcomeModel().CreateLearningOutcome(ctx, dbo.MustGetDB(ctx), outcome, op)
	data.OutcomeID = outcome.ID
	switch err {
	case nil:
		c.JSON(http.StatusOK, newOutcomeCreateResponse(ctx, op, &data, outcome))
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @ID getLearningOutcomesById
// @Summary getLearningOutcome
// @Tags learning_outcomes
// @Description learning outcomes info
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Success 200 {object} OutcomeView
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes/{outcome_id} [get]
func (s *Server) getOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	outcomeID := c.Param("id")
	outcome, err := model.GetOutcomeModel().GetLearningOutcomeByID(ctx, dbo.MustGetDB(ctx), outcomeID, op)
	switch err {
	//case model.ErrInvalidResourceId:
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
		c.JSON(http.StatusOK, newOutcomeView(ctx, op, outcome))
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @ID updateLearningOutcomes
// @Summary update learning outcome
// @Tags learning_outcomes
// @Description update learning outcomes by id
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Param outcome body OutcomeCreateView true "learning outcome"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /learning_outcomes/{outcome_id} [put]
func (s *Server) updateOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	outcomeID := c.Param("id")
	var data OutcomeCreateView
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "updateOutcome: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	outcome, err := data.outcomeWithID(outcomeID)
	if err != nil {
		log.Warn(ctx, "updateOutcome: outcome failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	// permission check has to delegated to business lay for recognizing org's permission or author's permission
	err = model.GetOutcomeModel().UpdateLearningOutcome(ctx, outcome, op)
	switch err {
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(AssessMsgOneStudent))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusNotAcceptable, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
// @Failure 404 {object} NotFoundResponse
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

	err := model.GetOutcomeModel().DeleteLearningOutcome(ctx, outcomeID, op)
	switch err {
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
// @Param search_key query string false "search by search_key"
// @Param assumed query integer false "search by assumed: 1 true, 0 false, -1 all"
// @Param publish_status query string false "search by publish_status" Enums(draft, pending, published, rejected)
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at, updated_at, -updated_at)
// @Success 200 {object} OutcomeSearchResponse
// @Failure 400 {object} BadRequestResponse
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
	total, outcomes, err := model.GetOutcomeModel().SearchLearningOutcome(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	//case model.ErrInvalidResourceId:
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
		c.JSON(http.StatusOK, newOutcomeSearchResponse(ctx, op, total, outcomes))
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @ID lockLearningOutcomes
// @Summary lock learning outcome
// @Tags learning_outcomes
// @Description edit published learning outcomes
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Success 200 {string} OutcomeLockResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
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
	newID, err := model.GetOutcomeModel().LockLearningOutcome(ctx, dbo.MustGetDB(ctx), outcomeID, op)
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
		c.JSON(http.StatusOK, OutcomeLockResponse{newID})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @ID publishLearningOutcomes
// @Summary publish learning outcome
// @Tags learning_outcomes
// @Description submit publish learning outcomes
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Param PublishOutcomeRequest body PublishOutcomeReq false "publish scope"
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

	var req PublishOutcomeReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warn(ctx, "publishOutcome: ShouldBindJSON failed", log.String("outcome_id", outcomeID),
			log.Any("req", req),
			log.Any("op", op))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	req.Scope = op.OrgID
	err = model.GetOutcomeModel().PublishLearningOutcome(ctx, outcomeID, req.Scope, op)

	switch err {
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidContentStatusToPublish:
		c.JSON(http.StatusNotAcceptable, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
	err = model.GetOutcomeModel().ApproveLearningOutcome(ctx, outcomeID, op)
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
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @ID rejectLearningOutcomes
// @Summary reject learning outcome
// @Tags learning_outcomes
// @Description reject learning outcomes
// @Accept json
// @Produce json
// @Param outcome_id path string true "outcome id"
// @Param OutcomeRejectReq body OutcomeRejectReq true "reject reason"
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
	var reason OutcomeRejectReq
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
	err = model.GetOutcomeModel().RejectLearningOutcome(ctx, dbo.MustGetDB(ctx), outcomeID, reason.RejectReason, op)
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
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @ID approveLearningOutcomesBulk
// @Summary bulk approve learning outcome
// @Tags learning_outcomes
// @Description approve learning outcomes
// @Accept json
// @Produce json
// @Param id_list body OutcomeIDList true "outcome id list"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /bulk_approve/learning_outcomes [put]
func (s *Server) bulkApproveOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data OutcomeIDList
	err := c.ShouldBindJSON(&data)
	if err != nil || len(data.OutcomeIDs) == 0 {
		log.Warn(ctx, "bulkApproveOutcome: ShouldBind failed", log.Any("req", data), log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = model.GetOutcomeModel().BulkApproveLearningOutcome(ctx, utils.SliceDeduplication(data.OutcomeIDs), op)
	switch err {
	case model.ErrInvalidResourceId:
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
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @ID rejectLearningOutcomesBulk
// @Summary bulk reject learning outcome
// @Tags learning_outcomes
// @Description reject learning outcomes
// @Accept json
// @Produce json
// @Param id_list body OutcomeBulkRejectRequest true "outcome id list"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /bulk_reject/learning_outcomes [put]
func (s *Server) bulkRejectOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data OutcomeBulkRejectRequest
	err := c.ShouldBindJSON(&data)
	if err != nil || len(data.OutcomeIDs) == 0 {
		log.Warn(ctx, "bulkRejectOutcome: ShouldBind failed", log.Any("req", data), log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = model.GetOutcomeModel().BulkRejectLearningOutcome(ctx, utils.SliceDeduplication(data.OutcomeIDs), data.RejectReason, op)
	switch err {
	case model.ErrInvalidResourceId:
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
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @ID publishLearningOutcomesBulk
// @Summary publish bulk learning outcome
// @Tags learning_outcomes
// @Description submit publish learning outcomes
// @Accept json
// @Produce json
// @Param id_list body OutcomeIDList true "outcome id list"
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

	err = model.GetOutcomeModel().BulkPubLearningOutcome(ctx, dbo.MustGetDB(ctx), data.OutcomeIDs, "", op)
	switch err {
	case model.ErrInvalidResourceId:
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
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @ID deleteOutcomeBulk
// @Summary bulk delete learning outcome
// @Tags learning_outcomes
// @Description bulk delete learning outcomes
// @Accept json
// @Produce json
// @Param id_list body OutcomeIDList true "outcome id list"
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
	err = model.GetOutcomeModel().BulkDelLearningOutcome(ctx, dbo.MustGetDB(ctx), data.OutcomeIDs, op)
	switch err {
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
// @Success 200 {object} OutcomeSearchResponse
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

	total, outcomes, err := model.GetOutcomeModel().SearchPrivateOutcomes(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	case model.ErrInvalidResourceId:
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
		c.JSON(http.StatusOK, newOutcomeSearchResponse(ctx, op, total, outcomes))
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
// @Success 200 {object} OutcomeSearchResponse
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
		log.Warn(ctx, "queryPrivateOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	total, outcomes, err := model.GetOutcomeModel().SearchPendingOutcomes(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	case model.ErrBadRequest:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidResourceId:
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
		c.JSON(http.StatusOK, newOutcomeSearchResponse(ctx, op, total, outcomes))
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}
