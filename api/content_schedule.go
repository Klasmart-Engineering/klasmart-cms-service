package api

import (
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

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
// @Param page_size query int false "page"
// @Param page query int false "page size"
// @Param request body entity.ContentConditionRequest true "request args"
// @Success 200 {object} entity.GetLessonPlansCanScheduleResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /contents_lesson_plans [post]
func (s *Server) getLessonPlansCanSchedule(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	var err error
	defer func() {
		if err == nil {
			return
		}
		switch err {
		case constant.ErrInvalidArgs:
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		default:
			s.defaultErrorHandler(c, err)
		}
	}()
	condition := &entity.ContentConditionRequest{}
	err = c.ShouldBindJSON(condition)
	if err != nil {
		log.Error(ctx, "invalid args", log.Err(err))
		err = constant.ErrInvalidArgs
		return
	}
	condition.Pager = utils.GetPager(c.Query("page"), c.Query("page_size"))

	if len(condition.GroupNames) == 0 {
		log.Info(ctx, "group_name is required")
		err = constant.ErrInvalidArgs
		return
	}
	for _, gn := range condition.GroupNames {
		if !utils.ContainsString([]string{
			entity.LessonPlanGroupNameOrganizationContent.String(),
			entity.LessonPlanGroupNameBadanamuContent.String(),
			entity.LessonPlanGroupNameMoreFeaturedContent.String(),
		}, gn) {
			log.Info(ctx, "group_name is invalid", log.Any("group_name", gn))
			err = constant.ErrInvalidArgs
			return
		}
	}

	condition.OrderBy = "create_at"
	condition.ContentType = []int{entity.ContentTypePlan}
	condition.PublishStatus = []string{entity.ContentStatusPublished}
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
