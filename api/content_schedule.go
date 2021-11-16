package api

import (
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary getLessonPlansCanSchedule
// @ID getLessonPlansCanSchedule
// @Description get lesson plans for schedule
// @Accept json
// @Produce json
// @Tags content
// @Success 200 {object} []entity.LessonPlanForSchedule
// @Failure 500 {object} InternalServerErrorResponse
// @Router /contents_lesson_plans [get]
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

	condition := &entity.ContentConditionRequest{
		ContentType:   []int{entity.ContentTypePlan},
		PublishStatus: []string{entity.ContentStatusPublished},
		OrderBy:       "create_at",
	}
	err = model.GetContentFilterModel().FilterPublishContent(ctx, condition, op)
	if err == model.ErrNoAvailableVisibilitySettings {
		condition.VisibilitySettings = []string{"none"}
		err = nil
	}
	if err != nil {
		return
	}
	r, err := model.GetContentModel().GetLessonPlansCanSchedule(ctx, op, condition)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, r)
}
