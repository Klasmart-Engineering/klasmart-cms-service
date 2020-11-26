package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

// @Summary setUserSetting
// @ID setUserSetting
// @Description set user setting
// @Accept json
// @Produce json
// @Param userSetting body entity.UserSettingJsonContent true "user setting json content"
// @Tags userSetting
// @Success 200 {object} entity.IDResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /user_settings [post]
func (s *Server) setUserSetting(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	data := new(entity.UserSettingJsonContent)
	err := c.ShouldBindJSON(data)
	if err != nil {
		log.Info(ctx, "setUserSetting: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	id, err := model.GetUserSettingModel().SetByOperator(ctx, op, data)
	if err != nil {
		log.Info(ctx, "setUserSetting: GetUserSettingModel.SetByUserID error",
			log.Err(err),
			log.Any("op", op),
			log.Any("data", data),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, entity.IDResponse{ID: id})
}

// @Summary getUserSettingByOperator
// @ID getUserSettingByOperator
// @Description get user setting by user id
// @Accept json
// @Produce json
// @Tags userSetting
// @Success 200 {object} entity.UserSettingJsonContent
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /user_settings [get]
func (s *Server) getUserSettingByOperator(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	result, err := model.GetUserSettingModel().GetByOperator(ctx, op)
	if err == constant.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
		return
	}
	if err != nil {
		log.Info(ctx, "setUserSetting: GetUserSettingModel.getUserSettingByOperator error",
			log.Err(err),
			log.Any("op", op),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	c.JSON(http.StatusOK, result)
}
