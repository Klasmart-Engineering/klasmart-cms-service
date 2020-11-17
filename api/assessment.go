package api

import (
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

// @Summary list assessments
// @Description list assessments
// @Tags assessments
// @ID listAssessment
// @Accept json
// @Produce json
// @Param status query string false "status search"
// @Param teacher_name query string false "teacher name fuzzy search"
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Param order_by query string false "list order by" enums(class_end_time,-class_end_time,complete_time,-complete_time) default(-class_end_time)
// @Success 200 {object} entity.ListAssessmentsResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments [get]
func (s *Server) listAssessments(c *gin.Context) {
	ctx := c.Request.Context()

	cmd := entity.ListAssessmentsQuery{}
	{
		status := c.Query("status")
		if status != "" {
			status := entity.ListAssessmentsStatus(status)
			if !status.Valid() {
				log.Info(ctx, "list assessments: invalid list assessments status",
					log.String("status", string(status)),
				)
				c.JSON(http.StatusBadRequest, L(GeneralUnknown))
				return
			}
			if status != entity.ListAssessmentsStatusAll {
				temp := status.AssessmentStatus()
				cmd.Status = &temp
			}
		}

		teacherName := c.Query("teacher_name")
		if teacherName != "" {
			cmd.TeacherName = &teacherName
		}

		orderBy := c.Query("order_by")
		if orderBy != "" {
			orderBy := entity.ListAssessmentsOrderBy(orderBy)
			if !orderBy.Valid() {
				log.Info(ctx, "list assessments: invalid order by",
					log.String("status", string(status)),
				)
				c.JSON(http.StatusBadRequest, L(GeneralUnknown))
				return
			}
			cmd.OrderBy = &orderBy
		} else {
			orderBy := entity.ListAssessmentsOrderByClassEndTimeDesc
			cmd.OrderBy = &orderBy
		}

		pager := utils.GetDboPager(c.Query("page"), c.Query("page_size"))
		cmd.Page, cmd.PageSize = pager.Page, pager.PageSize
	}

	result, err := model.GetAssessmentModel().List(ctx, dbo.MustGetDB(ctx), GetOperator(c), cmd)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		log.Error(ctx, "list assessments: list failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
}

// @Summary add assessments
// @Description add assessments
// @Tags assessments
// @ID addAssessment
// @Accept json
// @Produce json
// @Param assessment body entity.AddAssessmentCommand true "add assessment command"
// @Success 200 {object} entity.AddAssessmentResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments [post]
func (s *Server) addAssessment(c *gin.Context) {
	ctx := c.Request.Context()

	cmd := entity.AddAssessmentCommand{}
	if err := c.ShouldBind(&cmd); err != nil {
		log.Info(ctx, "add assessment: bind failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	newID, err := model.GetAssessmentModel().Add(ctx, GetOperator(c), cmd)
	switch err {
	case nil:
		c.JSON(http.StatusOK, entity.AddAssessmentResult{ID: newID})
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		log.Error(ctx, "add assessment: add failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get assessment detail
// @Description get assessment detail
// @Tags assessments
// @ID getAssessment
// @Accept json
// @Produce json
// @Param id path string true "assessment id"
// @Success 200 {object} entity.AssessmentDetailView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments/{id} [get]
func (s *Server) getAssessmentDetail(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		log.Info(ctx, "get assessment detail: require id")
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	item, err := model.GetAssessmentModel().Detail(ctx, dbo.MustGetDB(ctx), GetOperator(c), id)
	switch err {
	case nil:
		c.JSON(http.StatusOK, item)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case constant.ErrRecordNotFound, sql.ErrNoRows:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	default:
		log.Info(ctx, "get assessment detail: get detail failed",
			log.Err(err),
			log.String("id", id),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary update assessment
// @Description update assessment
// @Tags assessments
// @ID updateAssessment
// @Accept json
// @Produce json
// @Param id path string true "assessment id"
// @Param update_assessment_command body entity.UpdateAssessmentCommand true "update assessment assessment command"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments/{id} [put]
func (s *Server) updateAssessment(c *gin.Context) {
	ctx := c.Request.Context()

	cmd := entity.UpdateAssessmentCommand{}
	{
		if err := c.ShouldBind(&cmd); err != nil {
			log.Info(ctx, "update assessment: bind failed",
				log.Err(err),
			)
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return
		}

		id := c.Param("id")
		if id == "" {
			log.Info(ctx, "update assessment: require id")
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return
		}
		cmd.ID = id
	}

	{
		if cmd.AttendanceIDs != nil && len(*cmd.AttendanceIDs) == 0 {
			c.JSON(http.StatusBadRequest, L(AssessMsgOneStudent))
			return
		}
	}

	err := model.GetAssessmentModel().Update(ctx, GetOperator(c), cmd)
	switch err {
	case nil:
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		log.Info(ctx, "update assessment: update failed")
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
