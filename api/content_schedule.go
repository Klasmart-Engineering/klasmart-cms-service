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
// @Param group_name query da.LessonPlanGroupName true "group_name=Organization%20Content,Badanamu%20Content,%20More Featured Content"
// @Param page_size query int false "page"
// @Param page query int false "page size"
// @Success 200 {object} entity.GetLessonPlansCanScheduleResponse
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

	groupNames := c.QueryArray("group_name")
	if len(groupNames) == 0 {
		log.Info(ctx, "query group_name is required")
		err = constant.ErrInvalidArgs
		return
	}
	for _, gn := range groupNames {
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

	condition := &entity.ContentConditionRequest{
		ContentType:   []int{entity.ContentTypePlan},
		PublishStatus: []string{entity.ContentStatusPublished},
		OrderBy:       "create_at",
		Pager:         utils.GetPager(c.Query("page"), c.Query("page_size")),
		GroupNames:    groupNames,
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
