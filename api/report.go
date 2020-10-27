package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"strings"
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
// @Param status query string true "status" enums(all, all_achieved, not_achieved, not_attempted)
// @Param status sortBy string true "sort by" enums(descending, ascending)
// @Success 200 {object} entity.StudentsReport
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/students [get]
func (s *Server) listStudentsReport(ctx *gin.Context) {
	requestContext := ctx.Request.Context()
	operator := s.getOperator(ctx)
	cmd := entity.ListStudentsReportCommand{
		TeacherID:    ctx.Query("teacher_id"),
		ClassID:      ctx.Query("class_id"),
		LessonPlanID: ctx.Query("lesson_plan_id"),
		Status:       ctx.Query("status"),
		SortBy:       ctx.Query("sort_by"),
		Operator:     &operator,
	}
	result, err := model.GetReportModel().ListStudentsReport(requestContext, dbo.MustGetDB(requestContext), cmd)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		ctx.JSON(http.StatusBadRequest, L(Unknown))
	default:
		ctx.JSON(http.StatusInternalServerError, L(Unknown))
	}
}

// @Summary get student report
// @Description get student report
// @Tags reports
// @ID getStudentDetailReport
// @Accept json
// @Produce json
// @Param id path string true "student id"
// @Param teacher_id query string true "teacher_id"
// @Param class_id query string true "class_id"
// @Param lesson_plain_id query string true "lesson plain id"
// @Success 200 {object} entity.StudentDetailReport
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/students/{id} [get]
func (s *Server) getStudentDetailReport(ctx *gin.Context) {
	requestContext := ctx.Request.Context()
	operator := s.getOperator(ctx)
	cmd := entity.GetStudentDetailReportCommand{
		StudentID:    ctx.Param("id"),
		TeacherID:    ctx.Query("teacher_id"),
		ClassID:      ctx.Query("class_id"),
		LessonPlanID: ctx.Query("lesson_plan_id"),
		Operator:     &operator,
	}
	result, err := model.GetReportModel().GetStudentDetailReport(requestContext, dbo.MustGetDB(requestContext), cmd)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		ctx.JSON(http.StatusBadRequest, L(Unknown))
	case constant.ErrRecordNotFound, sql.ErrNoRows:
		ctx.JSON(http.StatusNotFound, L(Unknown))
	default:
		ctx.JSON(http.StatusInternalServerError, L(Unknown))
	}
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
	operator := s.getOperator(c)
	ctx := c.Request.Context()
	teacherID := c.Query("teacher_id")
	classID := c.Query("class_id")
	if len(strings.TrimSpace(teacherID)) == 0 || len(strings.TrimSpace(classID)) == 0 {
		log.Info(ctx, "teacherID and classID is require",
			log.String("teacherID", teacherID),
			log.String("classID", classID),
			log.Any("operator", operator),
		)
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	result, err := model.GetReportModel().GetReportLessPlanInfo(ctx, dbo.MustGetDB(ctx), teacherID, classID, operator)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(Unknown))
	}
}
