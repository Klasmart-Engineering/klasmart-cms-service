package api

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/storage"
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
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "update schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	data.ID = id
	operator, _ := GetOperator(c)
	newID, err := model.GetScheduleModel().Update(ctx, dbo.MustGetDB(ctx), operator, &data)
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
		case dbo.ErrRecordNotFound:
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
		log.Info(ctx, errMsg, log.String("repeat_edit_options", string(editType)))
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}
	operator, _ := GetOperator(c)
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
	// validate data
	if err := utils.GetValidator().Struct(data); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		log.Info(ctx, "add schedule: verify data failed", log.Err(err))
		return
	}
	// add schedule
	id, err := model.GetScheduleModel().Add(ctx, dbo.MustGetDB(ctx), op, data)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"id": id,
		})
		return
	}
	if err == constant.ErrFileNotFound {
		c.JSON(http.StatusBadRequest, err.Error())
		log.Info(ctx, "add schedule: verify data failed", log.Err(err), log.Any("requestData", data))
		return
	}
	if err == constant.ErrConflict {
		c.JSON(http.StatusConflict, err.Error())
		log.Info(ctx, "add schedule: schedule start_at or end_at conflict", log.Err(err), log.Any("requestData", data))
		return
	}
	c.JSON(http.StatusInternalServerError, err.Error())
	log.Error(ctx, "add schedule error", log.Err(err), log.Any("schedule", data))
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
	log.Error(ctx, "get schedule by id error", log.Err(err), log.Any("id", id))
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
		startAt = utils.BeginOfDayByTimeStamp(time.Now().Unix()).Unix()
	}
	startAt = utils.BeginOfDayByTimeStamp(startAt).Unix()
	condition.StartAtLe = sql.NullInt64{
		Int64: startAt,
		Valid: startAt == 0,
	}
	condition.OrgID = sql.NullString{
		String: op.OrgID,
		Valid:  op.OrgID != "",
	}
	log.Info(ctx, "querySchedule", log.Any("condition", condition))

	teachers, err := model.GetScheduleModel().GetTeacherByName(ctx, teacherName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if len(teachers) <= 0 {
		log.Info(ctx, "querySchedule:teacher info not found")
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
	log.Info(ctx, "querySchedule", log.Any("condition", condition))

	total, result, err := da.GetScheduleDA().PageByTeacherID(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		log.Error(ctx, "querySchedule:error", log.Any("condition", condition), log.Err(err))
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

func (s *Server) getScheduleTimeView(c *gin.Context) {
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusBadRequest, responseMsg("operate not exist"))
		return
	}
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
		log.Info(ctx, "getScheduleTimeView:view_type is empty or invalid", log.String("view_type", viewType))
		return
	}
	condition := &da.ScheduleCondition{
		OrgID: sql.NullString{
			String: op.OrgID,
			Valid:  op.OrgID != "",
		},
		StartAtLe: sql.NullInt64{
			Int64: start,
			Valid: start > 0,
		},
		EndAtGe: sql.NullInt64{
			Valid: end > 0,
			Int64: end,
		},
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

func (s *Server) getAttachmentUploadPath(c *gin.Context) {
	ctx := c.Request.Context()
	ext := c.Param("ext")
	name := fmt.Sprintf("%s.%s", utils.NewID(), ext)
	url, err := storage.DefaultStorage().GetUploadFileTempPath(ctx, model.ScheduleAttachment_Storage_Partition, name)
	if err != nil {
		log.Error(ctx, "uploadAttachment:get upload file path error", log.Err(err), log.String("fileName", name), log.String("partition", model.ScheduleAttachment_Storage_Partition))
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"attachment_url": url,
	})
}
