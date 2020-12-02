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
	if !s.hasScheduleRWPermission(c, op, external.ScheduleEditEvent) {
		return
	}
	ctx := c.Request.Context()
	id := c.Param("id")
	data := entity.ScheduleUpdateView{}
	if err := c.ShouldBind(&data); err != nil {
		log.Info(ctx, "update schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
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
	if data.StartAt < now || data.StartAt >= data.EndAt {
		log.Info(ctx, "schedule start_at or end_at is invalid",
			log.Int64("StartAt", data.StartAt),
			log.Int64("EndAt", data.EndAt),
			log.Int64("now", now))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
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
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(ScheduleMsgOverlap))
	case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(ScheduleMsgEditOverlap))
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgEditOverlap))
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
	if !s.hasScheduleRWPermission(c, op, external.ScheduleDeleteEvent) {
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

	err := model.GetScheduleModel().Delete(ctx, op, id, editType)
	switch err {
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case dbo.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgEditOverlap))
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
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules [post]
func (s *Server) addSchedule(c *gin.Context) {
	op := s.getOperator(c)
	if !s.hasScheduleRWPermission(c, op, external.ScheduleCreateEvent) {
		return
	}
	ctx := c.Request.Context()
	data := new(entity.ScheduleAddView)
	if err := c.ShouldBind(data); err != nil {
		log.Info(ctx, "add schedule: should bind body failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
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
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
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
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(ScheduleMsgOverlap))
	case constant.ErrFileNotFound:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, entity.IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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

	filterClassIDs, err := s.getClassIDs(ctx, op)
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
	op := s.getOperator(c)
	ctx := c.Request.Context()
	viewType := c.Query("view_type")
	timeAtStr := c.Query("time_at")
	timeAt, err := strconv.ParseInt(timeAtStr, 10, 64)
	if err != nil {
		log.Info(ctx, "getScheduleTimeView: time_at is empty or invalid", log.String("time_at", timeAtStr))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	offsetStr := c.Query("time_zone_offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		log.Info(ctx, "getScheduleTimeView: time_zone_offset invalid", log.String("time_zone_offset", offsetStr))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	loc := utils.GetTimeLocationByOffset(offset)
	log.Debug(ctx, "time location", log.Any("op", op), log.Any("location", loc), log.Int("offset", offset))
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
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
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
	condition := new(da.ScheduleCondition)
	condition.StartAndEndTimeViewRange = startAndEndTimeViewRange
	condition.SubjectIDs = entity.SplitStringToNullStrings(c.Query("subject_ids"))
	condition.ProgramIDs = entity.SplitStringToNullStrings(c.Query("program_ids"))
	schoolIDs := entity.SplitStringToNullStrings(c.Query("school_ids"))
	teacherIDs := entity.SplitStringToNullStrings(c.Query("teacher_ids"))

	filterClassIDs, err := s.getClassIDs(ctx, op)
	if err != nil {
		log.Error(ctx, "getScheduleTimeView:getClassIDs error",
			log.Err(err),
			log.Any("op", op),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if len(filterClassIDs) == 0 {
		log.Info(ctx, "getScheduleTimeView:filterClassIDs is empty", log.Any("operator", op))
		c.JSON(http.StatusOK, []*entity.ScheduleListView{})
		return
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
			return
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
			return
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
	result, err := model.GetScheduleModel().Query(ctx, condition)
	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}
	if err == constant.ErrRecordNotFound {
		log.Info(ctx, "record not found", log.String("viewType", viewType), log.String("timeAtStr", timeAtStr), log.Any("condition", condition))
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
		return
	}
	log.Info(ctx, "record not found", log.Err(err), log.String("viewType", viewType), log.String("timeAtStr", timeAtStr), log.Any("condition", condition))
	c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
		Status: sql.NullString{
			String: string(entity.ScheduleStatusClosed),
			Valid:  true,
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

func (s *Server) getClassIDs(ctx context.Context, op *entity.Operator) ([]string, error) {
	schoolClassIDs, err := s.getClassIDsBySchoolPermission(ctx, op, external.ScheduleViewSchoolCalendar)
	if err != nil {
		log.Error(ctx, "getClassIDsByPermission:getClassIDsBySchoolPermission error",
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}

	orgClassIDs, err := s.getClassIDsByOrgPermission(ctx, op, external.ScheduleViewOrgCalendar)
	if err != nil {
		log.Error(ctx, "getClassIDsByPermission:getClassIDsByOrgPermission error",
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}

	myClassIDs := make([]string, 0)
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ScheduleViewMyCalendar)
	if err != nil {
		log.Error(ctx, "getScheduleTimeView:GetPermissionServiceProvider.HasOrganizationPermission error",
			log.Err(err),
			log.String("PermissionName", external.ScheduleViewMyCalendar.String()),
			log.Any("op", op),
		)
		return nil, err
	}
	if hasPermission {
		myClassIDs, err = model.GetScheduleModel().GetOrgClassIDsByUserIDs(ctx, op, []string{op.UserID}, op.OrgID)
		if err != nil {
			log.Error(ctx, "getScheduleTimeView:GetScheduleModel.GetMyOrgClassIDs error",
				log.Err(err),
				log.Any("op", op),
			)
			return nil, err
		}
	}

	log.Info(ctx, "getClassIDs", log.Any("Operator", op),
		log.Strings("schoolClassIDs", schoolClassIDs),
		log.Strings("orgClassIDs", orgClassIDs),
		log.Strings("myClassIDs", myClassIDs),
	)
	classIDs := make([]string, 0, len(schoolClassIDs)+len(orgClassIDs)+len(myClassIDs))
	classIDs = append(classIDs, schoolClassIDs...)
	classIDs = append(classIDs, orgClassIDs...)
	classIDs = append(classIDs, myClassIDs...)

	classIDs = utils.SliceDeduplication(classIDs)

	return classIDs, nil
}

func (s *Server) getClassIDsBySchoolPermission(ctx context.Context, op *entity.Operator, permissionName external.PermissionName) ([]string, error) {
	classIDs := make([]string, 0)
	schoolInfoList, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, permissionName)
	if err != nil {
		log.Error(ctx, "check permission error",
			log.String("permission", permissionName.String()),
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}
	schoolIDs := make([]string, len(schoolInfoList))
	for i, item := range schoolInfoList {
		schoolIDs[i] = item.ID
	}
	classMap, err := external.GetClassServiceProvider().GetBySchoolIDs(ctx, op, schoolIDs)
	if err != nil {
		log.Error(ctx, "getClassIDsBySchoolPermission:GetClassServiceProvider GetBySchoolIDs error",
			log.String("permission", permissionName.String()),
			log.Strings("schoolIDs", schoolIDs),
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}
	if classList, ok := classMap[op.OrgID]; ok {
		for _, item := range classList {
			classIDs = append(classIDs, item.ID)
		}
	}
	return classIDs, nil
}

func (s *Server) getClassIDsByOrgPermission(ctx context.Context, op *entity.Operator, permissionName external.PermissionName) ([]string, error) {
	//external.ScheduleViewOrgCalendar
	orgInfoList, err := external.GetOrganizationServiceProvider().GetByPermission(ctx, op, permissionName)
	if err != nil {
		log.Error(ctx, "getClassIDsByOrgPermission：check permission error",
			log.String("permission", permissionName.String()),
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}
	orgIDs := make([]string, len(orgInfoList))
	for i, item := range orgInfoList {
		orgIDs[i] = item.ID
	}
	log.Info(ctx, "getClassIDsByOrgPermission", log.Any("orgInfoList", orgInfoList))
	classMap, err := external.GetClassServiceProvider().GetByOrganizationIDs(ctx, op, orgIDs)
	if err != nil {
		log.Error(ctx, "getClassIDsByOrgPermission:GetClassServiceProvider GetByOrganizationIDs error",
			log.String("permission", permissionName.String()),
			log.Strings("orgIDs", orgIDs),
			log.Any("operator", op),
			log.Err(err),
		)
		return nil, err
	}
	log.Info(ctx, "getClassIDsByOrgPermission", log.Any("op", op), log.Any("classMap", classMap), log.Any("classMap[op.OrgID]", classMap[op.OrgID]))
	var classIDs []string
	if classList, ok := classMap[op.OrgID]; ok {
		log.Info(ctx, "getClassIDsByOrgPermission", log.Any("classList", classList))
		classIDs = make([]string, len(classList))
		for i, item := range classList {
			classIDs[i] = item.ID
		}
	}
	log.Info(ctx, "getClassIDsByOrgPermission", log.Strings("classIDs", classIDs), log.String("permissionName", permissionName.String()))
	return classIDs, nil
}

func (s *Server) hasScheduleRWPermission(c *gin.Context, op *entity.Operator, permissionName external.PermissionName) bool {
	ctx := c.Request.Context()
	hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permissionName)
	if err != nil {
		log.Error(ctx, "check permission error",
			log.String("permission", string(permissionName)),
			log.Any("operator", op),
			log.Err(err),
		)
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return false
	}
	if !hasPermission {
		log.Info(ctx, "no permission",
			log.String("permission", string(permissionName)),
			log.Any("Operator", op),
		)
		c.JSON(http.StatusForbidden, L(ScheduleMsgNoPermission))
		return false
	}
	return true
}

//func (s Server) getScheduleReadPermission(c *gin.Context, op *entity.Operator) map[external.PermissionName]bool {
//	ctx := c.Request.Context()
//	result := make(map[external.PermissionName]bool)
//	viewOrg, err := external.GetPermissionServiceProvider().HasPermission(ctx, op, external.ScheduleViewOrgCalendar)
//	if err != nil {
//		log.Error(ctx, "check permission error",
//			log.String("permission", string(external.ScheduleViewOrgCalendar)),
//			log.Any("operator", op),
//			log.Err(err),
//		)
//		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
//		return result
//	}
//	viewMy, err := external.GetPermissionServiceProvider().HasPermission(ctx, op, external.ScheduleViewMyCalendar)
//	if err != nil {
//		log.Error(ctx, "check permission error",
//			log.String("permission", string(external.ScheduleViewMyCalendar)),
//			log.Any("operator", op),
//			log.Err(err),
//		)
//		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
//		return result
//	}
//	if viewOrg || viewMy {
//		result[external.ScheduleViewOrgCalendar] = viewOrg
//		result[external.ScheduleViewMyCalendar] = viewMy
//		return result
//	}
//	c.JSON(http.StatusForbidden, L(ScheduleMsgNoPermission))
//	return result
//}
