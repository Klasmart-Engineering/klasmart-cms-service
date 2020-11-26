package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary list student report
// @Description list student report
// @Tags reports
// @ID listStudentsReport
// @Accept json
// @Produce json
// @Param teacher_id query string true "teacher_id"
// @Param class_id query string true "class_id"
// @Param lesson_plan_id query string true "lesson plan id"
// @Param status query string false "status" enums(all, achieved, not_achieved, not_attempted) default(all)
// @Param sort_by query string false "sort by" enums(desc, asc) default(desc)
// @Success 200 {object} entity.StudentsReport
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/students [get]
func (s *Server) listStudentsReport(ctx *gin.Context) {
	requestContext := ctx.Request.Context()
	operator := s.getOperator(ctx)
	cmd := entity.ListStudentsReportCommand{
		TeacherID:    ctx.Query("teacher_id"),
		ClassID:      ctx.Query("class_id"),
		LessonPlanID: ctx.Query("lesson_plan_id"),
		Status:       entity.ReportOutcomeStatusOption(ctx.DefaultQuery("status", string(entity.ReportOutcomeStatusOptionAll))),
		SortBy:       entity.ReportSortBy(ctx.DefaultQuery("sort_by", string(entity.ReportSortByDesc))),
		Operator:     operator,
	}
	result, err := model.GetReportModel().ListStudentsReport(requestContext, dbo.MustGetDB(requestContext), operator, cmd)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		ctx.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		ctx.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		ctx.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get student report
// @Description get student report
// @Tags reports
// @ID getStudentReport
// @Accept json
// @Produce json
// @Param id path string true "student id"
// @Param teacher_id query string true "teacher_id"
// @Param class_id query string true "class_id"
// @Param lesson_plan_id query string true "lesson plan id"
// @Success 200 {object} entity.StudentReport
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/students/{id} [get]
func (s *Server) getStudentReport(ctx *gin.Context) {
	requestContext := ctx.Request.Context()
	operator := s.getOperator(ctx)
	cmd := entity.GetStudentReportCommand{
		StudentID:    ctx.Param("id"),
		TeacherID:    ctx.Query("teacher_id"),
		ClassID:      ctx.Query("class_id"),
		LessonPlanID: ctx.Query("lesson_plan_id"),
		Operator:     operator,
	}
	result, err := model.GetReportModel().GetStudentReport(requestContext, dbo.MustGetDB(requestContext), operator, cmd)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		ctx.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrRecordNotFound, sql.ErrNoRows:
		ctx.JSON(http.StatusNotFound, L(GeneralUnknown))
	case constant.ErrForbidden:
		ctx.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		ctx.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get teacher report
// @Description get teacher report
// @Tags reports
// @ID getTeacherReport
// @Accept json
// @Produce json
// @Param id path string true "teacher id"
// @Success 200 {object} entity.TeacherReport
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/teachers/{id} [get]
func (s *Server) getTeacherReport(ctx *gin.Context) {
	requestContext := ctx.Request.Context()
	operator := s.getOperator(ctx)
	teacherID := ctx.Param("id")
	result, err := model.GetReportModel().GetTeacherReport(requestContext, dbo.MustGetDB(requestContext), operator, teacherID)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		ctx.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrRecordNotFound, sql.ErrNoRows:
		ctx.JSON(http.StatusNotFound, L(GeneralUnknown))
	case constant.ErrForbidden:
		ctx.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		ctx.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
