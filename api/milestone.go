package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

// @ID createMilestone
// @Summary create milestone
// @Tags milestone
// @Description Create milestone
// @Accept json
// @Produce json
// @Param milestone body MilestoneView true "create milestone"
// @Success 200 {object} MilestoneView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /milestones [post]
func (s *Server) createMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data MilestoneView
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "createMilestone: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.CreateMilestone)
	if err != nil {
		log.Warn(ctx, "createMilestone: HasOrganizationPermission failed", log.Any("op", op), log.Any("data", data), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "createMilestone: no permission", log.Any("op", op), log.String("perm", string(external.CreateMilestone)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	milestone := data.toMilestone(op)
	if err != nil {
		log.Warn(ctx, "createMilestone: outcome failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err = model.GetMilestoneModel().Create(ctx, op, milestone, data.OutcomeIDs)
	data.MilestoneID = milestone.ID
	switch err {
	case nil:
		c.JSON(http.StatusOK, milestone)
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(AssessMsgExistShortcode))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID obtainMilestone
// @Summary get milestone by id
// @Tags milestone
// @Description milestone info
// @Accept json
// @Produce json
// @Param milestone_id path string true "milestone id"
// @Success 200 {object} MilestoneView
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /milestones/{milestone_id} [get]
func (s *Server) obtainMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	milestoneID := c.Param("id")
	if milestoneID == "" {
		log.Warn(ctx, "obtainMilestone: illegal param", log.String("milestone", milestoneID))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	milestone, err := model.GetMilestoneModel().Obtain(ctx, op, milestoneID)

	switch err {
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		view, err := fromMilestone(ctx, op, milestone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, view)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID updateMilestone
// @Summary update milestone
// @Tags milestone
// @Description update milestone info
// @Accept json
// @Produce json
// @Param milestone_id path string true "milestone id"
// @Param milestone body MilestoneView true "milestone"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 409 {object} ConflictResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /milestones/{milestone_id} [put]
func (s *Server) updateMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	milestoneID := c.Param("id")
	var data MilestoneView
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "updateMilestone: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	data.MilestoneID = milestoneID

	milestone := data.toMilestone(op)
	if err != nil {
		log.Warn(ctx, "updateMilestone: toMilestone failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.EditUnpublishedMilestone, external.EditPublishedMilestone,
	})
	if err != nil {
		log.Error(ctx, "updateMilestone: HasOrganizationPermission failed", log.Any("op", op), log.Any("data", data), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	err = model.GetMilestoneModel().Update(ctx, op, perms, milestone, data.OutcomeIDs)
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

// @ID deleteMilestone
// @Summary delete milestone
// @Tags milestone
// @Description delete milestone
// @Accept json
// @Produce json
// @Param milestones body MilestoneList true "delete milestone"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 406 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /milestones/{milestone_id} [delete]
func (s *Server) deleteMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data MilestoneList
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "deleteMilestone: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.DeleteUnpublishedMilestone, external.DeletePublishedMilestone,
	})
	if err != nil {
		log.Error(ctx, "deleteMilestone: HasOrganizationPermission failed", log.Any("op", op), log.Any("data", data), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	err = model.GetMilestoneModel().Delete(ctx, op, perms, data.IDs)
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

// @ID searchMilestone
// @Summary search milestone
// @Tags milestone
// @Description search milestone
// @Accept json
// @Produce json
// @Param search_key query string false "search by search_key"
// @Param name query string false "search by name"
// @Param description query string false "search by description"
// @Param shortcode query string false "search by shortcode"
// @Param status query string false "search by publish_status" Enums(draft, published)
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at, updated_at, -updated_at)
// @Success 200 {object} MilestoneSearchResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /milestones [get]
func (s *Server) searchMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var condition entity.MilestoneCondition
	err := c.ShouldBindQuery(&condition)
	if err != nil {
		log.Warn(ctx, "searchMilestone: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	var hasPerm bool
	if condition.Status == entity.OutcomeStatusPublished {
		hasPerm, err = external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ViewPublishedMilestone)
	} else {
		hasPerm, err = external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ViewUnPublishedMilestone)
	}

	if err != nil {
		log.Error(ctx, "searchMilestone: HasOrganizationPermission failed", log.Err(err), log.Any("op", op))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "searchMilestone: HasOrganizationPermission failed", log.Any("op", op))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	total, milestones, err := model.GetMilestoneModel().Search(ctx, op, &condition)
	switch err {
	case nil:
		response, err := fromMilestones(ctx, op, total, milestones)
		if err != nil {
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, response)
	default:
		log.Error(ctx, "searchMilestone: Search failed", log.Err(err), log.Any("op", op), log.Any("cond", condition))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID occupyMilestone
// @Summary lock milestone
// @Tags milestone
// @Description occupy before edit
// @Accept json
// @Produce json
// @Param milestone_id path string true "milestone id"
// @Success 200 {string} MilestoneView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 406 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /milestones/{milestone_id}/occupy [put]
func (s *Server) occupyMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	milestoneID := c.Param("id")
	if milestoneID == "" {
		log.Warn(ctx, "occupyMilestone: milestoneID is null", log.String("milestone", milestoneID))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.EditPublishedMilestone)
	if err != nil {
		log.Error(ctx, "occupyMilestone: HasOrganizationPermission failed", log.String("milestone", milestoneID), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "occupyMilestone: HasOrganizationPermission failed", log.Any("op", op),
			log.String("perm", string(external.EditPublishedMilestone)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	milestone, err := model.GetMilestoneModel().Occupy(ctx, op, milestoneID)
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
		c.JSON(http.StatusOK, MilestoneView{MilestoneID: milestone.ID})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID publishMilestone
// @Summary publish milestone
// @Tags milestone
// @Description publish milestone
// @Accept json
// @Produce json
// @Param milestones body MilestoneList true "publish milestone"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /milestones/publish [put]
func (s *Server) publishMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var data MilestoneList
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "publishMilestone: ShouldBindJSON failed",
			log.Any("op", op))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err = model.GetMilestoneModel().Publish(ctx, op, data.IDs)

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
