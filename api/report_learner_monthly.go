package api

import (
	"github.com/gin-gonic/gin"
)

// @Summary get learner monthly report overview
// @Description get learner monthly report overview
// @Tags reports
// @ID getLearnerMonthlyReportOverview
// @Accept json
// @Produce json
// @Param time_range query string true "time_range"
// @Success 200 {object} entity.LearnerMonthlyReportOverview
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/learner_monthly_overview [get]
func (s *Server) getLearnerMonthlyReportOverview(c *gin.Context) {
	//ctx := c.Request.Context()
	//var err error
	//defer func() {
	//	if err == nil {
	//		return
	//	}
	//	switch err {
	//	case constant.ErrInvalidArgs:
	//		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//	default:
	//		s.defaultErrorHandler(c, err)
	//	}
	//}()
	//op := s.getOperator(c)
	//tr := entity.TimeRange(c.Query("time_range"))
	//res, err := model.GetReportModel().GetLearnerWeeklyReportOverview(ctx, op, tr)
	//if err != nil {
	//	return
	//}
	//c.JSON(http.StatusOK, res)
}
