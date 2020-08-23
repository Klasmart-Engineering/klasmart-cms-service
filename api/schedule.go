package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
)

func (s *Server) updateSchedule(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	if id == "" {
		err := errors.New("update schedule: require id")
		log.Error(ctx, err.Error())
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	data := entity.ScheduleUpdateView{}
	if err := c.ShouldBind(data); err != nil {
		log.Error(ctx, "update schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	data.ID = id
	operator, _ := GetOperator(c)
	if err := model.GetScheduleModel().Update(ctx, operator, &data); err != nil {
		log.Error(ctx, "update schedule: update failed", log.Err(err))
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
}

func (s *Server) deleteSchedule(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	if id == "" {
		err := errors.New("delete schedule: require param id")
		log.Error(ctx, err.Error())
		c.JSON(http.StatusBadRequest, err.Error())
	}
	editType := entity.ScheduleEditType(c.Query("repeat_edit_options"))
	if !editType.Valid() {
		err := errors.New("delete schedule: invalid edit type")
		log.Error(ctx, err.Error(), log.String("repeat_edit_options", string(editType)))
		c.JSON(http.StatusBadRequest, err.Error())
	}
	operator, _ := GetOperator(c)
	if err := model.GetScheduleModel().Delete(ctx, operator, id, editType); err != nil {
		log.Error(ctx, "delete schedule: delete failed",
			log.Err(err),
			log.String("schedule_id", id),
			log.String("repeat_edit_options", string(editType)),
		)
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
}

func (s *Server) addSchedule(c *gin.Context) {
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusBadRequest, responseMsg("operate not exist"))
		return
	}
	ctx := c.Request.Context()
	data := new(entity.ScheduleAddView)
	if err := c.ShouldBind(data); err != nil {
		log.Error(ctx, "add schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := utils.GetValidator().Struct(data); err != nil {
		log.Error(ctx, "add schedule: verify data failed", log.Err(err))
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err := model.GetScheduleModel().Add(ctx, op, data)
	if err != nil {
		log.Error(ctx, "add schedule error", log.Err(err))
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, nil)
}
func (s *Server) getScheduleByID(c *gin.Context) {

}
func (s *Server) querySchedule(c *gin.Context) {

}
