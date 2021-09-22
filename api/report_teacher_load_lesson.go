package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

// @Summary get teacher load Report
// @Description teacher load summary
// @Tags reports/teacherLoading
// @ID summaryTeacherLoadLessons
// @Accept json
// @Produce json
// @Param overview body entity.TeacherLoadLessonSummaryRequest true "request"
// @Success 200 {object} []entity.TeacherLoadLessonSummaryResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/teacher_loading/lessons_summary [post]
func (s *Server) summaryTeacherLoadLessons(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	var request entity.TeacherLoadLessonSummaryRequest
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
