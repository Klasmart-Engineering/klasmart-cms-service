package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

// @Summary  getClassAttendance
// @Tags reports/studentProgress
// @ID getLearnOutcomeClassAttendance
// @Accept json
// @Produce json
// @Param request body entity.ClassAttendanceRequest true "request "
// @Success 200 {object} entity.ClassAttendanceResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_progress/class_attendance [post]
func (s *Server) getClassAttendance(c *gin.Context) {
	ctx := c.Request.Context()
	var request entity.ClassAttendanceRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "getClassAttendance: ShouldBindQuery failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetReportModel().ClassAttendanceStatistics(ctx, &request)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
