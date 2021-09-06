package api

import (
	"github.com/gin-gonic/gin"
)

// @Summary get Classes&Assignments Report
// @Description get Classes&Assignments overview
// @Tags reports/studentUsage
// @ID getClassesAssignmentsOverview
// @Accept json
// @Produce json
// @Param class_ids query []string false "class id list"
// @Param durations query []string false "time durations, for example: [startTime1-endTime1, startTime2-endTime2]"
// @Success 200 {object} entity.ClassesAssignmentsOverView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/classes_assignments_overview [get]
func (s *Server) getClassesAssignmentsOverview(c *gin.Context) {
	panic("unimplemented")
}

// @Summary get Classes&Assignments Report
// @Description get Classes&Assignments Report
// @Tags reports/studentUsage
// @ID getClassesAssignments
// @Accept json
// @Produce json
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param class_ids query []string false "class id list"
// @Param durations query []string false "time durations, for example: [startTime1-endTime1, startTime2-endTime2]"
// @Success 200 {object} []entity.ClassesAssignmentDetailView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/classes_assignments [get]
func (s *Server) getClassesAssignments(c *gin.Context) {
	panic("unimplemented")
}

// @Summary get Classes&Assignments Report
// @Description get Classes&Assignments unattended
// @Tags reports/studentUsage
// @ID getClassesAssignmentsUnattended
// @Accept json
// @Produce json
// @Param page query integer false "page"
// @Param page_size query integer false "page size"
// @Param durations query []string false "time durations, for example: [startTime1-endTime1, startTime2-endTime2]"
// @Success 200 {object} []entity.ClassesAssignmentsUnattendedStudentsView
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_usage/classes_assignments/{class_id}/unattended [get]
func (s *Server) getClassesAssignmentsUnattended(c *gin.Context) {
	panic("unimplemented")
}
