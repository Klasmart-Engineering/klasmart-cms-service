package api

import "github.com/gin-gonic/gin"

// @Summary  getAssignmentsCompletion
// @Tags reports/studentProgress
// @ID getAssignmentsCompletion
// @Accept json
// @Produce json
// @Param request body entity.AssignmentRequest true "request "
// @Success 200 {object} entity.AssignmentResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_progress/assignment_completion [post]
func (s *Server) getAssignmentsCompletion(c *gin.Context) {
	panic("implement me")
}
