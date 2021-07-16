package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

// @ID createMilestone
// @Summary create milestone
// @Tags milestone
// @Description Create milestone
// @Accept json
// @Produce json
// @Param milestone body model.MilestoneView true "create milestone"
// @Success 200 {object} model.MilestoneView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 409 {object} ConflictResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /milestones [post]
func (s *Server) createMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data model.MilestoneView
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
	milestone, err := data.ToMilestone(ctx, op)
	if err != nil {
		log.Warn(ctx, "createMilestone: ToMilestone failed", log.Err(err), log.Any("op", op), log.Any("req", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err = model.GetMilestoneModel().Create(ctx, op, milestone, utils.SliceDeduplication(data.OutcomeAncestorIDs))
	data.MilestoneID = milestone.ID
	switch err {
	case nil:
		if data.WithPublish {
			c.JSON(http.StatusOK, "ok")
			return
		}
		views, err := model.FromMilestones(ctx, op, []*entity.Milestone{milestone})
		if err != nil {
			log.Error(ctx, "createMilestone: fromMilestones failed",
				log.Any("milestones", views))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, views[0])
	case constant.ErrConflict:
		log.Warn(ctx, "createMilestone: Create failed", log.Any("op", op), log.Any("req", data))
		c.JSON(http.StatusConflict, L(AssessMsgMilestoneExistShortcode))
	default:
		log.Error(ctx, "createMilestone: Create failed", log.Any("op", op), log.Any("req", data))
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
// @Success 200 {object} model.MilestoneView
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
		log.Warn(ctx, "obtainMilestone: Obtain failed", log.Any("op", op), log.String("milestone", milestoneID))
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		views, err := model.FromMilestones(ctx, op, []*entity.Milestone{milestone})
		if err != nil {
			log.Error(ctx, "obtainMilestone: fromMilestones failed",
				log.String("milestone", milestoneID),
				log.Any("milestones", views))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, views[0])
	default:
		log.Error(ctx, "obtainMilestone: Obtain failed", log.String("milestone", milestoneID))
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
// @Param milestone body model.MilestoneView true "milestone"
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
	var data model.MilestoneView
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "updateMilestone: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	data.MilestoneID = milestoneID

	milestone, err := data.ToMilestone(ctx, op)
	if err != nil {
		log.Warn(ctx, "updateMilestone: ToMilestone failed", log.Err(err), log.Any("op", op), log.Any("req", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	permName := []external.PermissionName{external.EditUnpublishedMilestone, external.EditPublishedMilestone}
	if data.WithPublish {
		permName = append(permName, external.CreateMilestone)
	}
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, permName)
	if err != nil {
		log.Error(ctx, "updateMilestone: HasOrganizationPermission failed", log.Any("op", op), log.Any("perm", permName), log.Any("data", data), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	err = model.GetMilestoneModel().Update(ctx, op, perms, milestone, utils.SliceDeduplication(data.OutcomeAncestorIDs))
	switch err {
	case constant.ErrOperateNotAllowed:
		log.Warn(ctx, "updateMilestone: Update failed", log.Any("op", op), log.Any("req", data))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrOutOfDate:
		log.Warn(ctx, "updateMilestone: Update failed", log.Any("op", op), log.Any("req", data))
		c.JSON(http.StatusForbidden, L(AssessMsgUnLockedMilestone))
	case model.ErrResourceNotFound:
		log.Warn(ctx, "updateMilestone: Update failed", log.Any("op", op), log.Any("req", data))
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidPublishStatus:
		log.Warn(ctx, "updateMilestone: Update failed", log.Any("op", op), log.Any("req", data))
		c.JSON(http.StatusConflict, L(GeneralUnknown))
	case constant.ErrConflict:
		log.Warn(ctx, "updateMilestone: Update failed", log.Any("op", op), log.Any("req", data))
		c.JSON(http.StatusConflict, L(AssessMsgMilestoneExistShortcode))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
		if ok {
			user, err := external.GetUserServiceProvider().Get(ctx, op, lockedByErr.LockedBy.ID)
			if err != nil {
				log.Error(ctx, "updateMilestone: GetUserServiceProvider failed",
					log.Err(err),
					log.Any("op", op),
					log.String("req", milestoneID),
					log.String("locked", lockedByErr.LockedBy.ID))
				c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
				return
			}
			log.Warn(ctx, "updateMilestone", log.Any("op", op), log.Any("req", data))
			c.JSON(http.StatusConflict, LD(AssessErrorMsgLocked, user))
			return
		}
		log.Error(ctx, "updateMilestone: Update failed", log.Any("op", op), log.Any("req", data))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID deleteMilestone
// @Summary delete milestone
// @Tags milestone
// @Description delete milestone
// @Accept json
// @Produce json
// @Param milestones body model.MilestoneList true "delete milestone"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /milestones [delete]
func (s *Server) deleteMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data model.MilestoneList
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "deleteMilestone: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	perms, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.DeleteUnpublishedMilestone,
		external.DeletePublishedMilestone,
		external.DeleteMyUnpublishedMilestone,
		external.DeleteOrgPendingMilestone,
		external.DeleteMyPendingMilestone,
	})
	if err != nil {
		log.Error(ctx, "deleteMilestone: HasOrganizationPermission failed", log.Any("op", op), log.Any("data", data), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	err = model.GetMilestoneModel().Delete(ctx, op, perms, data.IDs)
	switch err {
	case constant.ErrOperateNotAllowed:
		log.Warn(ctx, "deleteMilestone: Delete failed", log.Any("op", op), log.Strings("req", data.IDs))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrOutOfDate:
		log.Warn(ctx, "deleteMilestone: Update failed", log.Any("op", op), log.Any("req", data))
		c.JSON(http.StatusForbidden, L(AssessMsgUnLockedMilestone))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
		if ok {
			user, err := external.GetUserServiceProvider().Get(ctx, op, lockedByErr.LockedBy.ID)
			if err != nil {
				log.Error(ctx, "deleteMilestone: Delete failed",
					log.Err(err),
					log.Any("op", op),
					log.Strings("req", data.IDs),
					log.String("locked", lockedByErr.LockedBy.ID))
				c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
				return
			}
			log.Warn(ctx, "deleteMilestone: Delete failed", log.Any("op", op), log.Strings("req", data.IDs))
			lable := AssessMsgLockedMilestone
			if len(data.IDs) == 1 {
				lable = AssessErrorMsgLocked
			}
			c.JSON(http.StatusConflict, LD(lable, user))
			return
		}
		log.Error(ctx, "deleteMilestone: Delete failed", log.Any("op", op), log.Strings("req", data.IDs))
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
// @Param status query string false "search by publish_status" Enums(draft, pending, published, rejected)
// @Param author_id query string false "search by author"
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at, updated_at, -updated_at)
// @Success 200 {object} model.MilestoneSearchResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
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

	if condition.OrganizationID == "" {
		condition.OrganizationID = op.OrgID
	}

	var hasPerm bool
	perm := external.ViewUnPublishedMilestone
	if condition.Status == entity.OutcomeStatusPublished {
		perm = external.ViewUnPublishedMilestone
	}

	hasPerm, err = external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, perm)
	if err != nil {
		log.Error(ctx, "searchMilestone: HasOrganizationPermission failed", log.Err(err), log.String("perm", string(perm)), log.Any("op", op))
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
		views, err := model.FromMilestones(ctx, op, milestones)
		if err != nil {
			log.Error(ctx, "searchMilestone: Search failed", log.Any("op", op), log.Any("req", condition))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, model.MilestoneSearchResponse{
			Total:      total,
			Milestones: views,
		})
	default:
		log.Error(ctx, "searchMilestone: Search failed", log.Any("op", op), log.Any("req", condition))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID searchPrivateMilestone
// @Summary search private milestone
// @Tags milestone
// @Description search private milestone
// @Accept json
// @Produce json
// @Param search_key query string false "search by search_key"
// @Param name query string false "search by name"
// @Param description query string false "search by description"
// @Param shortcode query string false "search by shortcode"
// @Param status query string false "search by publish_status" Enums(draft, pending, published, rejected)
// @Param author_id query string false "search by author"
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at, updated_at, -updated_at)
// @Success 200 {object} model.MilestoneSearchResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /private_milestones [get]
// search private milestone as an author user
func (s *Server) searchPrivateMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var condition entity.MilestoneCondition
	err := c.ShouldBindQuery(&condition)
	if err != nil {
		log.Warn(ctx, "searchPrivateMilestone: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	if condition.OrganizationID == "" {
		condition.OrganizationID = op.OrgID
	}

	if condition.Status == "" {
		condition.Status = string(entity.MilestoneStatusPublished)
	}
	// only query private milestone
	condition.AuthorID = op.UserID
	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.ViewMyUnpublishedMilestone,
		external.ViewMyPendingMilestone,
	})
	if err != nil {
		log.Error(ctx, "searchPrivateMilestone: HasOrganizationPermissions failed",
			log.Any("op", op),
			log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	// check search private permissions
	allowSearchPrivate := false
	if (condition.Status == string(entity.MilestoneStatusDraft) ||
		condition.Status == string(entity.MilestoneStatusRejected) ||
		condition.Status == string(entity.MilestoneStatusPending)) &&
		hasPerm[external.ViewMyUnpublishedMilestone] {
		allowSearchPrivate = true
	}

	if condition.Status == string(entity.MilestoneStatusPending) && hasPerm[external.ViewMyPendingMilestone] {
		allowSearchPrivate = true
	}

	if !allowSearchPrivate {
		log.Warn(ctx, "searchPrivateMilestone: no permission",
			log.Any("op", op),
			log.Any("condition", condition),
			log.Any("hasPerm", hasPerm))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}

	total, milestones, err := model.GetMilestoneModel().Search(ctx, op, &condition)
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
		views, err := model.FromMilestones(ctx, op, milestones)
		if err != nil {
			log.Error(ctx, "searchPrivateMilestone: search failed",
				log.Any("op", op),
				log.Any("req", condition))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, model.MilestoneSearchResponse{
			Total:      total,
			Milestones: views,
		})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID searchPendingMilestone
// @Summary search pending milestone
// @Tags milestone
// @Description search pending milestone
// @Accept json
// @Produce json
// @Param search_key query string false "search by search_key"
// @Param name query string false "search by name"
// @Param description query string false "search by description"
// @Param shortcode query string false "search by shortcode"
// @Param status query string false "search by publish_status" Enums(draft, pending, published, rejected)
// @Param author_id query string false "search by author"
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param order_by query string false "order by" Enums(name, -name, created_at, -created_at, updated_at, -updated_at)
// @Success 200 {object} model.MilestoneSearchResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /pending_milestones [get]
// search pending milestone as an admin user
func (s *Server) searchPendingMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var condition entity.MilestoneCondition
	err := c.ShouldBindQuery(&condition)
	if err != nil {
		log.Warn(ctx, "searchPendingMilestone: ShouldBind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	if condition.OrganizationID == "" {
		condition.OrganizationID = op.OrgID
	}

	// only query pending milestone
	condition.Status = string(entity.MilestoneStatusPending)
	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.ViewPendingMilestone,
		external.ViewMyPendingMilestone})
	if err != nil {
		log.Error(ctx, "searchPendingMilestone: HasOrganizationPermissions failed",
			log.Any("op", op),
			log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	// check search pending permissions
	allowSearchPending := false
	if condition.AuthorID == op.UserID && hasPerm[external.ViewMyPendingMilestone] {
		allowSearchPending = true
	}

	if hasPerm[external.ViewPendingMilestone] {
		allowSearchPending = true
	}

	if !allowSearchPending {
		log.Warn(ctx, "searchPendingMilestone: no permission",
			log.Any("op", op),
			log.Any("condition", condition),
			log.Any("hasPerm", hasPerm))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}

	total, milestones, err := model.GetMilestoneModel().Search(ctx, op, &condition)
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
		views, err := model.FromMilestones(ctx, op, milestones)
		if err != nil {
			log.Error(ctx, "searchPendingMilestone: search failed",
				log.Any("op", op),
				log.Any("req", condition))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		c.JSON(http.StatusOK, model.MilestoneSearchResponse{
			Total:      total,
			Milestones: views,
		})
	default:
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
// @Success 200 {string} model.MilestoneView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 409 {object} ConflictResponse
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
	switch err {
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrOutOfDate:
		log.Warn(ctx, "occupyMilestone: Update failed", log.Any("op", op))
		c.JSON(http.StatusForbidden, L(AssessMsgUnLockedMilestone))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, model.MilestoneView{MilestoneID: milestone.ID})
	default:
		lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
		if ok {
			user, err := external.GetUserServiceProvider().Get(ctx, op, lockedByErr.LockedBy.ID)
			if err != nil {
				log.Error(ctx, "occupyMilestone: Delete failed",
					log.Err(err),
					log.Any("op", op),
					log.String("req", milestoneID),
					log.String("locked", lockedByErr.LockedBy.ID))
				c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
				return
			}
			log.Warn(ctx, "occupyMilestone: Occupy failed", log.Any("op", op), log.String("req", milestoneID))
			c.JSON(http.StatusConflict, LD(AssessErrorMsgLocked, user))
			return
		}
		log.Error(ctx, "occupyMilestone: Occupy failed", log.Any("op", op), log.String("req", milestoneID))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @ID publishMilestonesBulk
// @Summary publish bulk milestone
// @Tags milestone
// @Description submit publish milestones
// @Accept json
// @Produce json
// @Param id_list body model.MilestoneList true "milestone id list"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /bulk_publish/milestones [put]
func (s *Server) bulkPublishMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data model.MilestoneList
	err := c.ShouldBindJSON(&data)
	if err != nil || len(data.IDs) == 0 {
		log.Warn(ctx, "bulkPublishMilestone: ShouldBind failed", log.Any("req", data), log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.CreateMilestone)
	if err != nil {
		log.Warn(ctx, "bulkPublishMilestone: HasOrganizationPermission failed",
			log.Any("op", op),
			log.Any("permissionName", external.CreateMilestone),
			log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "bulkPublishMilestone: no permission",
			log.Any("op", op),
			log.Any("permissionName", external.CreateMilestone))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}

	err = model.GetMilestoneModel().BulkPublish(ctx, op, utils.SliceDeduplication(data.IDs))
	switch err {
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
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

// @ID bulkApproveMilestone
// @Summary bulk approve milestone
// @Tags milestone
// @Description bulk approve milestone
// @Accept json
// @Produce json
// @Param id_list body model.MilestoneList true "milestone id list"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /bulk_approve/milestones [put]
func (s *Server) bulkApproveMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data model.MilestoneList
	err := c.ShouldBindJSON(&data)
	if err != nil || len(data.IDs) == 0 {
		log.Warn(ctx, "bulkApproveMilestone: ShouldBind failed", log.Any("req", data), log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ApprovePendingMilestone)
	if err != nil {
		log.Error(ctx, "bulkApproveMilestone: HasOrganizationPermission failed", log.Strings("ids", data.IDs), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "bulkApproveMilestone: no permission",
			log.Any("op", op), log.Strings("ids", data.IDs),
			log.String("perm", string(external.ApprovePendingMilestone)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	err = model.GetMilestoneModel().BulkApprove(ctx, op, utils.SliceDeduplication(data.IDs))
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

// @ID bulkRejectMilestone
// @Summary bulk reject milestone
// @Tags milestone
// @Description bulk reject milestone
// @Accept json
// @Produce json
// @Param bulk_reject_list body model.MilestoneBulkRejectRequest true "milestone id list"
// @Success 200 {string} string "ok"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /bulk_reject/milestones [put]
func (s *Server) bulkRejectMilestone(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data model.MilestoneBulkRejectRequest
	err := c.ShouldBindJSON(&data)
	if err != nil || len(data.MilestoneIDs) == 0 {
		log.Warn(ctx, "bulkRejectMilestone: ShouldBind failed", log.Any("req", data), log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPerm, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.RejectPendingMilestone)
	if err != nil {
		log.Error(ctx, "bulkRejectMilestone: HasOrganizationPermission failed",
			log.Strings("ids", data.MilestoneIDs),
			log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPerm {
		log.Warn(ctx, "bulkRejectMilestone: no permission",
			log.Any("op", op), log.Strings("ids", data.MilestoneIDs),
			log.String("perm", string(external.RejectPendingMilestone)))
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
		return
	}
	err = model.GetMilestoneModel().BulkReject(ctx, op, utils.SliceDeduplication(data.MilestoneIDs), data.RejectReason)
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
