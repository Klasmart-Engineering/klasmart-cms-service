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
	data := entity.ScheduleUpdateView{}
	if err := c.ShouldBind(&data); err != nil {
		log.Info(ctx, "update schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	data.ID = id
	if !data.EditType.Valid() {
		errMsg := "update schedule: invalid edit type"
		log.Info(ctx, errMsg, log.String("edit_type", string(data.EditType)))
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}
	//if strings.TrimSpace(data.Attachment) != "" {
	//	if !model.GetScheduleModel().ExistScheduleAttachmentFile(ctx, data.Attachment) {
	//		c.JSON(http.StatusBadRequest, "schedule attachment file not found")
	//		return
	//	}
	//}

	operator := GetOperator(c)
	data.OrgID = operator.OrgID
	if data.IsAllDay {
		data.StartAt = utils.BeginOfDayByTimeStamp(data.StartAt, s.getLocation(c)).Unix()
		data.EndAt = utils.EndOfDayByTimeStamp(data.EndAt, s.getLocation(c)).Unix()
	}
	log.Debug(ctx, "request data", log.Any("operator", operator), log.Any("requestData", data))
	newID, err := model.GetScheduleModel().Update(ctx, dbo.MustGetDB(ctx), operator, &data, s.getLocation(c))
	if err != nil {
		log.Info(ctx, "update schedule: update failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("data", data),
		)
		switch err {
		case constant.ErrInvalidArgs:
			c.JSON(http.StatusBadRequest, err.Error())
		case constant.ErrConflict:
			c.JSON(http.StatusConflict, err.Error())
		case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
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
	editType := entity.ScheduleEditType(c.Query("repeat_edit_options"))
	if !editType.Valid() {
		errMsg := "delete schedule: invalid edit type"
		log.Info(ctx, errMsg, log.String("edit_type", string(editType)))
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}
	operator := GetOperator(c)
	if err := model.GetScheduleModel().Delete(ctx, dbo.MustGetDB(ctx), operator, id, editType); err != nil {
		log.Info(ctx, "delete schedule: delete failed",
			log.Err(err),
			log.String("schedule_id", id),
			log.String("repeat_edit_options", string(editType)),
		)
		switch err {
		case constant.ErrInvalidArgs:
			c.JSON(http.StatusBadRequest, err.Error())
		case dbo.ErrRecordNotFound:
			c.JSON(http.StatusNotFound, err.Error())
		default:
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
}

func (s *Server) addSchedule(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	data := new(entity.ScheduleAddView)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "add schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	data.OrgID = op.OrgID
	if data.IsAllDay {
		data.StartAt = utils.BeginOfDayByTimeStamp(data.StartAt, s.getLocation(c)).Unix()
		data.EndAt = utils.EndOfDayByTimeStamp(data.EndAt, s.getLocation(c)).Unix()
	}
	log.Debug(ctx, "request data", log.Any("operator", op), log.Any("requestData", data))
	// add schedule
	id, err := model.GetScheduleModel().Add(ctx, dbo.MustGetDB(ctx), op, data, s.getLocation(c))
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"id": id,
		})
		return
	}
	if err == constant.ErrFileNotFound {
		log.Info(ctx, "add schedule: verify data failed,attachment not found", log.Err(err), log.Any("requestData", data))
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if err == constant.ErrConflict {
		log.Info(ctx, "add schedule: schedule start_at or end_at conflict", log.Err(err), log.Any("requestData", data))
		c.JSON(http.StatusConflict, err.Error())
		return
	}
	log.Error(ctx, "add schedule error", log.Err(err), log.Any("schedule", data))
	c.JSON(http.StatusInternalServerError, err.Error())
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
	log.Error(ctx, "get schedule by id error", log.Err(err), log.Any("id", id))
	c.JSON(http.StatusInternalServerError, err.Error())
}
func (s *Server) querySchedule(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	condition := new(da.ScheduleCondition)
	condition.OrderBy = da.NewScheduleOrderBy(c.Query("order_by"))
	condition.Pager = utils.GetDboPager(c.Query("page"), c.Query("page_size"))

	startAtStr := c.Query("start_at")
	if strings.TrimSpace(startAtStr) != "" {
		startAt, err := strconv.ParseInt(startAtStr, 10, 64)
		if err != nil {
			log.Info(ctx, "querySchedule:invalid start_at params",
				log.String("startAt", startAtStr),
				log.Any("condition", condition))
			c.JSON(http.StatusBadRequest, "invalid 'start_at' params")
		}
		startAt = utils.BeginOfDayByTimeStamp(startAt, s.getLocation(c)).Unix()
		condition.StartAtGe = sql.NullInt64{
			Int64: startAt,
			Valid: startAt > 0,
		}
	}

	condition.OrgID = sql.NullString{
		String: op.OrgID,
		Valid:  op.OrgID != "",
	}

	teacherName := c.Query("teacher_name")
	if strings.TrimSpace(teacherName) != "" {
		teachers, err := model.GetScheduleModel().GetTeacherByName(ctx, teacherName)
		if err != nil {
			log.Info(ctx, "get teacher info by name error",
				log.Err(err),
				log.String("teacherName", teacherName),
				log.Any("condition", condition))
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		if len(teachers) <= 0 {
			log.Info(ctx, "querySchedule:teacher info not found",
				log.String("teacherName", teacherName),
				log.Any("condition", condition))
			c.JSON(http.StatusBadRequest, "teacher info not found")
			return
		}
		teacherIDs := make([]string, len(teachers))
		for i, item := range teachers {
			teacherIDs[i] = item.ID
		}
		condition.TeacherIDs = entity.NullStrings{
			Valid:   len(teacherIDs) > 0,
			Strings: teacherIDs,
		}
	}
	log.Info(ctx, "querySchedule", log.Any("condition", condition))
	total, result, err := model.GetScheduleModel().Page(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		log.Error(ctx, "querySchedule:error", log.Any("condition", condition), log.Err(err))
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"data":  result,
	})
}

func (s *Server) getLocation(c *gin.Context) *time.Location {
	lc := GetTimeLocation(c)
	log.Debug(c.Request.Context(), "time location info", log.Any("location", lc))
	return lc
}

const (
	ViewTypeDay      = "day"
	ViewTypeWorkweek = "workWeek"
	ViewTypeWeek     = "week"
	ViewTypeMonth    = "month"
)

func (s *Server) getScheduleTimeView(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	viewType := c.Query("view_type")
	timeAtStr := c.Query("time_at")
	timeAt, err := strconv.ParseInt(timeAtStr, 10, 64)
	if err != nil {
		log.Info(ctx, "getScheduleTimeView: time_at is empty or invalid", log.String("time_at", timeAtStr))
		c.JSON(http.StatusBadRequest, errors.New("time_at is required"))
		return
	}
	timeUtil := utils.TimeUtil{TimeStamp: timeAt}
	var (
		start int64
		end   int64
	)
	switch viewType {
	case ViewTypeDay:
		start = utils.BeginOfDayByTimeStamp(timeAt, s.getLocation(c)).Unix()
		end = utils.EndOfDayByTimeStamp(timeAt, s.getLocation(c)).Unix()
	case ViewTypeWorkweek:
		start, end = timeUtil.FindWorkWeekTimeRange(s.getLocation(c))
	case ViewTypeWeek:
		start, end = timeUtil.FindWeekTimeRange(s.getLocation(c))
	case ViewTypeMonth:
		start, end = timeUtil.FindMonthRange(s.getLocation(c))
	default:
		log.Info(ctx, "getScheduleTimeView:view_type is empty or invalid", log.String("view_type", viewType))
		c.JSON(http.StatusBadRequest, errors.New("view_type is required"))
		return
	}
	startAndEndTimeViewRange := make([]sql.NullInt64, 2)
	startAndEndTimeViewRange[0] = sql.NullInt64{
		Valid: start <= 0,
		Int64: start,
	}
	startAndEndTimeViewRange[1] = sql.NullInt64{
		Valid: end <= 0,
		Int64: end,
	}
	condition := &da.ScheduleCondition{
		OrgID: sql.NullString{
			String: op.OrgID,
			Valid:  op.OrgID != "",
		},
		StartAndEndTimeViewRange: startAndEndTimeViewRange,
	}

	log.Debug(ctx, "condition info", log.String("viewType", viewType), log.String("timeAtStr", timeAtStr), log.Any("condition", condition))
	result, err := model.GetScheduleModel().Query(ctx, dbo.MustGetDB(ctx), condition)
	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}
	if err == constant.ErrRecordNotFound {
		log.Info(ctx, "record not found", log.String("viewType", viewType), log.String("timeAtStr", timeAtStr), log.Any("condition", condition))
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	log.Info(ctx, "record not found", log.Err(err), log.String("viewType", viewType), log.String("timeAtStr", timeAtStr), log.Any("condition", condition))
	c.JSON(http.StatusInternalServerError, err.Error())
}

//func (s *Server) getAttachmentUploadPath(c *gin.Context) {
//	ctx := c.Request.Context()
//	ext := c.Param("ext")
//	name := fmt.Sprintf("%s.%s", utils.NewID(), ext)
//	url, err := storage.DefaultStorage().GetUploadFileTempPath(ctx, model.ScheduleAttachment_Storage_Partition, name)
//	if err != nil {
//		log.Error(ctx, "uploadAttachment:get upload file path error", log.Err(err), log.String("fileName", name), log.String("partition", model.ScheduleAttachment_Storage_Partition))
//		c.JSON(http.StatusInternalServerError, err.Error())
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{
//		"attachment_url": url,
//	})
//}
