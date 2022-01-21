package api

import (
	"database/sql"
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary list student report overview
// @Description list student report overview
// @Tags reports
// @ID listStudentsAchievementOverviewReport
// @Accept json
// @Produce json
// @Param time_range query entity.TimeRange true "time_range"
// @Success 200 {object} entity.StudentsAchievementOverviewReportResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/students_achievement_overview [get]
func (s *Server) listStudentsAchievementOverviewReport(ctx *gin.Context) {

}

// @Summary list student report
// @Description list student report
// @Tags reports
// @ID listStudentsAchievementReport
// @Accept json
// @Produce json
// @Param teacher_id query string true "teacher_id"
// @Param class_id query string true "class_id"
// @Param lesson_plan_id query string true "lesson plan id"
// @Param status query string false "status" enums(all, achieved, not_achieved, not_attempted) default(all)
// @Param sort_by query string false "sort by" enums(desc, asc) default(desc)
// @Success 200 {object} entity.StudentsAchievementReportResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/students [get]
func (s *Server) listStudentsAchievementReport(ctx *gin.Context) {
	requestContext := ctx.Request.Context()
	operator := s.getOperator(ctx)
	cmd := entity.ListStudentsAchievementReportRequest{
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
// @ID getStudentAchievementReport
// @Accept json
// @Produce json
// @Param id path string true "student id"
// @Param teacher_id query string true "teacher_id"
// @Param class_id query string true "class_id"
// @Param lesson_plan_id query string true "lesson plan id"
// @Success 200 {object} entity.StudentAchievementReportResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/students/{id} [get]
func (s *Server) getStudentAchievementReport(ctx *gin.Context) {
	requestContext := ctx.Request.Context()
	operator := s.getOperator(ctx)
	cmd := entity.GetStudentAchievementReportRequest{
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

// @Summary get teachers report
// @Description get teacher sreport
// @Tags reports
// @ID getTeachersReport
// @Accept json
// @Produce json
// @Success 200 {object} entity.TeacherReport
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/teachers [get]
func (s *Server) getTeachersReport(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)
	var err error
	var result *entity.TeacherReport
	defer func() {
		switch err {
		case nil:
			c.JSON(http.StatusOK, result)
		case constant.ErrInvalidArgs:
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		case constant.ErrRecordNotFound, sql.ErrNoRows:
			c.JSON(http.StatusNotFound, L(GeneralUnknown))
		case constant.ErrForbidden:
			c.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
		default:
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		}
	}()
	teacherIDs, err := model.GetReportModel().GetTeacherIDsCanViewReports(ctx, operator, external.TeacherViewPermissionParams{
		ViewOrgOrSchoolReports: external.ReportLearningOutcomesInCategories616,
		ViewSchoolReports:      external.ReportSchoolsSkillsTaught641,
		ViewOrgReports:         external.ReportOrganizationsSkillsTaught640,
		ViewMyReports:          external.ReportMySkillsTaught642,
	})
	if err != nil {
		return
	}
	result, err = model.GetReportModel().GetTeacherReport(ctx, dbo.MustGetDB(ctx), operator, teacherIDs...)
	if err != nil {
		return
	}
	return
}

// @Summary list student performance report
// @Description list student performance report
// @Tags reports
// @ID listStudentsPerformanceReport
// @Accept json
// @Produce json
// @Param teacher_id query string true "teacher_id"
// @Param class_id query string true "class_id"
// @Param lesson_plan_id query string true "lesson plan id"
// @Success 200 {object} entity.ListStudentsPerformanceReportResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/performance/students [get]
func (s *Server) listStudentsPerformanceReport(ctx *gin.Context) {
	reqCtx := ctx.Request.Context()
	operator := s.getOperator(ctx)
	req := entity.ListStudentsPerformanceReportRequest{
		TeacherID:    ctx.Query("teacher_id"),
		ClassID:      ctx.Query("class_id"),
		LessonPlanID: ctx.Query("lesson_plan_id"),
		Operator:     operator,
	}
	result, err := model.GetReportModel().ListStudentsPerformanceReport(reqCtx, dbo.MustGetDB(reqCtx), operator, req)
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

// @Summary get student performance report
// @Description get student performance report
// @Tags reports
// @ID getStudentPerformanceReport
// @Accept json
// @Produce json
// @Param id path string true "student id"
// @Param teacher_id query string true "teacher_id"
// @Param class_id query string true "class_id"
// @Param lesson_plan_id query string true "lesson plan id"
// @Success 200 {object} entity.GetStudentPerformanceReportResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/performance/students/{id} [get]
func (s *Server) getStudentPerformanceReport(ctx *gin.Context) {
	reqCtx := ctx.Request.Context()
	operator := s.getOperator(ctx)
	req := entity.GetStudentPerformanceReportRequest{
		StudentID:    ctx.Param("id"),
		TeacherID:    ctx.Query("teacher_id"),
		ClassID:      ctx.Query("class_id"),
		LessonPlanID: ctx.Query("lesson_plan_id"),
		Operator:     operator,
	}
	result, err := model.GetReportModel().GetStudentPerformanceReport(reqCtx, dbo.MustGetDB(reqCtx), operator, req)
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

// @Summary list teaching load report
// @Description list teaching load report
// @Tags reports
// @ID listTeachingLoadReport
// @Accept json
// @Produce json
// @Param teaching_load body entity.ReportListTeachingLoadArgs true "query teaching load"
// @Success 200 {object} entity.ReportListTeachingLoadResult
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/teaching_loading [post]
func (s *Server) listTeachingLoadReport(c *gin.Context) {
	ctx := c.Request.Context()
	operator := s.getOperator(c)
	var args entity.ReportListTeachingLoadArgs
	if err := c.ShouldBindJSON(&args); err != nil {
		log.Error(ctx, "listTeachingLoadReport: c.ShouldBindJson failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	result, err := model.GetReportTeachingLoadModel().ListTeachingLoadReport(ctx, dbo.MustGetDB(ctx), operator, &args)
	switch err {
	case nil:
		if result == nil {
			c.JSON(http.StatusOK, entity.ReportListTeachingLoadResult{})
			return
		}
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(ReportMsgNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}
