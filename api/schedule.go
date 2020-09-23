package api

import (
	"database/sql"
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

// @Summary updateSchedule
// @ID updateSchedule
// @Description update a schedule data
// @Accept json
// @Produce json
// @Param schedule_id path string true "schedule id"
// @Param scheduleData body entity.ScheduleUpdateView true "schedule data to update"
// @Tags schedule
// @Success 200 {object} entity.IDResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 409 {object} ConflictResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules/{schedule_id} [put]
func (s *Server) updateSchedule(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	data := entity.ScheduleUpdateView{}
	if err := c.ShouldBind(&data); err != nil {
		log.Info(ctx, "update schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	loc := utils.GetTimeLocationByOffset(data.TimeZoneOffset)
	log.Debug(ctx, "time location", log.Any("location", loc), log.Int("offset", data.TimeZoneOffset))
	data.ID = id
	if !data.EditType.Valid() {
		errMsg := "update schedule: invalid edit type"
		log.Info(ctx, errMsg, log.String("edit_type", string(data.EditType)))
		c.JSON(http.StatusBadRequest, L(Unknown))
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
	now := time.Now().Unix()
	if data.StartAt < now || data.StartAt >= data.EndAt {
		log.Info(ctx, "schedule start_at or end_at is invalid",
			log.Int64("StartAt", data.StartAt),
			log.Int64("EndAt", data.EndAt),
			log.Int64("now", now))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	if data.IsAllDay {
		timeUtil := utils.NewTimeUtil(data.StartAt, loc)
		data.StartAt = timeUtil.BeginOfDayByTimeStamp().Unix()
		timeUtil.TimeStamp = data.EndAt
		data.EndAt = timeUtil.EndOfDayByTimeStamp().Unix()
	}
	log.Debug(ctx, "request data", log.Any("operator", operator), log.Any("requestData", data))
	data.Location = loc
	newID, err := model.GetScheduleModel().Update(ctx, operator, &data)
	switch err {
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(Unknown))
	case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: newID})
	default:
		c.JSON(http.StatusInternalServerError, err.Error())
	}
}

// @Summary deleteSchedule
// @ID deleteSchedule
// @Description delete a schedule data
// @Accept json
// @Produce json
// @Param schedule_id path string true "schedule id"
// @Param repeat_edit_options query string true "repeat edit options" enums(only_current,with_following)
// @Tags schedule
// @Success 200 {string} string "OK"
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules/{schedule_id} [delete]
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

	err := model.GetScheduleModel().Delete(ctx, operator, id, editType)
	switch err {
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case dbo.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(Unknown))
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
	default:
		c.JSON(http.StatusInternalServerError, err.Error())
	}
}

// @Summary addSchedule
// @ID addSchedule
// @Description add a schedule data
// @Accept json
// @Produce json
// @Param scheduleData body entity.ScheduleAddView true "schedule data to add"
// @Tags schedule
// @Success 200 {object} entity.IDResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 409 {object} ConflictResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules [post]
func (s *Server) addSchedule(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	data := new(entity.ScheduleAddView)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "add schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	loc := utils.GetTimeLocationByOffset(data.TimeZoneOffset)
	log.Debug(ctx, "time location", log.Any("location", loc), log.Int("offset", data.TimeZoneOffset))
	data.OrgID = op.OrgID
	now := time.Now().Unix()
	if data.StartAt < now || data.StartAt >= data.EndAt {
		log.Info(ctx, "schedule start_at or end_at is invalid",
			log.Int64("StartAt", data.StartAt),
			log.Int64("EndAt", data.EndAt),
			log.Int64("now", now))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	if data.IsAllDay {
		timeUtil := utils.NewTimeUtil(data.StartAt, loc)
		data.StartAt = timeUtil.BeginOfDayByTimeStamp().Unix()
		timeUtil.TimeStamp = data.EndAt
		data.EndAt = timeUtil.EndOfDayByTimeStamp().Unix()
	}
	log.Debug(ctx, "request data", log.Any("operator", op), log.Any("requestData", data))
	// add schedule
	data.Location = loc
	id, err := model.GetScheduleModel().Add(ctx, op, data)
	switch err {
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(Unknown))
	case constant.ErrFileNotFound:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(Unknown))
	}
}

// @Summary getScheduleByID
// @ID getScheduleByID
// @Description get schedule by id
// @Accept json
// @Produce json
// @Param schedule_id path string true "schedule id"
// @Tags schedule
// @Success 200 {object} entity.ScheduleDetailsView
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules/{schedule_id} [get]
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
	log.Error(ctx, "get schedule by id error", log.Err(err), log.Any("id", id))
	c.JSON(http.StatusInternalServerError, err.Error())
}

