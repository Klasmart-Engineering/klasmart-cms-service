package api

import (
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary get teacher load Report
// @Description teacher load list
// @Tags reports/teacherLoad
// @ID listTeacherLoadLessons
// @Accept json
// @Produce json
// @Param overview body entity.TeacherLoadLessonRequest true "request"
// @Success 200 {array}  entity.TeacherLoadLesson
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/teacher_load/lessons_list [post]
func (s *Server) listTeacherLoadLessons(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.TeacherLoadLessonRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "listTeacherLoadLessons: ShouldBindQuery failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	args, err := request.Validate(ctx, op)
	if err != nil {
		log.Error(ctx, "listTeacherLoadLessons: validate failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetReportModel().ListTeacherLoadLessons(ctx, op, &args)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary get teacher load Report
// @Description teacher load summary
// @Tags reports/teacherLoad
// @ID summaryTeacherLoadLessons
// @Accept json
// @Produce json
// @Param overview body entity.TeacherLoadLessonRequest true "request"
// @Success 200 {object} entity.TeacherLoadLessonSummary
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/teacher_load/lessons_summary [post]
func (s *Server) summaryTeacherLoadLessons(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.TeacherLoadLessonRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "summaryTeacherLoadLessons: ShouldBindQuery failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	args, err := request.Validate(ctx, op)
	if err != nil {
		log.Error(ctx, "summaryTeacherLoadLessons: validate failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetReportModel().SummaryTeacherLoadLessons(ctx, op, &args)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary get teacher missed lessons
// @Description teacher missed lessons
// @Tags reports/teacherLoad
// @ID listTeacherMissedLessons
// @Accept json
// @Produce json
// @Param overview body entity.TeacherLoadMissedLessonsRequest true "request"
// @Success 200 {object} entity.TeacherLoadMissedLessonsResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/teacher_load/missed_lessons [post]
func (s *Server) listTeacherMissedLessons(c *gin.Context) {
	ctx := c.Request.Context()
	var request entity.TeacherLoadMissedLessonsRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "summaryTeacherLoadMissedLessons: ShouldBindQuery failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if request.Page <= 0 {
		request.Page = constant.DefaultPageIndex
	}
	if request.PageSize <= 0 {
		request.PageSize = constant.DefaultPageSize
	}
	result, err := model.GetReportModel().MissedLessonsList(ctx, &request)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary get teacher load overview
// @Tags reports/teacherLoad
// @ID getTeacherLoadOverview
// @Accept json
// @Produce json
// @Param time_range query string true "time_range"
// @Success 200 {object} entity.TeacherLoadOverview
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/teacher_load_overview [get]
func (s *Server) getTeacherLoadOverview(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var err error
	var result *entity.TeacherLoadOverview
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
	teacherIDs, err := model.GetReportModel().GetTeacherIDsCanViewReports(ctx, op, external.TeacherViewPermissionParams{
		ViewOrgReports:    external.ReportOrganizationTeachingLoad617,
		ViewSchoolReports: external.ReportSchoolTeachingLoad618,
		ViewMyReports:     external.ReportMyTeachingLoad619,
	})
	tr := entity.TimeRange(c.Query("time_range"))
	result, err = model.GetReportTeachingLoadModel().GetTeacherLoadOverview(ctx, op, tr, teacherIDs)
	if err != nil {
		return
	}
	return
}
