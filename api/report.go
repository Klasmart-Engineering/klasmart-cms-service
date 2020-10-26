package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary list student report
// @Description list student report
// @Tags reports
// @ID listStudentReport
// @Accept json
// @Produce json
// @Param lesson_plain_id query string true "lesson plain id"
// @Success 200 {object} entity.StudentReportList
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/students [get]
func (s *Server) listStudentReport(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, "not implemented")
}

// @Summary get student report
// @Description get student report
// @Tags reports
// @ID getStudentReportDetail
// @Accept json
// @Produce json
// @Param id path string true "student id"
// @Param lesson_plan_id query string true "lesson plan id"
// @Success 200 {object} entity.StudentReportDetail
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/students/{id} [get]
func (s *Server) getStudentReportDetail(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, "not implemented")
}
