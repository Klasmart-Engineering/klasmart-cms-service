package api

import "github.com/gin-gonic/gin"

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

}
