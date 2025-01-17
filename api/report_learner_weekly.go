package api

import (
	"net/http"

	"github.com/KL-Engineering/kidsloop-cms-service/external"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/gin-gonic/gin"
)

// @Summary get learner weekly report overview
// @Description get learner weekly report overview
// @Tags reports
// @ID getLearnerWeeklyReportOverview
// @Accept json
// @Produce json
// @Param time_range query string true "time_range"
// @Success 200 {object} entity.LearnerReportOverview
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learner_weekly_overview [get]
func (s *Server) getLearnerWeeklyReportOverview(c *gin.Context) {
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
		PermOrg:     external.ReportLearningSummmaryOrg.String(),
		PermSchool:  external.ReportLearningSummarySchool.String(),
		PermTeacher: external.ReportLearningSummaryTeacher.String(),
		PermStudent: external.ReportLearningSummaryStudent.String(),
	})
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, res)
}
