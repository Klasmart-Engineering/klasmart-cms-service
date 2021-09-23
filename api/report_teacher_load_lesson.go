package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
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
	result, err := model.GetTeacherLoadLessonsModel().List(ctx, op, &args)
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
	result, err := model.GetTeacherLoadLessonsModel().Summary(ctx, op, &args)
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
	op := s.getOperator(c)

	var request entity.TeacherLoadMissedLessonsRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error(ctx, "summaryTeacherLoadMissedLessons: ShouldBindQuery failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	args, err := request.Validate(ctx, op)
	if err != nil {
		log.Error(ctx, "summaryTeacherLoadMissedLessons: validate failed",
			log.Err(err),
			log.Any("request", request))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetTeacherLoadLessonsModel().MissedLessonsList(ctx, op, &args)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.defaultErrorHandler(c, err)
	}
}
