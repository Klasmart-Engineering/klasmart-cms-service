package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
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
	operator := s.getOperator(ctx)
	requestContext := ctx.Request.Context()
	lessonPlanID := ctx.Query("lesson_plan_id")
	result, err := model.GetReportModel().ListStudentReport(requestContext, dbo.MustGetDB(requestContext), lessonPlanID, operator)
	if err != nil {
		switch err {
		case constant.ErrInvalidArgs:
			ctx.JSON(http.StatusBadRequest, L(Unknown))
		default:
			ctx.JSON(http.StatusInternalServerError, L(Unknown))
		}
		return
	}
	ctx.JSON(http.StatusOK, result)
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
	operator := s.getOperator(ctx)
	requestContext := ctx.Request.Context()
	studentID := ctx.Param("id")
	lessonPlanID := ctx.Query("lesson_plan_id")
	result, err := model.GetReportModel().GetStudentReportDetail(ctx, dbo.MustGetDB(requestContext), studentID, lessonPlanID, operator)
	if err != nil {
		switch err {
		case constant.ErrInvalidArgs:
			ctx.JSON(http.StatusBadRequest, L(Unknown))
		default:
			ctx.JSON(http.StatusInternalServerError, L(Unknown))
		}
		return
	}
	ctx.JSON(http.StatusOK, result)
}

// @Summary get lessonPlans by teacher and class
// @Description get lessonPlans by teacher and class
// @Tags reports
// @ID getLessonPlans
// @Accept json
// @Produce json
// @Param teacher_id query string true "teacher id"
// @Param class_id query string true "class id"
// @Success 200 {array} entity.ReportLessonPlanInfo
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/lesson_plans [get]
func (s *Server) getLessonPlans(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, L(Unknown))
}
