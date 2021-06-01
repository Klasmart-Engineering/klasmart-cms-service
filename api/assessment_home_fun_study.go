package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
)

// @Summary list home fun studies
// @Description list home fun studies
// @Tags homeFunStudies
// @ID listHomeFunStudies
// @Accept json
// @Produce json
// @Param query query string false "fuzzy query teacher name and student name"
// @Param status query string false "query status" enums(all,in_progress,complete)
// @Param order_by query string false "list order by" enums(latest_feedback_at,-latest_feedback_at,complete_at,-complete_at) default(-latest_feedback_at)
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Success 200 {object} entity.ListHomeFunStudiesResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /home_fun_studies [get]
func (s *Server) listHomeFunStudies(c *gin.Context) {
	ctx := c.Request.Context()

	args := entity.ListHomeFunStudiesArgs{}

	if status := c.Query("status"); status != "" && status != constant.ListOptionAll {
		args.Status = entity.NullAssessmentStatus{
			Value: entity.AssessmentStatus(status),
			Valid: true,
		}
	}

	args.Query = c.Query("query")

	if orderBy := entity.ListHomeFunStudiesOrderBy(c.Query("order_by")); orderBy.Valid() {
		args.OrderBy = entity.NullListHomeFunStudiesOrderBy{
			Value: orderBy,
			Valid: true,
		}
	}

	args.Pager = utils.GetDboPager(c.Query("page"), c.Query("page_size"))

	result, err := model.GetHomeFunStudyModel().List(ctx, s.getOperator(c), args)
	if err != nil {
		log.Error(ctx, "listHomeFunStudies: model.GetHomeFunStudyModel().List",
			log.Err(err),
			log.Any("args", args),
		)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get home fun study
// @Description get home fun study detail
// @Tags homeFunStudies
// @ID getHomeFunStudy
// @Accept json
// @Produce json
// @Param id path string true "home fun study id"
// @Success 200 {object} entity.GetHomeFunStudyResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /home_fun_studies/{id} [get]
func (s *Server) getHomeFunStudy(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		log.Info(ctx, "getHomeFunStudy: require id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	result, err := model.GetHomeFunStudyModel().GetDetail(ctx, s.getOperator(c), id)
	if err != nil {
		log.Info(ctx, "model.GetHomeFunStudyModel().Get: get failed",
			log.Err(err),
			log.String("id", id),
		)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrRecordNotFound, sql.ErrNoRows:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary assess home fun study
// @Description assess home fun study
// @Tags homeFunStudies
// @ID assessHomeFunStudy
// @Accept json
// @Produce json
// @Param id path string true "home fun study id"
// @Param assess_home_fun_study_args body entity.AssessHomeFunStudyArgs true "assess home fun study args, body id don't need"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /home_fun_studies/{id}/assess [put]
func (s *Server) assessHomeFunStudy(c *gin.Context) {
	ctx := c.Request.Context()

	args := entity.AssessHomeFunStudyArgs{}
	if err := c.ShouldBind(&args); err != nil {
		log.Error(ctx, "assessHomeFunStudy: c.ShouldBind: bind failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if id := c.Param("id"); id == "" {
		log.Error(ctx, "assessHomeFunStudy: require id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	} else {
		args.ID = id
	}

	err := model.GetHomeFunStudyModel().Assess(ctx, dbo.MustGetDB(ctx), s.getOperator(c), args)
	if err != nil {
		log.Info(ctx, "assessHomeFunStudy: model.GetHomeFunStudyModel().Assess: assess failed",
			log.Err(err),
			log.Any("args", args),
		)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrRecordNotFound, sql.ErrNoRows:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrHomeFunStudyHasNewFeedback:
		c.JSON(http.StatusInternalServerError, L(AssessMsgNewVersion))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
