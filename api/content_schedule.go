package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary getLessonPlansCanSchedule
// @ID getLessonPlansCanSchedule
// @Description get lesson plans for schedule
// @Accept json
// @Produce json
// @Tags content
// @Success 200 {object} []*entity.LessonPlanForSchedule
// @Failure 500 {object} InternalServerErrorResponse
// @Router /contents/schedule_lesson_plans [get]
func (s *Server) getLessonPlansCanSchedule(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	var err error
	defer func() {
		if err == nil {
			return
		}
		switch err {
		default:
			s.defaultErrorHandler(c, err)
		}
	}()

	r, err := model.GetContentModel().GetLessonPlansCanSchedule(ctx, op)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, r)
}
