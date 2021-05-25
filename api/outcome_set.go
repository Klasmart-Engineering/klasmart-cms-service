package api

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @ID createOutcomeSet
// @Summary createOutcomeSet
// @Tags outcome_set
// @Description Create learning outcome sets
// @Accept json
// @Produce json
// @Param outcome body model.OutcomeSetCreateView true "create outcome set"
// @Success 200 {object} model.OutcomeSetCreateView
// @Failure 400 {object} BadRequestResponse
// @Failure 401 {object} UnAuthorizedResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 409 {object} ConflictResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /sets [post]
func (s *Server) createOutcomeSet(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data model.OutcomeSetCreateView
	err := c.ShouldBindJSON(&data)
	data.SetName = strings.TrimSpace(data.SetName)
	if err != nil || data.SetName == "" {
		log.Warn(ctx, "createOutcomeSet: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.CreateLearningOutcome,
		external.EditMyUnpublishedLearningOutcome,
		external.EditOrgUnpublishedLearningOutcome,
		external.EditPublishedLearningOutcome,
	})
	if err != nil {
		log.Warn(ctx, "createOutcomeSet: HasOrganizationPermissions failed", log.Any("op", op), log.Any("data", data), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	hasPerm := false
	for _, v := range perms {
		if v == true {
			hasPerm = true
			break
		}
	}
	if !hasPerm {
		log.Warn(ctx, "createOutcomeSet: no permission", log.Any("op", op))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	data.SetID, err = model.GetOutcomeSetModel().CreateOutcomeSet(ctx, op, data.SetName)
	switch err {
	case nil:
		c.JSON(http.StatusOK, data)
	case constant.ErrDuplicateRecord:
		c.JSON(http.StatusConflict, L(AssessMsgExistingSet))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

type PullOutcomeSetRequest struct {
	SetName string `json:"set_name" form:"set_name"`
}
type PullOutcomeSetResponse struct {
	Sets []*model.OutcomeSetCreateView `json:"sets" form:"sets"`
}

// @ID pullOutcomeSet
// @Summary getOutcomeSet
// @Tags outcome_set
// @Description outcome_set info
// @Accept json
// @Produce json
// @Param set_name query string true "search by name"
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at, updated_at, -updated_at)
// @Success 200 {object} PullOutcomeSetResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /sets [get]
func (s *Server) pullOutcomeSet(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var request PullOutcomeSetRequest
	err := c.ShouldBindQuery(&request)
	if err != nil || request.SetName == "" {
		log.Warn(ctx, "pullOutcomeSet: ShouldBindQuery failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.CreateLearningOutcome,
		external.EditMyUnpublishedLearningOutcome,
		external.EditOrgUnpublishedLearningOutcome,
		external.EditPublishedLearningOutcome,
	})
	if err != nil {
		log.Warn(ctx, "pullOutcomeSet: HasOrganizationPermissions failed", log.Any("op", op), log.Err(err), log.Any("req", request))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	hasPerm := false
	for _, v := range perms {
		if v == true {
			hasPerm = true
			break
		}
	}
	if !hasPerm {
		log.Warn(ctx, "pullOutcomeSet: no permission", log.Any("op", op))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	outcomeSets, err := model.GetOutcomeSetModel().PullOutcomeSet(ctx, op, request.SetName)
	var response PullOutcomeSetResponse
	response.Sets = make([]*model.OutcomeSetCreateView, len(outcomeSets))
	for i := range outcomeSets {
		set := model.OutcomeSetCreateView{
			SetID:   outcomeSets[i].ID,
			SetName: outcomeSets[i].Name,
		}
		response.Sets[i] = &set
	}
	switch err {
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, response)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

type BulkBindOutcomeSetRequest struct {
	OutcomeIDs []string `json:"outcome_ids" form:"outcome_ids"`
	SetIDs     []string `json:"set_ids" form:"set_ids"`
}

// @ID bulkBindOutcomeSet
// @Summary bind learning outcome set
// @Tags outcome_set
// @Description bulk bind learning outcome
// @Accept json
// @Produce json
// @Param outcome body BulkBindOutcomeSetRequest true "learning outcome"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 406 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /sets/bulk_bind [post]
func (s *Server) bulkBindOutcomeSet(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var request BulkBindOutcomeSetRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Warn(ctx, "bulkBindOutcomeSet: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.CreateLearningOutcome,
		external.EditMyUnpublishedLearningOutcome,
		external.EditOrgUnpublishedLearningOutcome,
		external.EditPublishedLearningOutcome,
	})
	if err != nil {
		log.Warn(ctx, "bulkBindOutcomeSet: HasOrganizationPermissions failed",
			log.Err(err),
			log.Any("op", op),
			log.Strings("outcome", request.OutcomeIDs),
			log.Strings("set", request.SetIDs))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	hasPerm := false
	for _, v := range perms {
		if v == true {
			hasPerm = true
			break
		}
	}
	if !hasPerm {
		log.Warn(ctx, "bulkBindOutcomeSet: no permission", log.Any("op", op))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	err = model.GetOutcomeSetModel().BulkBindOutcomeSet(ctx, op, request.OutcomeIDs, request.SetIDs)
	switch err {
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(AssessMsgOneStudent))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusNotAcceptable, L(GeneralUnknown))
	case constant.ErrHasLocked:
		c.JSON(http.StatusNotAcceptable, L(AssessMsgLockedLo))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
