package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

// @ID createLearningOutcomes
// @Summary createOutcome
// @Tags learning_outcomes
// @Description Create learning outcomes
// @Accept json
// @Produce json
// @Param outcome body OutcomeCreateView true "create outcome"
// @Success 200 {object} OutcomeCreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /learning_outcomes [post]
func (s *Server) createOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	var data OutcomeCreateView
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "createOutcome: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	outcome, err := data.outcome()
	if err != nil {
		log.Warn(ctx, "createOutcome: outcome failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	err = model.GetOutcomeModel().CreateLearningOutcome(ctx, dbo.MustGetDB(ctx), outcome, op)
	data.OutcomeID = outcome.ID
	switch err {
	//case model.ErrInvalidResourceId:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrResourceNotFound:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrInvalidContentType:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, newOutcomeCreateResponse(ctx, &data, outcome))
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
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /learning_outcomes/{outcome_id} [get]
func (s *Server) getOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	outcomeID := c.Param("id")
	outcome, err := model.GetOutcomeModel().GetLearningOutcomeByID(ctx, dbo.MustGetDB(ctx), outcomeID, op)
	switch err {
	//case model.ErrInvalidResourceId:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrInvalidContentType:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, newOutcomeView(ctx, outcome))
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
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /learning_outcomes/{outcome_id} [put]
func (s *Server) updateOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	outcomeID := c.Param("id")
	var data OutcomeCreateView
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "updateOutcome: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	outcome, err := data.outcomeWithID(outcomeID)
	if err != nil {
		log.Warn(ctx, "updateOutcome: outcome failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	err = model.GetOutcomeModel().UpdateLearningOutcome(ctx, outcome, op)
	switch err {
	//case model.ErrInvalidResourceId:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrInvalidContentType:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusNotAcceptable, L(Unknown))
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
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /learning_outcomes/{outcome_id} [delete]
func (s *Server) deleteOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	outcomeID := c.Param("id")
	if outcomeID == "" {
		log.Warn(ctx, "deleteOutcome: outcomeID is null", log.String("outcome_id", outcomeID))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	err := model.GetOutcomeModel().DeleteLearningOutcome(ctx, outcomeID, op)
	switch err {
	//case model.ErrInvalidResourceId:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrResourceNotFound:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrInvalidContentType:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(Unknown))
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
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at)
// @Success 200 {object} OutcomeSearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /learning_outcomes [get]
func (s *Server) queryOutcomes(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	var condition entity.OutcomeCondition
	err := c.ShouldBindQuery(&condition)
	if err != nil {
		log.Warn(ctx, "queryOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	if err != nil {
		log.Warn(ctx, "queryOutcomes: outcome failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	total, outcomes, err := model.GetOutcomeModel().SearchLearningOutcome(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	//case model.ErrInvalidResourceId:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrResourceNotFound:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrInvalidContentType:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, newOutcomeSearchResponse(ctx, total, outcomes))
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
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /learning_outcomes/{outcome_id}/lock [put]
func (s *Server) lockOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	outcomeID := c.Param("id")
	if outcomeID == "" {
		log.Warn(ctx, "lockOutcome: outcomeID is null", log.String("outcome_id", outcomeID))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	newID, err := model.GetOutcomeModel().LockLearningOutcome(ctx, dbo.MustGetDB(ctx), outcomeID, op)
	switch err {
	//case model.ErrInvalidResourceId:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrInvalidPublishStatus:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case model.ErrContentAlreadyLocked:
		c.JSON(http.StatusNotAcceptable, L(Unknown))
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
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /learning_outcomes/{outcome_id}/publish [put]
func (s *Server) publishOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	outcomeID := c.Param("id")
	if outcomeID == "" {
		log.Warn(ctx, "publishOutcome: outcomeID is null", log.String("outcome_id", outcomeID))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	var req PublishOutcomeReq
	err := c.ShouldBindJSON(&req)
	if err.Error() != "EOF" {
		log.Warn(ctx, "publishOutcome: ShouldBindJSON failed", log.String("outcome_id", outcomeID))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	err = model.GetOutcomeModel().PublishLearningOutcome(ctx, outcomeID, req.Scope, op)
	switch err {
	//case model.ErrInvalidResourceId:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case model.ErrInvalidContentStatusToPublish:
		c.JSON(http.StatusNotAcceptable, L(Unknown))
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
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /learning_outcomes/{outcome_id}/approve [put]
func (s *Server) approveOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	outcomeID := c.Param("id")
	err := model.GetOutcomeModel().ApproveLearningOutcome(ctx, outcomeID, op)
	switch err {
	//case model.ErrInvalidResourceId:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrInvalidContentType:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusNotAcceptable, L(Unknown))
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
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /learning_outcomes/{outcome_id}/reject [put]
func (s *Server) rejectOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	outcomeID := c.Param("id")
	var reason OutcomeRejectReq
	err := c.ShouldBindJSON(&reason)
	if err != nil {
		log.Warn(ctx, "updateOutcome: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	err = model.GetOutcomeModel().RejectLearningOutcome(ctx, dbo.MustGetDB(ctx), outcomeID, reason.RejectReason, op)
	switch err {
	//case model.ErrInvalidResourceId:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrInvalidContentType:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusNotAcceptable, L(Unknown))
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
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bulk_publish/learning_outcomes [put]
func (s *Server) bulkPublishOutcomes(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	var data struct {
		OutcomeIDs []string `json:"outcome_ids"`
	}
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "bulkPublishOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	err = model.GetOutcomeModel().BulkPubLearningOutcome(ctx, dbo.MustGetDB(ctx), data.OutcomeIDs, "", op)
	switch err {
	case model.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
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
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bulk/learning_outcomes [delete]
func (s *Server) bulkDeleteOutcomes(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	var data struct {
		OutcomeIDs []string `json:"outcome_ids"`
	}
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "bulkPublishOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	err = model.GetOutcomeModel().BulkDelLearningOutcome(ctx, dbo.MustGetDB(ctx), data.OutcomeIDs, op)
	switch err {
	//case model.ErrInvalidResourceId:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrNoContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case model.ErrInvalidContentData:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequireContentName:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrRequirePublishScope:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	//case entity.ErrInvalidContentType:
	//	c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
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
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at)
// @Success 200 {object} OutcomeSearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /private_learning_outcomes [get]
func (s *Server) queryPrivateOutcomes(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	var condition entity.OutcomeCondition
	err := c.ShouldBindQuery(&condition)
	if err != nil {
		log.Warn(ctx, "queryPrivateOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	total, outcomes, err := model.GetOutcomeModel().SearchPrivateOutcomes(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	case model.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, newOutcomeSearchResponse(ctx, total, outcomes))
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
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at)
// @Success 200 {object} OutcomeSearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /pending_learning_outcomes [get]
func (s *Server) queryPendingOutcomes(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	var condition entity.OutcomeCondition
	err := c.ShouldBindQuery(&condition)
	if err != nil {
		log.Warn(ctx, "queryPrivateOutcomes: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	total, outcomes, err := model.GetOutcomeModel().SearchPendingOutcomes(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	case model.ErrBadRequest:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, newOutcomeSearchResponse(ctx, total, outcomes))
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}
