package api

import (
	"database/sql"
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

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
// @Param class_type query string false "class type"
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

	args := entity.QueryAssessmentsArgs{}
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
				args.Status = entity.NullAssessmentStatus{
					Value: status.AssessmentStatus(),
					Valid: true,
				}
			}
		}

		teacherName := c.Query("teacher_name")
		if teacherName != "" {
			args.TeacherName = entity.NullString{
				String: teacherName,
				Valid:  true,
			}
		}

		orderBy := c.Query("order_by")
		if orderBy != "" {
			orderBy := entity.AssessmentOrderBy(orderBy)
			if !orderBy.Valid() {
				log.Info(ctx, "list assessments: invalid order by",
					log.String("status", string(status)),
				)
				c.JSON(http.StatusBadRequest, L(GeneralUnknown))
				return
			}
			args.OrderBy = entity.NullAssessmentsOrderBy{
				Value: orderBy,
				Valid: true,
			}
		} else {
			args.OrderBy = entity.NullAssessmentsOrderBy{
				Value: entity.ListAssessmentsOrderByClassEndTimeDesc,
				Valid: true,
			}
		}

		args.Pager = utils.GetDboPager(c.Query("page"), c.Query("page_size"))

		classType := c.Query("class_type")
		if classType != "" {
			args.ClassType = entity.NullScheduleClassType{
				Value: entity.ScheduleClassType(classType),
				Valid: true,
			}
		}
	}

	result, err := model.GetOutcomeAssessmentModel().List(ctx, dbo.MustGetDB(ctx), s.getOperator(c), args)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		log.Error(ctx, "list assessments: list failed",
			log.Err(err),
			log.Any("cmd", args),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
}

// @Summary get assessments summary
// @Description get assessments summary
// @Tags assessments
// @ID getAssessmentsSummary
// @Accept json
// @Produce json
// @Param status query string false "status search"
// @Param teacher_name query string false "teacher name fuzzy search"
// @Param class_type query string false "class type"
// @Success 200 {object} entity.AssessmentsSummary
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments_summary [get]
func (s *Server) getAssessmentsSummary(c *gin.Context) {
	ctx := c.Request.Context()
	args := entity.QueryAssessmentsSummaryArgs{}
	status := c.Query("status")
	if status != "" {
		status := entity.ListAssessmentsStatus(status)
		if !status.Valid() {
			log.Info(ctx, "getAssessmentsSummary: invalid list assessments status",
				log.String("status", string(status)),
			)
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return
		}
		if status != entity.ListAssessmentsStatusAll {
			args.Status = entity.NullAssessmentStatus{
				Value: status.AssessmentStatus(),
				Valid: true,
			}
		}
	}
	teacherName := c.Query("teacher_name")
	if teacherName != "" {
		args.TeacherName = entity.NullString{
			String: teacherName,
			Valid:  true,
		}
	}
	classType := c.Query("class_type")
	if classType != "" {
		args.ClassType = entity.NullScheduleClassType{
			Value: entity.ScheduleClassType(classType),
			Valid: true,
		}
	}
	operator := s.getOperator(c)
	if operator.OrgID == "" {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	result, err := model.GetOutcomeAssessmentModel().Summary(ctx, dbo.MustGetDB(ctx), operator, args)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	default:
		log.Error(ctx, "list assessments: list failed",
			log.Err(err),
			log.Any("args", args),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
}

// @Summary get assessment detail
// @Description get assessment detail
// @Tags assessments
// @ID getAssessment
// @Accept json
// @Produce json
// @Param id path string true "assessment id"
// @Success 200 {object} entity.AssessmentDetail
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

	item, err := model.GetOutcomeAssessmentModel().GetDetail(ctx, dbo.MustGetDB(ctx), s.getOperator(c), id)
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
// @Param update_assessment_command body entity.UpdateAssessmentArgs true "update assessment assessment command"
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments/{id} [put]
func (s *Server) updateAssessment(c *gin.Context) {
	ctx := c.Request.Context()

	args := entity.UpdateAssessmentArgs{}
	if err := c.ShouldBind(&args); err != nil {
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
	args.ID = id

	if args.StudentIDs != nil && len(*args.StudentIDs) == 0 {
		c.JSON(http.StatusBadRequest, L(AssessMsgOneStudent))
		return
	}

	err := model.GetOutcomeAssessmentModel().Update(ctx, dbo.MustGetDB(ctx), s.getOperator(c), args)
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
