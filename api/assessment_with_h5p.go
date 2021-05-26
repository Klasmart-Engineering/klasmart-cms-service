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

// @Summary list h5p assessments
// @Description list h5p assessments
// @Tags h5pAssessments
// @ID listH5PAssessments
// @Accept json
// @Produce json
// @Param type query string false "h5p assessment type" enums(study_h5p)
// @Param query query string false "query teacher name or class name"
// @Param query_type query string false "query type" enums(all,class_name,teacher_name) default(all)
// @Param status query string false "query status" enums(all,in_progress,complete) default(all)
// @Param order_by query string false "list order by" enums(create_at,-create_at,complete_time,-complete_time) default(-complete_time)
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Success 200 {object} entity.ListH5PAssessmentsResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /h5p_assessments [get]
func (s *Server) listH5PAssessments(c *gin.Context) {
	ctx := c.Request.Context()

	args := entity.ListH5PAssessmentsArgs{
		Type:      entity.AssessmentTypeStudyH5P,
		QueryType: entity.ListH5PAssessmentsQueryTypeTeacherName,
	}
	args.Query = c.Query("query")
	//args.QueryType = entity.ListH5PAssessmentsQueryType(c.Query("query_type"))
	//if status := entity.AssessmentStatus(c.Query("status")); status.Valid() {
	//	args.Status = entity.NullAssessmentStatus{
	//		Value: status,
	//		Valid: true,
	//	}
	//}

	if status := c.Query("status"); entity.AssessmentStatus(status).Valid() {
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

	result, err := model.GetH5PAssessmentModel().List(ctx, s.getOperator(c), dbo.MustGetDB(ctx), args)
	if err != nil {
		log.Error(ctx, "list h5p assessments: call model list failed",
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

// @Summary get h5p assessment detail
// @Description get h5p assessment detail
// @Tags h5pAssessments
// @ID getH5PAssessmentDetail
// @Accept json
// @Produce json
// @Param id path string true "h5p assessment id"
// @Success 200 {object} entity.GetH5PAssessmentDetailResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /h5p_assessments/{id} [get]
func (s *Server) getH5PAssessmentDetail(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		log.Error(ctx, "get h5p assessment detail: require id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	result, err := model.GetH5PAssessmentModel().GetDetail(ctx, s.getOperator(c), dbo.MustGetDB(ctx), id)
	if err != nil {
		log.Info(ctx, "get h5p assessment detail: call model failed",
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
// @Description update h5p assessment
// @Tags h5pAssessments
// @ID updateH5PAssessment
// @Accept json
// @Produce json
// @Param id path string true "h5p assessment id"
// @Param update_h5p_assessment_args body entity.UpdateH5PAssessmentArgs true "update h5p assessment args"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /h5p_assessments/{id} [put]
func (s *Server) updateH5PAssessment(c *gin.Context) {
	ctx := c.Request.Context()

	args := entity.UpdateH5PAssessmentArgs{}
	if err := c.ShouldBind(&args); err != nil {
		log.Error(ctx, "update h5p assessment: bind body json failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if id := c.Param("id"); id == "" {
		log.Error(ctx, "update h5p assessment: require id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	} else {
		args.ID = id
	}

	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		return model.GetH5PAssessmentModel().Update(ctx, s.getOperator(c), dbo.MustGetDB(ctx), args)
	})
	if err != nil {
		log.Error(ctx, "update h5p assessment: call model failed",
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
