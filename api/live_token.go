package api

import (
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary getScheduleLiveToken
// @ID getScheduleLiveToken
// @Description get schedule live token
// @Accept json
// @Produce json
// @Param schedule_id path string true "schedule id"
// @Param live_token_type query string true "live token type" enums(preview, live)
// @Tags schedule
// @Success 200 {object} entity.LiveTokenView
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules/{schedule_id}/live/token [get]
func (s *Server) getScheduleLiveToken(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	scheduleID := c.Param("id")
	tokenType := entity.LiveTokenType(c.Query("live_token_type"))
	if !tokenType.Valid() {
		log.Info(ctx, "token type is invalid", log.String("tokenType", string(tokenType)))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	token, err := model.GetLiveTokenModel().MakeScheduleLiveToken(ctx, op, scheduleID, tokenType)
	switch err {
	case nil:
		c.JSON(http.StatusOK, entity.LiveTokenView{Token: token})
	case constant.ErrRecordNotFound:
		log.Info(ctx, "schedule not found", log.Err(err), log.String("scheduleID", scheduleID))
		c.JSON(http.StatusNotFound, L(ScheduleMsgEditOverlap))
	case model.ErrGoLiveTimeNotUp:
		log.Info(ctx, "go live time not up", log.Err(err), log.String("scheduleID", scheduleID))
		c.JSON(http.StatusBadRequest, L(ScheduleMsgGoLiveTimeNotUp))
	case model.ErrGoLiveNotAllow:
		log.Info(ctx, "go live not allow", log.Err(err), log.String("scheduleID", scheduleID))
		c.JSON(http.StatusBadRequest, L(ScheduleMsgTimeExpired))
	case constant.ErrInvalidArgs:
		log.Info(ctx, "invalid args", log.Err(err), log.String("scheduleID", scheduleID))
		c.JSON(http.StatusBadRequest, L(ScheduleMsgEditOverlap))
	default:
		log.Error(ctx, "make schedule live token error", log.Err(err), log.String("scheduleID", scheduleID))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary getContentLiveToken
// @ID getContentLiveToken
// @Description get content live token
// @Accept json
// @Produce json
// @Param content_id path string true "content id"
// @Tags content
// @Success 200 {object} entity.LiveTokenView
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /contents/{content_id}/live/token [get]
func (s *Server) getContentLiveToken(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	contentID := c.Param("content_id")
	token, err := model.GetLiveTokenModel().MakeContentLiveToken(ctx, op, contentID)
	switch err {
	case nil:
		c.JSON(http.StatusOK, entity.LiveTokenView{Token: token})
	case constant.ErrRecordNotFound:
		log.Info(ctx, "content not found", log.Err(err), log.String("contentID", contentID))
		c.JSON(http.StatusNotFound, L(GeneralUnknown))

	default:
		log.Error(ctx, "make content live token error", log.Err(err), log.String("contentID", contentID))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
