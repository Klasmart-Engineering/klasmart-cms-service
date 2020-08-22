package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
)

func (s *Server) operator(c *gin.Context) *entity.Operator {
	panic("not implemented")
}

func (s *Server) deleteSchedule(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("schedule_id")
	if id == "" {
		err := errors.New("delete schedule: require param schedule_id")
		log.Error(ctx, err.Error())
		c.JSON(http.StatusBadRequest, err.Error())
	}
	editType := entity.ScheduleEditType(c.Query("repeat_edit_options"))
	if !editType.Valid() {
		err := errors.New("delete schedule: invalid edit type")
		log.Error(ctx, err.Error(), log.String("repeat_edit_options", string(editType)))
		c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := model.GetScheduleModel().Delete(ctx, s.operator(c), id, editType); err != nil {
		log.Error(ctx, "delete schedule: delete failed",
			log.Err(err),
			log.String("schedule_id", id),
			log.String("repeat_edit_options", string(editType)),
		)
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	c.JSON(http.StatusOK, "ok")
}
