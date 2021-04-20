package api

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

// @Summary classAddMembersEvent
// @Description add members to class
// @Tags classAddMembersEvent
// @ID classAddMembersEvent
// @Accept json
// @Produce json
// @Param event body entity.ClassEventBody true "add member to class event"
// @Success 200 {object} SuccessRequestResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /classes_members [post]
func (s *Server) classAddMembersEvent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	log.Debug(ctx, "class add user event start")

	body := new(entity.ClassEventBody)
	if err := c.ShouldBind(body); err != nil {
		log.Info(ctx, "json data bind failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	event := new(entity.ClassUpdateMembersEvent)
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

	err := model.GetClassEventBusModel().PubAddMembers(ctx, op, event)
	if err != nil {
		log.Error(ctx, "class add user event error",
			log.Err(err),
			log.Any("event", event),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	log.Debug(ctx, "class add user event success", log.Any("event", event))
	c.JSON(http.StatusOK, nil)
}

// @Summary classDeleteMembersEvent
// @Description class delete members
// @Tags classDeleteMembersEvent
// @ID classDeleteMembersEvent
// @Accept json
// @Produce json
// @Param event body entity.ClassEventBody true "delete members to class event"
// @Success 200 {object} SuccessRequestResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /classes_members [delete]
func (s *Server) classDeleteMembersEvent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	log.Debug(ctx, "class delete user event start")

	body := new(entity.ClassEventBody)
	if err := c.ShouldBind(body); err != nil {
		log.Info(ctx, "json data bind failed",
			log.Err(err),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	event := new(entity.ClassUpdateMembersEvent)
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

	err := model.GetClassEventBusModel().PubDeleteMembers(ctx, op, event)
	if err != nil {
		log.Error(ctx, "class delete user event error",
			log.Err(err),
			log.Any("event", event),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	log.Debug(ctx, "class delete user event success", log.Any("event", event), log.String("token", body.Token))

	c.JSON(http.StatusOK, nil)
}
