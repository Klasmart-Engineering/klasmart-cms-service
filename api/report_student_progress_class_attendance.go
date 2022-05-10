package api

import (
	"net/http"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/gin-gonic/gin"
)

// @Summary  getClassAttendance
// @Tags reports/studentProgress
// @ID getClassAttendance
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
	op := s.getOperator(c)
	var request entity.ClassAttendanceRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "getClassAttendance: ShouldBindQuery failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err = s.checkPermissionForReportStudentProgress(ctx, op, request.ClassID, request.StudentID)
	if err != nil {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}

	result, err := model.GetReportModel().ClassAttendanceStatistics(ctx, op, &request)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
