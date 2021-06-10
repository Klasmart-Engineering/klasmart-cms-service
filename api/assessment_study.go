package api

import (
	"context"
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

// @Summary list study assessments
// @Description list study assessments
// @Tags studyAssessments
// @ID listStudyAssessments
// @Accept json
// @Produce json
// @Param query query string false "query teacher name or class name"
// @Param query_type query string false "query type" enums(all,class_name,teacher_name) default(all)
// @Param status query string false "query status" enums(all,in_progress,complete) default(all)
// @Param order_by query string false "list order by" enums(create_at,-create_at,complete_time,-complete_time) default(-complete_time)
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Success 200 {object} entity.ListStudyAssessmentsResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /study_assessments [get]
func (s *Server) listStudyAssessments(c *gin.Context) {
	ctx := c.Request.Context()

	args := entity.ListStudyAssessmentsArgs{
		ClassTypes: []entity.ScheduleClassType{entity.ScheduleClassTypeHomework},
		QueryType:  entity.ListStudyAssessmentsQueryTypeTeacherName,
	}
	args.Query = c.Query("query")

	if status := c.Query("status"); status != "" && status != constant.ListOptionAll {
		args.Status = entity.NullAssessmentStatus{
			Value: entity.AssessmentStatus(status),
			Valid: true,
		}
	}
	if orderBy := c.Query("order_by"); orderBy != "" {
		args.OrderBy = entity.NullAssessmentsOrderBy{
			Value: entity.AssessmentOrderBy(orderBy),
			Valid: true,
		}
	} else {
		args.OrderBy = entity.NullAssessmentsOrderBy{
			Value: entity.AssessmentOrderByCreateAtDesc,
			Valid: true,
		}
	}
	args.Pager = utils.GetDboPager(c.Query("page"), c.Query("page_size"))

	result, err := model.GetStudyAssessmentModel().List(ctx, s.getOperator(c), dbo.MustGetDB(ctx), &args)
	if err != nil {
		log.Error(ctx, "list study assessments: call model list failed",
			log.Err(err),
			log.Any("args", args),
		)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get study assessment detail
// @Description get study assessment detail
// @Tags studyAssessments
// @ID getStudyAssessmentDetail
// @Accept json
// @Produce json
// @Param id path string true "study assessment id"
// @Success 200 {object} entity.AssessmentDetail
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /study_assessments/{id} [get]
func (s *Server) getStudyAssessmentDetail(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		log.Error(ctx, "get study assessment detail: require id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	result, err := model.GetStudyAssessmentModel().GetDetail(ctx, s.getOperator(c), dbo.MustGetDB(ctx), id)
	if err != nil {
		log.Info(ctx, "get study assessment detail: call model failed",
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

// @Summary
// @Description update study assessment
// @Tags studyAssessments
// @ID updateStudyAssessment
// @Accept json
// @Produce json
// @Param id path string true "study assessment id"
// @Param update_assessment_args body entity.UpdateAssessmentArgs true "update assessment args"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /study_assessments/{id} [put]
func (s *Server) updateStudyAssessment(c *gin.Context) {
	ctx := c.Request.Context()

	args := entity.UpdateAssessmentArgs{}
	if err := c.ShouldBind(&args); err != nil {
		log.Error(ctx, "update study assessment: bind body json failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if id := c.Param("id"); id == "" {
		log.Error(ctx, "update study assessment: require id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	} else {
		args.ID = id
	}

	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		return model.GetStudyAssessmentModel().Update(ctx, s.getOperator(c), dbo.MustGetDB(ctx), &args)
	})
	if err != nil {
		log.Error(ctx, "update study assessment: call model failed",
			log.Err(err),
			log.Any("args", args),
		)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
