package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

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
		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"list":  outcomes,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

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
		c.JSON(http.StatusOK, gin.H{
			"outcome_id": newID,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) publishOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	outcomeID := c.Param("id")
	if outcomeID == "" {
		log.Warn(ctx, "publishOutcome: outcomeID is null", log.String("outcome_id", outcomeID))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	err := model.GetOutcomeModel().PublishLearningOutcome(ctx, outcomeID, "", op)
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
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) rejectOutcome(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	outcomeID := c.Param("id")
	var reason struct {
		RejectReason string `json:"reject_reason"'`
	}
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
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

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
		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"list":  outcomes,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

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
		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"list":  outcomes,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}
