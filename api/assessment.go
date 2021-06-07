package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
