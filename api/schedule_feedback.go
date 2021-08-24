package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary addScheduleFeedback
// @ID addScheduleFeedback
// @Description add ScheduleFeedback
// @Accept json
// @Produce json
// @Param feedback body entity.ScheduleFeedbackAddInput true "feedback data"
// @Tags scheduleFeedback
// @Success 200 {object} SuccessRequestResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_feedbacks [post]
func (s *Server) addScheduleFeedback(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	data := &entity.ScheduleFeedbackAddInput{}
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	id, err := model.GetScheduleFeedbackModel().Add(ctx, op, data)
	switch err {
	case nil:
		c.JSON(http.StatusOK, D(IDResponse{ID: id}))
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(ScheduleMessageEditOverlap))
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrOnlyStudentCanSubmitFeedback:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrHomeFunStudyHasCompleted:
		c.JSON(http.StatusBadRequest, L(ScheduleFeedbackCompleted))
	default:
		s.jsonInternalServerError(c, err)
	}
}

// @Summary queryFeedback
// @ID queryFeedback
// @Description query feedback list
// @Accept json
// @Produce json
// @Param schedule_id query string false "schedule id"
// @Param user_id query string false "user id"
// @Tags scheduleFeedback
// @Success 200 {array} entity.ScheduleFeedbackView
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_feedbacks [get]
func (s *Server) queryFeedback(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()

	scheduleID := c.Query("schedule_id")
	userID := c.Query("user_id")
	result, err := model.GetScheduleFeedbackModel().Query(ctx, op, &da.ScheduleFeedbackCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  scheduleID != "",
		},
		UserID: sql.NullString{
			String: userID,
			Valid:  userID != "",
		},
	})
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		s.jsonInternalServerError(c, err)
	}
}
