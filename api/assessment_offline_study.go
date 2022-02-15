package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary query user offline study(homeFun study)
// @Description query user offline study
// @Tags assessments
// @ID queryUserOfflineStudy
// @Accept json
// @Produce json
// @Param status query string false "status search"
// @Param query_key query string false "query key fuzzy search"
// @Param query_type query string false "query type" enums(TeacherName)
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Param order_by query string false "query order by" enums(complete_at,-complete_at,submit_at,-submit_at) default(-submit_at)
// @Success 200 {object} v2.OfflineStudyUserPageReply
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /user_offline_study [get]
func (s *Server) queryUserOfflineStudy(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	req := new(v2.AssessmentQueryReq)
	if err := c.ShouldBind(req); err != nil {
		return
	}

	if req.PageSize <= 0 || req.PageIndex <= 0 {
		req.PageIndex = constant.DefaultPageIndex
		req.PageSize = constant.DefaultPageSize
	}

	log.Debug(ctx, "queryUserOfflineStudy request", log.Any("req", req), log.Any("op", op))

	result, err := model.GetAssessmentOfflineStudyModel().Page(ctx, op, req)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary get user offline study by id
// @Description get user offline study by id
// @Tags assessments
// @ID getUserOfflineStudyByID
// @Accept json
// @Produce json
// @Param id path string true "user offline study id"
// @Success 200 {object} v2.GetOfflineStudyUserResultDetailReply
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /user_offline_study/{id} [get]
func (s *Server) getUserOfflineStudyByID(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	id := c.Param("id")

	result, err := model.GetAssessmentOfflineStudyModel().GetByID(ctx, op, id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary update user offline study
// @Description update user offline study
// @Tags assessments
// @ID updateUserOfflineStudy
// @Accept json
// @Produce json
// @Param id path string true "user offline study id"
// @Param req body v2.OfflineStudyUserResultUpdateReq true "req data"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /user_offline_study/{id} [put]
func (s *Server) updateUserOfflineStudy(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	req := new(v2.OfflineStudyUserResultUpdateReq)
	if err := c.ShouldBind(req); err != nil {
		log.Error(ctx, "updateUserOfflineStudy: bind failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	req.ID = c.Param("id")

	log.Debug(ctx, "queryUserOfflineStudy request", log.Any("req", req), log.Any("op", op))

	var err error
	if req.Action == v2.AssessmentActionDraft {
		err = model.GetAssessmentOfflineStudyModel().Draft(ctx, op, req)
	} else if req.Action == v2.AssessmentActionComplete {
		err = model.GetAssessmentOfflineStudyModel().Complete(ctx, op, req)
	} else {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	switch err {
	case nil:
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrRecordNotFound, sql.ErrNoRows:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrOfflineStudyHasNewFeedback:
		c.JSON(http.StatusBadRequest, L(AssessMsgNewVersion))
	default:
		s.defaultErrorHandler(c, err)
	}
}
