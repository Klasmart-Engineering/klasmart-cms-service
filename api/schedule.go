package api

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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
	op := s.getOperator(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	data := entity.ScheduleUpdateView{}
	if err := c.ShouldBind(&data); err != nil {
		log.Info(ctx, "update schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err := model.GetSchedulePermissionModel().HasScheduleEditPermission(c, op, data.ClassID)
	if err == constant.ErrUnAuthorized {
		c.JSON(http.StatusForbidden, L(ScheduleMsgNoPermission))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

	loc := utils.GetTimeLocationByOffset(data.TimeZoneOffset)
	log.Debug(ctx, "time location", log.Any("location", loc), log.Int("offset", data.TimeZoneOffset))
	data.ID = id
	if !data.EditType.Valid() {
		errMsg := "update schedule: invalid edit type"
		log.Info(ctx, errMsg, log.String("edit_type", string(data.EditType)))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	//if strings.TrimSpace(data.Attachment) != "" {
	//	if !model.GetScheduleModel().ExistScheduleAttachmentFile(ctx, data.Attachment) {
	//		c.JSON(http.StatusBadRequest, "schedule attachment file not found")
	//		return
	//	}
	//}

	operator := s.getOperator(c)
	data.OrgID = operator.OrgID
	now := time.Now().Unix()

	if (data.ClassType != entity.ScheduleClassTypeHomework) &&
		(!data.IsRepeat || (data.IsRepeat && data.EditType == entity.ScheduleEditOnlyCurrent)) {
		if data.StartAt < now || data.StartAt >= data.EndAt {
			log.Info(ctx, "schedule start_at or end_at is invalid",
				log.Int64("StartAt", data.StartAt),
				log.Int64("EndAt", data.EndAt),
				log.Int64("now", now),
				log.Any("data", data),
			)
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return
		}
	}
	start, end, ok := s.processScheduleDueDate(c, data.StartAt, data.EndAt, data.DueAt, data.ClassType, loc)
	if !ok {
		log.Info(ctx, "process schedule due date failure")
		return
	}
	data.StartAt = start
	data.EndAt = end

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
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(ScheduleMsgOverlap))
	case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(ScheduleMsgEditOverlap))
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgEditOverlap))
	case model.ErrScheduleEditMissTime:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgEditMissTime))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: newID})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
	op := s.getOperator(c)
	err := model.GetSchedulePermissionModel().HasScheduleOrgPermission(c, op, external.ScheduleDeleteEvent)
	if err == constant.ErrForbidden {
		c.JSON(http.StatusForbidden, L(ScheduleMsgNoPermission))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	ctx := c.Request.Context()
	id := c.Param("id")
	editType := entity.ScheduleEditType(c.Query("repeat_edit_options"))
	if !editType.Valid() {
		errMsg := "delete schedule: invalid edit type"
		log.Info(ctx, errMsg, log.String("edit_type", string(editType)))
		c.JSON(http.StatusBadRequest, errMsg)
		return
	}

	err = model.GetScheduleModel().Delete(ctx, op, id, editType)
	switch err {
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case dbo.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgEditOverlap))
	case model.ErrScheduleEditMissTime:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgDeleteMissTime))
	case model.ErrScheduleLessonPlanUnAuthed:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgLessonPlanInvalid))
	case nil:
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules [post]
func (s *Server) addSchedule(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	data := new(entity.ScheduleAddView)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "add schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	err := model.GetSchedulePermissionModel().HasScheduleEditPermission(c, op, data.ClassID)
	if err == constant.ErrUnAuthorized {
		c.JSON(http.StatusForbidden, L(ScheduleMsgNoPermission))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	loc := utils.GetTimeLocationByOffset(data.TimeZoneOffset)
	log.Debug(ctx, "time location", log.Any("location", loc), log.Int("offset", data.TimeZoneOffset))
	data.OrgID = op.OrgID
	now := time.Now().Unix()
	if (data.ClassType != entity.ScheduleClassTypeHomework) &&
		(!data.IsRepeat && (data.StartAt < now || data.StartAt >= data.EndAt)) {
		log.Info(ctx, "schedule start_at or end_at is invalid",
			log.Int64("StartAt", data.StartAt),
			log.Int64("EndAt", data.EndAt),
			log.Int64("now", now))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	start, end, ok := s.processScheduleDueDate(c, data.StartAt, data.EndAt, data.DueAt, data.ClassType, loc)
	if !ok {
		log.Info(ctx, "process schedule due date failure")
		return
	}
	data.StartAt = start
	data.EndAt = end

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
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(ScheduleMsgOverlap))
	case constant.ErrFileNotFound:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrScheduleLessonPlanUnAuthed:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgLessonPlanInvalid))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