// @Summary querySchedule
// @ID querySchedule
// @Description query schedule
// @Accept json
// @Produce json
// @Param teacher_name query string false "teacher name"
// @Param time_zone_offset query integer true "time zone offset"
// @Param start_at query integer false "search schedules by start_at"
// @Param order_by query string false "order by" enums(create_at, -create_at, start_at, -start_at)
// @Param page query integer false "page index, not paging if page <=0"
// @Param page_size query integer false "records per page, not paging if page_size <= 0"
// @Tags schedule
// @Success 200 {object} entity.SchedulePageView
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules [get]
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
			c.JSON(http.StatusBadRequest, L(Unknown))
			return
		}
		offsetStr := c.Query("time_zone_offset")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			log.Info(ctx, "querySchedule: time_zone_offset invalid", log.String("time_zone_offset", offsetStr))
			c.JSON(http.StatusBadRequest, L(Unknown))
			return
		}
		loc := utils.GetTimeLocationByOffset(offset)
		log.Debug(ctx, "time location", log.Any("location", loc), log.Int("offset", offset))
		timeUtil := utils.NewTimeUtil(startAt, loc)
		startAt = timeUtil.BeginOfDayByTimeStamp().Unix()
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
			c.JSON(http.StatusBadRequest, L(Unknown))
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
	total, result, err := model.GetScheduleModel().Page(ctx, condition)
	if err != nil {
		log.Error(ctx, "querySchedule:error", log.Any("condition", condition), log.Err(err))
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, entity.SchedulePageView{
		Total: total,
		Data:  result,
	})
}

const (
	ViewTypeDay      = "day"
	ViewTypeWorkweek = "work_week"
	ViewTypeWeek     = "week"
	ViewTypeMonth    = "month"
)

// @Summary getScheduleTimeView
// @ID getScheduleTimeView
// @Description get schedule time view
// @Accept json
// @Produce json
// @Param view_type query string true "search schedules by view_type" enums(day, work_week, week, month)
// @Param time_at query integer true "search schedules by time_at"
// @Param time_zone_offset query integer true "time zone offset"
// @Tags schedule
// @Success 200 {object} entity.ScheduleListView
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_time_view [get]
func (s *Server) getScheduleTimeView(c *gin.Context) {
	op := GetOperator(c)
	ctx := c.Request.Context()
	viewType := c.Query("view_type")
	timeAtStr := c.Query("time_at")
	timeAt, err := strconv.ParseInt(timeAtStr, 10, 64)
	if err != nil {
		log.Info(ctx, "getScheduleTimeView: time_at is empty or invalid", log.String("time_at", timeAtStr))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	offsetStr := c.Query("time_zone_offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		log.Info(ctx, "getScheduleTimeView: time_zone_offset invalid", log.String("time_zone_offset", offsetStr))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	loc := utils.GetTimeLocationByOffset(offset)
	log.Debug(ctx, "time location", log.Any("location", loc), log.Int("offset", offset))
	timeUtil := utils.NewTimeUtil(timeAt, loc)

	var (
		start int64
		end   int64
	)
	switch viewType {
	case ViewTypeDay:
		start = timeUtil.BeginOfDayByTimeStamp().Unix()
		end = timeUtil.EndOfDayByTimeStamp().Unix()
	case ViewTypeWorkweek:
		start, end = timeUtil.FindWorkWeekTimeRange()
	case ViewTypeWeek:
		start, end = timeUtil.FindWeekTimeRange()
	case ViewTypeMonth:
		start, end = timeUtil.FindMonthRange()
	default:
		log.Info(ctx, "getScheduleTimeView:view_type is empty or invalid", log.String("view_type", viewType))
		c.JSON(http.StatusBadRequest, L(Unknown))
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
	result, err := model.GetScheduleModel().Query(ctx, condition)
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
func (s *Server) updateStatus(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status")
	ctx := c.Request.Context()
	scheduleStatus := entity.ScheduleStatus(status)
	if !scheduleStatus.Valid() {
		log.Warn(ctx, "schedule status error", log.String("id", id), log.String("status", status))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	err := model.GetScheduleModel().UpdateScheduleStatus(ctx, dbo.MustGetDB(ctx), id, scheduleStatus)
	if err != nil {
		log.Error(ctx, "schedule status error",
			log.String("id", id),
			log.String("status", status),
			log.Err(err))
		c.JSON(http.StatusInternalServerError, L(Unknown))
		return
	}
	c.JSON(http.StatusOK, id)
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
