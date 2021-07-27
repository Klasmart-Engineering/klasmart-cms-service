package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

// @Summary get assessments summary
// @Description get assessments summary
// @Tags assessments
// @ID getAssessmentsSummary
// @Accept json
// @Produce json
// @Param status query string false "status search"
// @Param teacher_name query string false "teacher name fuzzy search"
// @Success 200 {object} entity.AssessmentsSummary
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments_summary [get]
func (s *Server) getAssessmentsSummary(c *gin.Context) {
	ctx := c.Request.Context()
	args := entity.QueryAssessmentsSummaryArgs{}

	if status := c.Query("status"); status != "" && status != constant.ListOptionAll {
		args.Status = entity.NullAssessmentStatus{
			Value: entity.AssessmentStatus(status),
			Valid: true,
		}
	}
	teacherName := c.Query("teacher_name")
	if teacherName != "" {
		args.TeacherName = entity.NullString{
			String: teacherName,
			Valid:  true,
		}
	}
	operator := s.getOperator(c)
	if operator.OrgID == "" {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	result, err := model.GetAssessmentModel().Summary(ctx, dbo.MustGetDB(ctx), operator, args)
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

// @Summary get student assessments
// @Description get student assessments
// @Tags assessments
// @ID getStudentAssessments
// @Accept json
// @Produce json
// @Param type query string true "type search"
// @Param status query string false "status search"
// @Param order_by query string false "order by"
// @Param teacher_id query string false "teacher id search"
// @Param assessment_id query string false "assessment id search"
// @Param create_at_ge query string false "create_at greater search"
// @Param create_at_le query string false "create_at less search"
// @Param update_at_le query string false "update_at greater search"
// @Param update_at_le query string false "update_at less search"
// @Param complete_at_ge query string false "complete_at greater search"
// @Param complete_at_le query string false "complete_at less search"
// @Param page query string false "page search"
// @Param page_size query string false "page size search"
// @Success 200 {object} entity.SearchStudentAssessmentsResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /assessments_for_student [get]
func (s *Server) getStudentAssessments(c *gin.Context) {
	ctx := c.Request.Context()
	args := entity.QueryAssessmentsSummaryArgs{}

	operator := s.getOperator(c)
	if operator.OrgID == "" {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	//check permission
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.AssessmentViewTeacherFeedback670)
	if err != nil {
		log.Error(ctx, "hasPermission: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 670 failed",
			log.Err(err),
			log.Any("operator", operator),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		log.Warn(ctx, "No permission", log.Any("operator", operator))
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	conditions := new(entity.StudentQueryAssessmentConditions)
	err = c.ShouldBindQuery(conditions)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	conditions.OrgID = operator.OrgID
	conditions.StudentID = operator.UserID
	total, result, err := model.GetAssessmentModel().StudentQuery(ctx, operator, dbo.MustGetDB(ctx), conditions)
	switch err {
	case nil:
		c.JSON(http.StatusOK, &entity.SearchStudentAssessmentsResponse{
			List:  result,
			Total: total,
		})
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
	case model.ErrInvalidOrderByValue:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidAssessmentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		log.Error(ctx, "list assessments: list failed",
			log.Err(err),
			log.Any("args", args),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
}
