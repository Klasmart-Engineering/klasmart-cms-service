package api

import (
	"net/http"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/gin-gonic/gin"
)

// @Summary get learner monthly report overview
// @Description get learner monthly report overview
// @Tags reports
// @ID getLearnerMonthlyReportOverview
// @Accept json
// @Produce json
// @Param time_range query string true "time_range"
// @Success 200 {object} entity.LearnerReportOverview
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learner_monthly_overview [get]
func (s *Server) getLearnerMonthlyReportOverview(c *gin.Context) {
	ctx := c.Request.Context()
	var err error
	defer func() {
		if err == nil {
			return
		}
		switch err {
		case constant.ErrInvalidArgs:
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		default:
			s.defaultErrorHandler(c, err)
		}
	}()
	op := s.getOperator(c)
	tr := entity.TimeRange(c.Query("time_range"))
	res, err := model.GetReportModel().GetLearnerReportOverview(ctx, op, &entity.LearnerReportOverviewCondition{
		TimeRange:   tr,
		PermOrg:     external.ReportStudentProgressReportOrganization.String(),
		PermSchool:  external.ReportStudentProgressReportSchool.String(),
		PermTeacher: external.ReportStudentProgressReportTeacher.String(),
		PermStudent: external.ReportStudentProgressReportStudent.String(),
	})
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, res)
}
