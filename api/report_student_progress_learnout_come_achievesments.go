package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary  getLearnOutcomeAchievement
// @Tags reports/studentProgress
// @ID getLearnOutcomeAchievement
// @Accept json
// @Produce json
// @Param request body entity.LearnOutcomeAchievementRequest true "request "
// @Success 200 {object} entity.LearnOutcomeAchievementResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_progress/learn_outcome_achievement [post]
func (s *Server) getLearnOutcomeAchievement(c *gin.Context) {
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
	op := s.getOperator(c)
	req := entity.LearnOutcomeAchievementRequest{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		log.Error(ctx, "invalid request", log.Err(err))
		err = constant.ErrInvalidArgs
		return
	}
	for _, duration := range req.Durations {
		_, _, err = duration.Value(ctx)
		if err != nil {
			return
		}
	}

	res, err := model.GetReportModel().GetStudentProgressLearnOutcomeAchievement(ctx, op, &req)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, res)
}
