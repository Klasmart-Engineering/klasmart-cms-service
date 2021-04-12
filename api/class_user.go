package api

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"net/http"
)

// @Summary addUserToClassEvent
// @Description add user to class event
// @Tags addUserToClassEvent
// @ID addUserToClassEvent
// @Accept json
// @Produce json
// @Param event body entity.ScheduleEventBody true "add user to class event"
// @Success 200 {object} SuccessRequestResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /classes_members [post]
func (s *Server) addUserToClassEvent(c *gin.Context) {
	ctx := c.Request.Context()
	//op := s.getOperator(c)
	log.Debug(ctx, "class add user event start")

	body := new(entity.ScheduleEventBody)
	if err := c.ShouldBind(body); err != nil {
		log.Info(ctx, "json data bind failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	event := new(entity.ScheduleClassEvent)
	if _, err := jwt.ParseWithClaims(body.Token, event, func(token *jwt.Token) (interface{}, error) {
		return config.Get().Schedule.ClassEventSecret, nil
	}); err != nil {
		log.Error(ctx, "class event error: parse with claims failed",
			log.Err(err),
			log.Any("token", body.Token),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	log.Debug(ctx, "class event", log.Any("event", event), log.String("token", body.Token))

	//err := model.GetScheduleEventModel().AddUserEvent(ctx, op, event)

	//switch event.Action {
	//case entity.ScheduleClassEventActionAdd:
	//	err = model.GetScheduleEventModel().AddUserEvent(ctx, op, event)
	//case entity.ScheduleClassEventActionDelete:
	//	err = model.GetScheduleEventModel().DeleteUserEvent(ctx, op, event)
	//default:
	//	log.Info(ctx, "event action not found", log.Any("event", event))
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//	return
	//}
	//switch err {
	//case nil:
	//	log.Debug(ctx, "class add user event success", log.Any("event", event))
	//	c.JSON(http.StatusOK, nil)
	//case constant.ErrInvalidArgs:
	//	log.Debug(ctx, "event args invalid", log.Any("event", event))
	//	c.JSON(http.StatusOK, nil)
	//default:
	//	log.Error(ctx, "class add user event error",
	//		log.Err(err),
	//		log.Any("event", event),
	//	)
	//	c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	//}
}

// @Summary deleteUserToClassEvent
// @Description delete user to class event
// @Tags deleteUserToClassEvent
// @ID deleteUserToClassEvent
// @Accept json
// @Produce json
// @Param event body entity.ScheduleEventBody true "add user to class event"
// @Success 200 {object} SuccessRequestResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /classes_members [delete]
func (s *Server) deleteUserToClassEvent(c *gin.Context) {
	ctx := c.Request.Context()
	//op := s.getOperator(c)
	log.Debug(ctx, "class add user event start")

	body := new(entity.ScheduleEventBody)
	if err := c.ShouldBind(body); err != nil {
		log.Info(ctx, "json data bind failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	event := new(entity.ScheduleClassEvent)
	if _, err := jwt.ParseWithClaims(body.Token, event, func(token *jwt.Token) (interface{}, error) {
		return config.Get().Schedule.ClassEventSecret, nil
	}); err != nil {
		log.Error(ctx, "class event error: parse with claims failed",
			log.Err(err),
			log.Any("token", body.Token),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	log.Debug(ctx, "class event", log.Any("event", event), log.String("token", body.Token))

	//err := model.GetScheduleEventModel().AddUserEvent(ctx, op, event)

	//switch event.Action {
	//case entity.ScheduleClassEventActionAdd:
	//	err = model.GetScheduleEventModel().AddUserEvent(ctx, op, event)
	//case entity.ScheduleClassEventActionDelete:
	//	err = model.GetScheduleEventModel().DeleteUserEvent(ctx, op, event)
	//default:
	//	log.Info(ctx, "event action not found", log.Any("event", event))
	//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	//	return
	//}
	//switch err {
	//case nil:
	//	log.Debug(ctx, "class add user event success", log.Any("event", event))
	//	c.JSON(http.StatusOK, nil)
	//case constant.ErrInvalidArgs:
	//	log.Debug(ctx, "event args invalid", log.Any("event", event))
	//	c.JSON(http.StatusOK, nil)
	//default:
	//	log.Error(ctx, "class add user event error",
	//		log.Err(err),
	//		log.Any("event", event),
	//	)
	//	c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	//}
}