func (s *Server) processScheduleDueDate(c *gin.Context, startSrc int64, endSrc int64, dueAt int64, classType entity.ScheduleClassType, loc *time.Location) (int64, int64, bool) {
	if dueAt <= 0 {
		return startSrc, endSrc, true
	}
	now := time.Now().Unix()
	ctx := c.Request.Context()
	lable := GeneralUnknown
	var day int64
	switch classType {
	case entity.ScheduleClassTypeTask:
		day = utils.GetTimeDiffToDayByTimeStamp(endSrc, dueAt, loc)
		lable = ScheduleMsgDueDateEarlierEndDate
	case entity.ScheduleClassTypeHomework:
		day = utils.GetTimeDiffToDayByTimeStamp(now, dueAt, loc)
		startSrc = utils.StartOfDayByTimeStamp(dueAt, loc)
		endSrc = utils.EndOfDayByTimeStamp(dueAt, loc)
		lable = ScheduleMsgDueDateEarlierToDay
	}
	if day < 0 {
		log.Info(ctx, "schedule dueAt is invalid",
			log.Int64("StartAt", startSrc),
			log.Int64("EndAt", endSrc),
			log.Int64("now", now),
			log.Int64("DueAt", dueAt),
			log.Any("classType", classType))
		c.JSON(http.StatusBadRequest, L(lable))
		return 0, 0, false
	}

	return startSrc, endSrc, true
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
	operator := s.getOperator(c)
	log.Info(ctx, "getScheduleByID", log.String("scheduleID", id))
	result, err := model.GetScheduleModel().GetByID(ctx, operator, id)
	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}
	if err == constant.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, L(ScheduleMsgDeleteOverlap))
		return
	}
	log.Error(ctx, "get schedule by id error", log.Err(err), log.Any("id", id))
	c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
	op := s.getOperator(c)
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
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return
		}
		offsetStr := c.Query("time_zone_offset")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			log.Info(ctx, "querySchedule: time_zone_offset invalid", log.String("time_zone_offset", offsetStr))
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
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

	filterClassIDs, err := model.GetSchedulePermissionModel().GetClassIDs(ctx, op)
	if err != nil {
		log.Error(ctx, "querySchedule:getClassIDs error",
			log.Err(err),
			log.Any("op", op),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if len(filterClassIDs) == 0 {
		log.Info(ctx, "querySchedule:filterClassIDs is empty", log.Any("operator", op))
		c.JSON(http.StatusOK, entity.SchedulePageView{
			Total: 0,
			Data:  []*entity.ScheduleSearchView{},
		})
		return
	}

	teacherName := c.Query("teacher_name")
	if strings.TrimSpace(teacherName) != "" {
		teachers, err := model.GetScheduleModel().GetTeacherByName(ctx, op, op.OrgID, teacherName)
		if err != nil {
			log.Info(ctx, "get teacher info by name error",
				log.Err(err),
				log.String("teacherName", teacherName),
				log.Any("operator", op),
				log.Any("condition", condition))
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		if len(teachers) <= 0 {
			log.Info(ctx, "querySchedule:teacher info not found",
				log.String("teacherName", teacherName),
				log.Any("operator", op),
				log.Any("condition", condition))
			c.JSON(http.StatusOK, entity.SchedulePageView{
				Total: 0,
				Data:  []*entity.ScheduleSearchView{},
			})
			return
		}
		teacherIDs := make([]string, len(teachers))
		for i, item := range teachers {
			teacherIDs[i] = item.ID
		}
		teacherClassIDs, err := model.GetScheduleModel().GetOrgClassIDsByUserIDs(ctx, op, teacherIDs, op.OrgID)
		if err != nil {
			log.Error(ctx, "querySchedule:GetScheduleModel.GetOrgClassIDsByUserIDs error",
				log.Err(err),
				log.Any("op", op),
				log.Strings("teacherIDs", teacherIDs),
			)
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
		log.Debug(ctx, "querySchedule:debug",
			log.Strings("teacherIDs", teacherIDs),
			log.Any("operator", op),
			log.Strings("teacherClassIDs", teacherClassIDs),
		)
		filterClassIDs = utils.IntersectAndDeduplicateStrSlice(filterClassIDs, teacherClassIDs)
	}
	condition.ClassIDs = entity.NullStrings{
		Strings: filterClassIDs,
		Valid:   true,
	}
	log.Info(ctx, "querySchedule", log.Any("condition", condition))
	total, result, err := model.GetScheduleModel().Page(ctx, op, condition)
	if err != nil {
		log.Error(ctx, "querySchedule:error", log.Any("condition", condition), log.Err(err))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, entity.SchedulePageView{
		Total: total,
		Data:  result,
	})
}

// @Summary getScheduleTimeView
// @ID getScheduleTimeView
// @Description get schedule time view
// @Accept json
// @Produce json
// @Param view_type query string true "search schedules by view_type" enums(day, work_week, week, month,year)
// @Param time_at query integer true "search schedules by time_at"
// @Param time_zone_offset query integer true "time zone offset"
// @Param school_ids query string false "school ids,separated by comma"
// @Param teacher_ids query string false "teacher id,separated by comma"
// @Param class_ids query string false "class id,separated by comma"
// @Param subject_ids query string false "subject id,separated by comma"
// @Param program_ids query string false "program id,separated by comma"
// @Tags schedule
// @Success 200 {object} entity.ScheduleListView
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_time_view [get]
func (s *Server) getScheduleTimeView(c *gin.Context) {
	ctx := c.Request.Context()
	offsetStr := c.Query("time_zone_offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		log.Info(ctx, "getScheduleTimeView: time_zone_offset invalid", log.String("time_zone_offset", offsetStr))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	}
	loc := utils.GetTimeLocationByOffset(offset)

	condition, err := s.getScheduleTimeViewCondition(c)
	if err != nil {
		return
	}
	result, err := model.GetScheduleModel().Query(ctx, condition, loc)
	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}
	if err == constant.ErrRecordNotFound {
		log.Info(ctx, "record not found", log.Any("condition", condition))
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
}

// @Summary getScheduledDates
// @ID getScheduledDates
// @Description get schedules dates(format:2006-01-02)
// @Accept json
// @Produce json
// @Param view_type query string true "search schedules by view_type" enums(day, work_week, week, month,year)
// @Param time_at query integer true "search schedules by time_at"
// @Param time_zone_offset query integer true "time zone offset"
// @Param school_ids query string false "school ids,separated by comma"
// @Param teacher_ids query string false "teacher id,separated by comma"
// @Param class_ids query string false "class id,separated by comma"
// @Param subject_ids query string false "subject id,separated by comma"
// @Param program_ids query string false "program id,separated by comma"
// @Tags schedule
// @Success 200 {array}  string
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_time_view/dates [get]
func (s *Server) getScheduledDates(c *gin.Context) {
	ctx := c.Request.Context()
	offsetStr := c.Query("time_zone_offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		log.Info(ctx, "getScheduleTimeView: time_zone_offset invalid", log.String("time_zone_offset", offsetStr))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	}
	loc := utils.GetTimeLocationByOffset(offset)

	condition, err := s.getScheduleTimeViewCondition(c)
	if err != nil {
		return
	}
	result, err := model.GetScheduleModel().QueryScheduledDates(ctx, condition, loc)
	if err != nil {
		log.Error(ctx, "getScheduledDates:GetScheduleModel.QueryScheduledDates error", log.Err(err), log.Any("condition", condition))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) getScheduleTimeViewCondition(c *gin.Context) (*da.ScheduleCondition, error) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	viewType := c.Query("view_type")
	timeAtStr := c.Query("time_at")
	timeAt, err := strconv.ParseInt(timeAtStr, 10, 64)
	if err != nil {
		log.Info(ctx, "getScheduleTimeView: time_at is empty or invalid", log.String("time_at", timeAtStr))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return nil, err
	}
	offsetStr := c.Query("time_zone_offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		log.Info(ctx, "getScheduleTimeView: time_zone_offset invalid", log.String("time_zone_offset", offsetStr))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return nil, err
	}
	loc := utils.GetTimeLocationByOffset(offset)
	log.Debug(ctx, "time location", log.Any("op", op), log.Any("location", loc), log.Int("offset", offset))
	timeUtil := utils.NewTimeUtil(timeAt, loc)

	var (
		start int64
		end   int64
	)
	switch entity.ScheduleViewType(viewType) {
	case entity.ScheduleViewTypeDay:
		start = timeUtil.BeginOfDayByTimeStamp().Unix()
		end = timeUtil.EndOfDayByTimeStamp().Unix()
	case entity.ScheduleViewTypeWorkweek:
		start, end = timeUtil.FindWorkWeekTimeRange()
	case entity.ScheduleViewTypeWeek:
		start, end = timeUtil.FindWeekTimeRange()
	case entity.ScheduleViewTypeMonth:
		start, end = timeUtil.FindMonthRange()
	case entity.ScheduleViewTypeYear:
		start = utils.StartOfYearByTimeStamp(timeAt, loc).Unix()
		end = utils.EndOfYearByTimeStamp(timeAt, loc).Unix()
	default:
		log.Info(ctx, "getScheduleTimeView:view_type is empty or invalid", log.String("view_type", viewType))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return nil, constant.ErrInvalidArgs
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
	condition := new(da.ScheduleCondition)
	condition.StartAndEndTimeViewRange = startAndEndTimeViewRange
	condition.SubjectIDs = entity.SplitStringToNullStrings(c.Query("subject_ids"))
	condition.ProgramIDs = entity.SplitStringToNullStrings(c.Query("program_ids"))
	schoolIDs := entity.SplitStringToNullStrings(c.Query("school_ids"))
	teacherIDs := entity.SplitStringToNullStrings(c.Query("teacher_ids"))

	filterClassIDs, err := model.GetSchedulePermissionModel().GetClassIDs(ctx, op)
	if err != nil {
		log.Error(ctx, "getScheduleTimeView:getClassIDs error",
			log.Err(err),
			log.Any("op", op),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return nil, err
	}
	if len(filterClassIDs) == 0 {
		log.Info(ctx, "getScheduleTimeView:filterClassIDs is empty", log.Any("operator", op))
		c.JSON(http.StatusOK, []*entity.ScheduleListView{})
		return nil, constant.ErrRecordNotFound
	}
	if schoolIDs.Valid {
		schoolClassIDs, err := s.GetClassIDsBySchoolIDs(ctx, op, schoolIDs.Strings)
		if err != nil {
			log.Error(ctx, "getScheduleTimeView:GetClassIDsBySchoolIDs error",
				log.Err(err),
				log.Any("op", op),
				log.Any("schoolIDs", schoolIDs),
			)
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return nil, err
		}
		filterClassIDs = utils.IntersectAndDeduplicateStrSlice(filterClassIDs, schoolClassIDs)
	}
	if teacherIDs.Valid {
		teacherClassIDs, err := model.GetScheduleModel().GetOrgClassIDsByUserIDs(ctx, op, teacherIDs.Strings, op.OrgID)
		if err != nil {
			log.Error(ctx, "getScheduleTimeView:GetScheduleModel.GetClassIDsBySchoolIDs error",
				log.Err(err),
				log.Any("op", op),
				log.Any("teacherIDs", teacherIDs),
			)
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return nil, err
		}
		filterClassIDs = utils.IntersectAndDeduplicateStrSlice(filterClassIDs, teacherClassIDs)
	}
	classIDs := entity.SplitStringToNullStrings(c.Query("class_ids"))
	if classIDs.Valid {
		filterClassIDs = utils.IntersectAndDeduplicateStrSlice(filterClassIDs, classIDs.Strings)
	}

	condition.ClassIDs = entity.NullStrings{
		Strings: filterClassIDs,
		Valid:   true,
	}

	log.Debug(ctx, "condition info",
		log.String("viewType", viewType),
		log.String("timeAtStr", timeAtStr),
		log.Any("condition", condition),
		log.Any("teacherIDs", teacherIDs),
		log.Any("classIDs", classIDs),
		log.Any("schoolIDs", schoolIDs),
	)
	return condition, nil
}

func (s *Server) GetClassIDsBySchoolIDs(ctx context.Context, op *entity.Operator, schoolIDs []string) ([]string, error) {
	schoolClassIDs := make([]string, 0)
	schoolClassInfos, err := external.GetClassServiceProvider().GetBySchoolIDs(ctx, op, schoolIDs)
	if err != nil {
		log.Error(ctx, "getScheduleTimeView:GetClassServiceProvider.GetBySchoolIDs error",
			log.Err(err),
			log.Any("op", op),
			log.Any("schoolIDs", schoolIDs),
		)
		return nil, err
	}
	for _, schoolClassInfo := range schoolClassInfos {
		for _, classInfo := range schoolClassInfo {
			schoolClassIDs = append(schoolClassIDs, classInfo.ID)
		}
	}
	return schoolClassIDs, nil
}

// @Summary updateStatus
// @ID updateStatus
// @Description update schedule status
// @Accept json
// @Produce json
// @Param schedule_id path string true "schedule id"
// @Param status query string true "schedule status" enums(NotStart, Started, Closed)
// @Tags schedule
// @Success 200 {object} entity.IDResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules/{schedule_id}/status [put]
func (s *Server) updateScheduleStatus(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status")
	ctx := c.Request.Context()
	scheduleStatus := entity.ScheduleStatus(status)
	if !scheduleStatus.Valid() {
		log.Warn(ctx, "schedule status error", log.String("id", id), log.String("status", status))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err := model.GetScheduleModel().UpdateScheduleStatus(ctx, dbo.MustGetDB(ctx), id, scheduleStatus)
	log.Info(ctx, "schedule status error", log.String("id", id), log.String("status", status))
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(ScheduleMsgEditOverlap))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, err.Error())
	}
}

// @Summary getParticipateClass
// @ID getParticipateClass
// @Description get participate Class
// @Accept json
// @Produce json
// @Tags schedule
// @Success 200 {array}  external.Class
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_participate/class [get]
func (s *Server) getParticipateClass(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	result, err := model.GetScheduleModel().GetParticipateClass(ctx, op)
	if err != nil {
		log.Error(ctx, "get participate  class error", log.Err(err), log.Any("op", op))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary get lessonPlans by teacher and class
// @Description get lessonPlans by teacher and class
// @Tags reports
// @ID getLessonPlans
// @Accept json
// @Produce json
// @Param teacher_id query string true "teacher id"
// @Param class_id query string true "class id"
// @Success 200 {array} entity.ScheduleShortInfo
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_lesson_plans [get]
func (s *Server) getLessonPlans(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	classID := c.Query("class_id")
	if len(strings.TrimSpace(classID)) == 0 {
		log.Info(ctx, "teacherID and classID is require",
			log.Any("operator", op),
		)
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	condition := &da.ScheduleCondition{
		StartLt: sql.NullInt64{
			Int64: time.Now().Add(constant.ScheduleAllowGoLiveTime).Unix(),
			Valid: true,
		},
		ClassID: sql.NullString{
			String: classID,
			Valid:  true,
		},
	}
	result, err := model.GetScheduleModel().GetLessonPlanByCondition(ctx, dbo.MustGetDB(ctx), op, condition)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get schedule real-time status
// @Description get schedule real-time status
// @Tags schedule
// @ID getScheduleRealTimeStatus
// @Accept json
// @Produce json
// @Param schedule_id path string true "schedule id"
// @Success 200 {object} entity.ScheduleRealTimeView
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules/{schedule_id}/real_time [get]
func (s Server) getScheduleRealTimeStatus(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	op := s.getOperator(c)
	result, err := model.GetScheduleModel().GetScheduleRealTimeStatus(ctx, op, id)
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(ScheduleMsgEditOverlap))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
