package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
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
		err := errors.New("update daschedule: require id")
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
	if !data.IsForce {
		conflict, err := model.GetScheduleModel().IsScheduleConflict(ctx, operator, data.StartAt, data.EndAt)
		if err != nil {
			log.Error(ctx, "update schedule: check conflict failed",
				log.Int64("start_at", data.StartAt),
				log.Int64("end_at", data.EndAt),
			)
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		if conflict {
			log.Warn(ctx, "update schedule: time conflict",
				log.Int64("start_at", data.StartAt),
				log.Int64("end_at", data.EndAt),
			)
			c.JSON(http.StatusConflict, "update schedule: time conflict")
			return
		}
	}
	newID, err := model.GetScheduleModel().Update(ctx, dbo.MustGetDB(ctx), operator, &data)
	if err != nil {
		log.Error(ctx, "update schedule: update failed", log.Err(err))
		switch {
		case entity.IsErrInvalidArgs(err):
			c.JSON(http.StatusBadRequest, err.Error())
		case err == dbo.ErrRecordNotFound:
			c.JSON(http.StatusNotFound, err.Error())
		default:
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": newID,
	})
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
	if err := model.GetScheduleModel().Delete(ctx, dbo.MustGetDB(ctx), operator, id, editType); err != nil {
		log.Error(ctx, "delete schedule: delete failed",
			log.Err(err),
			log.String("schedule_id", id),
			log.String("repeat_edit_options", string(editType)),
		)
		switch {
		case entity.IsErrInvalidArgs(err):
			c.JSON(http.StatusBadRequest, err.Error())
		case err == dbo.ErrRecordNotFound:
			c.JSON(http.StatusNotFound, err.Error())
		default:
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
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
	data.OrgID = op.OrgID

	if err := utils.GetValidator().Struct(data); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		log.Info(ctx, "add schedule: verify data failed", log.Err(err))
		return
	}
	if !data.IsForce {
		conflict, err := model.GetScheduleModel().IsScheduleConflict(ctx, op, data.StartAt, data.EndAt)
		if err != nil {
			log.Error(ctx, "add schedule: check conflict failed",
				log.Int64("start_at", data.StartAt),
				log.Int64("end_at", data.EndAt),
			)
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		if conflict {
			log.Warn(ctx, "add schedule: time conflict",
				log.Int64("start_at", data.StartAt),
				log.Int64("end_at", data.EndAt),
			)
			c.JSON(http.StatusConflict, "add schedule: time conflict")
			return
		}
	}
	id, err := model.GetScheduleModel().Add(ctx, dbo.MustGetDB(ctx), op, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		log.Error(ctx, "add schedule error", log.Err(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}
func (s *Server) getScheduleByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	log.Info(ctx, "getScheduleByID", log.String("scheduleID", id))
	result, err := model.GetScheduleModel().GetByID(ctx, dbo.MustGetDB(ctx), id)
	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}
	if err == constant.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusInternalServerError, err.Error())
	log.Error(ctx, "get daschedule by id error", log.Err(err))
}
func (s *Server) querySchedule(c *gin.Context) {
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusBadRequest, responseMsg("operate not exist"))
		return
	}
	ctx := c.Request.Context()
	condition := new(da.ScheduleCondition)
	condition.OrderBy = da.NewScheduleOrderBy(c.Query("order_by"))
	teacherName := c.Query("teacher_name")
	if strings.TrimSpace(teacherName) == "" {
		c.JSON(http.StatusBadRequest, errors.New("teacherName is empty"))
		return
	}
	condition.Pager = utils.GetPager(c.Query("page"), c.Query("page_size"))
	startAtStr := c.Query("start_at")
	startAt, err := strconv.ParseInt(startAtStr, 10, 64)
	if err != nil {
		startAt = utils.BeginOfDayByTimeStamp(startAt).Unix()
	}
	condition.StartAtGe = sql.NullInt64{
		Int64: startAt,
		Valid: startAt == 0,
	}
	condition.OrgID = sql.NullString{
		String: op.OrgID,
		Valid:  op.OrgID != "",
	}
	log.Info(ctx, "querySchedule", log.Any("condition", condition))

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
	condition.TeacherID = sql.NullString{
		String: teacher.ID,
		Valid:  true,
	}
	log.Info(ctx, "querySchedule", log.Any("condition", condition))

	total, result, err := da.GetScheduleDA().PageByTeacherID(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		log.Error(ctx, "querySchedule:error", log.Err(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"data":  result,
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
	condition := &da.ScheduleCondition{
		OrgID: sql.NullString{
			String: "1",
			Valid:  true,
		},
		StartAtGe: sql.NullInt64{
			Int64: start,
			Valid: start != 0,
		},
		EndAtLe: sql.NullInt64{Valid: true, Int64: end},
	}

	result, err := model.GetScheduleModel().Query(ctx, dbo.MustGetDB(ctx), condition)
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
