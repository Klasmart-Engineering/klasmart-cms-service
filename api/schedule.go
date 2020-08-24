package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
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
		c.JSON(http.StatusBadRequest, err.Error())
		log.Info(ctx, "add schedule: should bind body failed", log.Err(err))
		return
	}

	if err := utils.GetValidator().Struct(data); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		log.Info(ctx, "add schedule: verify data failed", log.Err(err))
		return
	}

	err := model.GetScheduleModel().Add(ctx, op, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		log.Error(ctx, "add schedule error", log.Err(err))
		return
	}
	c.JSON(http.StatusOK, nil)
}
func (s *Server) getScheduleByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	log.Info(ctx, "getScheduleByID", log.String("scheduleID", id))
	result, err := model.GetScheduleModel().GetByID(ctx, id)
	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}
	if err == constant.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusInternalServerError, err.Error())
	log.Error(ctx, "get schedule by id error", log.Err(err))
}
func (s *Server) querySchedule(c *gin.Context) {
	ctx := c.Request.Context()
	teacherName := c.Query("teacher_name")
	startTimeStr := c.Query("start_at")
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		startTime = utils.BeginOfDayByTimeStamp(startTime).Unix()
	}
	log.Info(ctx, "querySchedule", log.String("teacherName", teacherName), log.Int64("start_at", startTime))
	if strings.TrimSpace(teacherName) == "" {
		c.JSON(http.StatusBadRequest, errors.New("teacherName is empty"))
		return
	}
	teacherService, err := external.GetTeacherServiceProvider()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		log.Error(ctx, "querySchedule:get teacher service provider error", log.Err(err))
		return
	}
	teachers, err := teacherService.Query(ctx, teacherName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		log.Error(ctx, "querySchedule:query teacher info error", log.Err(err))
		return
	}
	if len(teachers) <= 0 {
		c.JSON(http.StatusNotFound, errors.New("teacher info not found"))
		log.Info(ctx, "querySchedule:teacher info not found")
		return
	}
	teacher := teachers[0]
	lastKey, result, err := model.GetScheduleModel().PageByTeacherID(ctx, &da.ScheduleCondition{
		TeacherID: teacher.ID,
		StartAt:   startTime,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		log.Error(ctx, "querySchedule:error", log.Err(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"last_key": lastKey,
		"data":     result,
	})
}

const (
	ViewTypeDay      = "Day"
	ViewTypeWorkweek = "Workweek"
	ViewTypeWeek     = "Week"
	ViewTypeMonth    = "Month"
)

func (s *Server) queryHomeSchedule(c *gin.Context) {
	ctx := c.Request.Context()
	viewType := c.Query("view_type")
	timeAtStr := c.Query("time_at")
	timeAt, err := strconv.ParseInt(timeAtStr, 10, 64)
	if err != nil {
		timeAt = time.Now().Unix()
	}
	timeUtil := utils.TimeUtil{TimeStamp: timeAt}
	var (
		start int64
		end   int64
	)
	switch viewType {
	case ViewTypeDay:
		start = utils.BeginOfDayByTimeStamp(timeAt).Unix()
		end = utils.EndOfDayByTimeStamp(timeAt).Unix()
	case ViewTypeWorkweek:
		start, end = timeUtil.FindWorkWeekTimeRange()
	case ViewTypeWeek:
		start, end = timeUtil.FindWeekTimeRange()
	case ViewTypeMonth:
		start, end = timeUtil.FindMonthRange()
	default:
		c.JSON(http.StatusBadRequest, errors.New("view_type is required"))
		return
	}
	result, err := model.GetScheduleModel().Query(ctx, &da.ScheduleCondition{
		OrgID:       "",
		StartAt:     start,
		FilterEndAt: entity.NullInt64{Valid: true, Int64: end},
	})
	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}
	if err == constant.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusInternalServerError, err.Error())
}
