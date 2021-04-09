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
// @Success 200 {object} SuccessRequestResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
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
	log.Debug(ctx, "request data", log.Any("operator", op), log.Any("requestData", data))
	_, err := model.GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, op, []external.PermissionName{
		external.ScheduleCreateEvent,
		external.ScheduleCreateMySchoolEvent,
		external.ScheduleCreateMyEvent,
	})
	if err == constant.ErrForbidden {
		c.JSON(http.StatusForbidden, L(ScheduleMessageNoPermission))
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
	err = s.verifyScheduleData(c, &entity.ScheduleEditValidation{
		ClassRosterTeacherIDs:  data.ClassRosterTeacherIDs,
		ClassRosterStudentIDs:  data.ClassRosterStudentIDs,
		ParticipantsTeacherIDs: data.ParticipantsTeacherIDs,
		ParticipantsStudentIDs: data.ParticipantsStudentIDs,
		ClassID:                data.ClassID,
		ClassType:              data.ClassType,
		Title:                  data.Title,
	})
	if err != nil {
		log.Debug(ctx, "request data verify error", log.Err(err), log.Any("operator", op), log.Any("requestData", data))
		return
	}

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

	processResult, ok := s.processScheduleDueDate(c, &entity.ProcessScheduleDueAtInput{
		StartAt:   data.StartAt,
		EndAt:     data.EndAt,
		DueAt:     data.DueAt,
		ClassType: data.ClassType,
		Location:  loc,
	})
	if !ok {
		log.Info(ctx, "process schedule due date failure")
		return
	}
	data.StartAt = processResult.StartAt
	data.EndAt = processResult.EndAt
	data.DueAt = processResult.DueAt

	data.Location = loc
	if !data.IsForce &&
		(data.ClassType == entity.ScheduleClassTypeOnlineClass ||
			data.ClassType == entity.ScheduleClassTypeOfflineClass) {
		conflictInput := &entity.ScheduleConflictInput{
			ClassRosterTeacherIDs:  data.ClassRosterTeacherIDs,
			ClassRosterStudentIDs:  data.ClassRosterStudentIDs,
			ParticipantsTeacherIDs: data.ParticipantsTeacherIDs,
			ParticipantsStudentIDs: data.ParticipantsStudentIDs,
			StartAt:                data.StartAt,
			EndAt:                  data.EndAt,
			RepeatOptions:          data.Repeat,
			Location:               loc,
			IgnoreScheduleID:       data.ID,
			ClassID:                data.ClassID,
			IsRepeat:               data.IsRepeat && data.EditType == entity.ScheduleEditWithFollowing,
		}
		if data.IsRepeat && data.EditType == entity.ScheduleEditWithFollowing {
			conflictInput.IsRepeat = true
		}
		conflictData, err := model.GetScheduleModel().ConflictDetection(ctx, op, conflictInput)
		if err == constant.ErrConflict {
			c.JSON(http.StatusOK, LD(ScheduleMessageUsersConflict, conflictData))
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
	}
	log.Debug(ctx, "process request data", log.Any("operator", op), log.Any("requestData", data))
	newID, err := model.GetScheduleModel().Update(ctx, operator, &data)
	switch err {
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrConflict:
		c.JSON(http.StatusConflict, L(ScheduleMessageOverlap))
	case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(ScheduleMessageEditOverlap))
	case constant.ErrOperateNotAllowed:
		c.JSON(http.StatusBadRequest, L(ScheduleMessageEditOverlap))
	case model.ErrScheduleEditMissTime:
		c.JSON(http.StatusBadRequest, L(ScheduleMessageEditMissTime))
	case model.ErrScheduleLessonPlanUnAuthed:
		c.JSON(http.StatusBadRequest, L(ScheduleMessageLessonPlanInvalid))
	case model.ErrScheduleEditMissTimeForDueAt:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgEditMissDueDate))
	case model.ErrScheduleAlreadyHidden:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgHidden))
	case model.ErrScheduleAlreadyFeedback:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgAssignmentNew))
	case nil:
		c.JSON(http.StatusOK, D(IDResponse{ID: newID}))
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
// @Failure 403 {object} ForbiddenResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules/{schedule_id} [delete]
func (s *Server) deleteSchedule(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	err := model.GetSchedulePermissionModel().HasScheduleOrgPermission(ctx, op, external.ScheduleDeleteEvent)
	if err == constant.ErrForbidden {
		c.JSON(http.StatusForbidden, L(ScheduleMessageNoPermission))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}

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
		c.JSON(http.StatusBadRequest, L(ScheduleMessageEditOverlap))
	case model.ErrScheduleEditMissTime:
		c.JSON(http.StatusBadRequest, L(ScheduleMessageDeleteMissTime))
	case model.ErrScheduleEditMissTimeForDueAt:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgDeleteMissDueDate))
	case model.ErrScheduleAlreadyHidden:
		c.JSON(http.StatusBadRequest, L(ScheduleMsgHidden))
	case model.ErrScheduleAlreadyFeedback:
		c.JSON(http.StatusBadRequest, L(scheduleMsgHide))
	case nil:
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

func (s *Server) verifyScheduleData(c *gin.Context, input *entity.ScheduleEditValidation) error {
	op := s.getOperator(c)
	ctx := c.Request.Context()

	if strings.TrimSpace(input.Title) == "" {
		log.Info(ctx, "schedule title required", log.Any("input", input))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return constant.ErrInvalidArgs
	}

	// Students and teachers must exist
	if (len(input.ClassRosterTeacherIDs) == 0 &&
		len(input.ParticipantsTeacherIDs) == 0) ||
		(len(input.ClassRosterStudentIDs) == 0 &&
			len(input.ParticipantsStudentIDs) == 0) {
		log.Info(ctx, "add schedule: data is Invalid", log.Any("input", input), log.Any("op", op))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return constant.ErrInvalidArgs
	}
	if !input.ClassType.Valid() {
		log.Info(ctx, "add schedule: invalid class type", log.Any("input", input))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return constant.ErrInvalidArgs
	}
	// if classID is not empty,verify has permission
	if input.ClassID != "" {
		//if len(input.ClassRosterTeacherIDs) == 0 && len(input.ClassRosterStudentIDs) == 0 {
		//	log.Info(ctx, "add schedule: classRoster data is Invalid", log.Any("data", input))
		//	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		//	return constant.ErrInvalidArgs
		//}
		// has permission to access the class
		err := model.GetSchedulePermissionModel().HasClassesPermission(ctx, op, []string{input.ClassID})
		if err == constant.ErrForbidden {
			c.JSON(http.StatusForbidden, L(ScheduleMessageNoPermission))
			return constant.ErrForbidden
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return err
		}
	}
	return nil
}

// @Summary addSchedule
// @ID addSchedule
// @Description add a schedule data
// @Accept json
// @Produce json
// @Param scheduleData body entity.ScheduleAddView true "schedule data to add"
// @Tags schedule
// @Success 200 {object} SuccessRequestResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
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
	log.Debug(ctx, "request data", log.Any("operator", op), log.Any("requestData", data))
	err := s.verifyScheduleData(c, &entity.ScheduleEditValidation{
		ClassRosterTeacherIDs:  data.ClassRosterTeacherIDs,
		ClassRosterStudentIDs:  data.ClassRosterStudentIDs,
		ParticipantsTeacherIDs: data.ParticipantsTeacherIDs,
		ParticipantsStudentIDs: data.ParticipantsStudentIDs,
		ClassID:                data.ClassID,
		ClassType:              data.ClassType,
		Title:                  data.Title,
	})
	if err != nil {
		log.Debug(ctx, "request data verify error", log.Err(err), log.Any("operator", op), log.Any("requestData", data))
		return
	}
	_, err = model.GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, op, []external.PermissionName{
		external.ScheduleCreateEvent,
		external.ScheduleCreateMySchoolEvent,
		external.ScheduleCreateMyEvent,
	})
	if err == constant.ErrForbidden {
		c.JSON(http.StatusForbidden, L(ScheduleMessageNoPermission))
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
	processResult, ok := s.processScheduleDueDate(c, &entity.ProcessScheduleDueAtInput{
		StartAt:   data.StartAt,
		EndAt:     data.EndAt,
		DueAt:     data.DueAt,
		ClassType: data.ClassType,
		Location:  loc,
	})
	if !ok {
		log.Info(ctx, "process schedule due date failure")
		return
	}
	data.StartAt = processResult.StartAt
	data.EndAt = processResult.EndAt
	data.DueAt = processResult.DueAt

	// add schedule
	data.Location = loc
	if !data.IsForce &&
		(data.ClassType == entity.ScheduleClassTypeOnlineClass ||
			data.ClassType == entity.ScheduleClassTypeOfflineClass) {
		conflictInput := &entity.ScheduleConflictInput{
			ClassRosterTeacherIDs:  data.ClassRosterTeacherIDs,
			ClassRosterStudentIDs:  data.ClassRosterStudentIDs,
			ParticipantsTeacherIDs: data.ParticipantsTeacherIDs,
			ParticipantsStudentIDs: data.ParticipantsStudentIDs,
			StartAt:                data.StartAt,
			EndAt:                  data.EndAt,
			RepeatOptions:          data.Repeat,
			Location:               loc,
			IsRepeat:               data.IsRepeat,
			ClassID:                data.ClassID,
		}
		conflictData, err := model.GetScheduleModel().ConflictDetection(ctx, op, conflictInput)
		if err == constant.ErrConflict {
			c.JSON(http.StatusOK, LD(ScheduleMessageUsersConflict, conflictData))
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return
		}
	}
	log.Debug(ctx, "process request data", log.Any("operator", op), log.Any("requestData", data))
	id, err := model.GetScheduleModel().Add(ctx, op, data)
	switch err {
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrFileNotFound:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrScheduleLessonPlanUnAuthed:
		c.JSON(http.StatusBadRequest, L(ScheduleMessageLessonPlanInvalid))
	case nil:
		c.JSON(http.StatusOK, D(IDResponse{ID: id}))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

func (s *Server) processScheduleDueDate(c *gin.Context, input *entity.ProcessScheduleDueAtInput) (*entity.ProcessScheduleDueAtView, bool) {
	now := time.Now().Unix()
	ctx := c.Request.Context()
	var day int64
	var result = new(entity.ProcessScheduleDueAtView)
	switch input.ClassType {
	case entity.ScheduleClassTypeTask:
		result.StartAt = input.StartAt
		result.EndAt = input.EndAt
		if input.DueAt <= 0 {
			result.DueAt = 0
			return result, true
		}
		day = utils.GetTimeDiffToDayByTimeStamp(input.EndAt, input.DueAt, input.Location)
		if day < 0 {
			log.Info(ctx, "schedule dueAt is invalid", log.Int64("now", now), log.Any("input", input))
			c.JSON(http.StatusBadRequest, L(ScheduleMessageDueDateEarlierEndDate))
			return nil, false
		}
		result.DueAt = utils.TodayEndByTimeStamp(input.DueAt, input.Location).Unix()
	case entity.ScheduleClassTypeHomework:
		if input.DueAt <= 0 {
			result.DueAt = 0
			return result, true
		}
		day = utils.GetTimeDiffToDayByTimeStamp(now, input.DueAt, input.Location)
		if day < 0 {
			log.Info(ctx, "schedule dueAt is invalid", log.Int64("now", now), log.Any("input", input))
			c.JSON(http.StatusBadRequest, L(ScheduleMessageDueDateEarlierToDay))
			return nil, false
		}
		result.DueAt = utils.TodayEndByTimeStamp(input.DueAt, input.Location).Unix()
	default:
		result.StartAt = input.StartAt
		result.EndAt = input.EndAt
		result.DueAt = 0
	}
	return result, true
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
		c.JSON(http.StatusNotFound, L(ScheduleMessageDeleteOverlap))
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
		startAt = utils.TodayZeroByTimeStamp(startAt, loc).Unix()
		condition.StartAtGe = sql.NullInt64{
			Int64: startAt,
			Valid: startAt > 0,
		}
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

		log.Debug(ctx, "querySchedule:debug",
			log.Strings("teacherIDs", teacherIDs),
			log.Any("operator", op),
		)
		condition.RelationIDs = entity.NullStrings{
			Strings: teacherIDs,
			Valid:   len(teacherIDs) > 0,
		}
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
// @Param view_type query string true "search schedules by view_type" enums(day, work_week, week, month,year,full_view)
// @Param time_at query integer false "search schedules by time_at"
// @Param time_zone_offset query integer false "time zone offset"
// @Param school_ids query string false "school id,separated by comma"
// @Param teacher_ids query string false "teacher id,separated by comma"
// @Param class_ids query string false "class id,separated by comma,special classes id is 'Undefined',this class members only under org"
// @Param subject_ids query string false "subject id,separated by comma"
// @Param program_ids query string false "program id,separated by comma"
// @Param class_types query string false "class type,separated by comma"
// @Param due_at_eq query integer false "get schedules equal to due_at"
// @Param start_at_ge query integer false "get schedules greater than or equal to start_at"
// @Param end_at_le query integer false "get schedules less than or equal to end_at"
// @Param filter_option query string false "get schedules by filter option" enums(any_time,only_mine)
// @Param order_by query string false "order by" enums(create_at, -create_at, start_at, -start_at)
// @Tags schedule
// @Success 200 {object} entity.ScheduleListView
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_time_view [get]
func (s *Server) getScheduleTimeView(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	offsetStr := c.Query("time_zone_offset")
	offset, _ := strconv.Atoi(offsetStr)
	loc := utils.GetTimeLocationByOffset(offset)
	log.Info(ctx, "getScheduleTimeView: time_zone_offset", log.String("time_zone_offset", offsetStr), log.Any("loc", loc))
	condition, err := s.getScheduleTimeViewCondition(c, loc)
	if err != nil {
		return
	}
	result, err := model.GetScheduleModel().Query(ctx, op, condition, loc)
	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}
	if err == constant.ErrRecordNotFound {
		log.Info(ctx, "record not found", log.Any("condition", condition))
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
		return
	}
	log.Debug(ctx, "getScheduleTimeView error", log.Err(err), log.Any("condition", condition), log.Any("condition", condition), log.String("offsetStr", offsetStr))
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
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_time_view/dates [get]
func (s *Server) getScheduledDates(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	offsetStr := c.Query("time_zone_offset")
	offset, _ := strconv.Atoi(offsetStr)
	loc := utils.GetTimeLocationByOffset(offset)
	log.Info(ctx, "getScheduleTimeView: time_zone_offset", log.String("time_zone_offset", offsetStr), log.Any("loc", loc))

	condition, err := s.getScheduleTimeViewCondition(c, loc)
	if err != nil {
		return
	}
	result, err := model.GetScheduleModel().QueryScheduledDates(ctx, op, condition, loc)
	if err != nil {
		log.Error(ctx, "getScheduledDates:GetScheduleModel.QueryScheduledDates error", log.Err(err), log.Any("condition", condition))
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) getScheduleTimeViewCondition(c *gin.Context, loc *time.Location) (*da.ScheduleCondition, error) {
	op := s.getOperator(c)
	ctx := c.Request.Context()

	permissionMap, err := model.GetSchedulePermissionModel().HasScheduleOrgPermissions(ctx, op, []external.PermissionName{
		external.ScheduleViewOrgCalendar,
		external.ScheduleViewSchoolCalendar,
		external.ScheduleViewMyCalendar,
	})
	if err == constant.ErrForbidden {
		c.JSON(http.StatusForbidden, L(ScheduleMessageNoPermission))
		return nil, constant.ErrForbidden
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return nil, constant.ErrInternalServer
	}

	viewType := c.Query("view_type")
	condition := new(da.ScheduleCondition)
	if viewType != entity.ScheduleViewTypeFullView.String() {
		timeAtStr := c.Query("time_at")
		timeAt, err := strconv.ParseInt(timeAtStr, 10, 64)
		if err != nil {
			log.Info(ctx, "getScheduleTimeView: time_at is empty or invalid", log.String("time_at", timeAtStr))
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return nil, err
		}
		var (
			start int64
			end   int64
		)
		switch entity.ScheduleViewType(viewType) {
		case entity.ScheduleViewTypeDay:
			start = utils.TodayZeroByTimeStamp(timeAt, loc).Unix()
			end = utils.TodayEndByTimeStamp(timeAt, loc).Unix()
		case entity.ScheduleViewTypeWorkweek:
			start, end = utils.FindWorkWeekTimeRange(timeAt, loc)
		case entity.ScheduleViewTypeWeek:
			start, end = utils.FindWeekTimeRange(timeAt, loc)
		case entity.ScheduleViewTypeMonth:
			start, end = utils.FindMonthRange(timeAt, loc)
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
			Valid: start >= 0,
			Int64: start,
		}
		startAndEndTimeViewRange[1] = sql.NullInt64{
			Valid: end >= 0,
			Int64: end,
		}
		condition.StartAndEndTimeViewRange = startAndEndTimeViewRange
	}

	relationIDs := make([]string, 0)
	condition.SubjectIDs = entity.SplitStringToNullStrings(c.Query("subject_ids"))
	condition.ProgramIDs = entity.SplitStringToNullStrings(c.Query("program_ids"))
	condition.ClassTypes = entity.SplitStringToNullStrings(c.Query("class_types"))
	condition.OrderBy = da.NewScheduleOrderBy(c.Query("order_by"))
	err = s.processTimeQuery(c, condition)
	if err != nil {
		return nil, err
	}
	condition.OrgID = sql.NullString{
		String: op.OrgID,
		Valid:  true,
	}
	schoolIDs := entity.SplitStringToNullStrings(c.Query("school_ids"))
	classIDs := entity.SplitStringToNullStrings(c.Query("class_ids"))
	relationIDs = append(relationIDs, schoolIDs.Strings...)
	relationIDs = append(relationIDs, classIDs.Strings...)
	hasUndefinedClass := false
	for _, classID := range classIDs.Strings {
		if classID == entity.ScheduleFilterUndefinedClass {
			hasUndefinedClass = true
			break
		}
	}
	if hasUndefinedClass {
		userInfo, err := model.GetSchedulePermissionModel().GetOnlyUnderOrgUsers(ctx, op)
		if err != nil {
			log.Error(ctx, "GetSchedulePermissionModel.GetOnlyUnderOrgUsers error",
				log.Err(err),
				log.Any("op", op),
			)
			c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
			return nil, constant.ErrInternalServer
		}
		for _, item := range userInfo {
			relationIDs = append(relationIDs, item.ID)
		}
	}

	if permissionMap[external.ScheduleViewOrgCalendar] {
		condition.RelationIDs = entity.NullStrings{
			Strings: relationIDs,
			Valid:   len(relationIDs) > 0,
		}
	} else if permissionMap[external.ScheduleViewSchoolCalendar] {
		if len(relationIDs) == 0 {
			schoolList, err := external.GetSchoolServiceProvider().GetByPermission(ctx, op, external.ScheduleViewSchoolCalendar)
			if err != nil {
				log.Error(ctx, "GetSchoolServiceProvider.GetByPermission error",
					log.Err(err),
					log.Any("op", op),
					log.String("permission", external.ScheduleViewSchoolCalendar.String()),
				)
				c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
				return nil, constant.ErrInternalServer
			}
			for _, item := range schoolList {
				relationIDs = append(relationIDs, item.ID)
			}
		}
		condition.RelationIDs = entity.NullStrings{
			Strings: relationIDs,
			Valid:   true,
		}
	} else if permissionMap[external.ScheduleViewMyCalendar] {
		condition.RelationID = sql.NullString{
			String: op.UserID,
			Valid:  true,
		}
		condition.RelationIDs = entity.NullStrings{
			Strings: relationIDs,
			Valid:   true,
		}
	}
	filterOption := c.Query("filter_option")
	switch entity.ScheduleFilterOption(filterOption) {
	case entity.ScheduleFilterAnyTime:
		condition.AnyTime = sql.NullBool{
			Bool:  true,
			Valid: true,
		}
	case entity.ScheduleFilterOnlyMine:
		condition.RelationIDs = entity.NullStrings{
			Strings: []string{op.UserID},
			Valid:   true,
		}
		condition.RelationSchoolIDs = schoolIDs
	}
	log.Debug(ctx, "condition info",
		log.String("viewType", viewType),
		log.Any("condition", condition),
	)
	return condition, nil
}

func (s *Server) processTimeQuery(c *gin.Context, condition *da.ScheduleCondition) error {
	ctx := c.Request.Context()
	dueAtStr := c.Query("due_at_eq")
	if dueAtStr != "" {
		dueAt, err := strconv.ParseInt(dueAtStr, 10, 64)
		if err != nil {
			log.Info(ctx, "getScheduleTimeView: time_at is empty or invalid", log.String("dueAt", dueAtStr))
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return err
		}
		condition.DueToEq = sql.NullInt64{
			Int64: dueAt,
			Valid: true,
		}
	}
	startAtGeStr := c.Query("start_at_ge")
	if startAtGeStr != "" {
		startAt, err := strconv.ParseInt(startAtGeStr, 10, 64)
		if err != nil {
			log.Info(ctx, "getScheduleTimeView: start_at_ge is empty or invalid", log.String("startAtGeStr", startAtGeStr))
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return err
		}
		condition.StartAtGe = sql.NullInt64{
			Int64: startAt,
			Valid: true,
		}
	}

	endAtLeStr := c.Query("end_at_le")
	if endAtLeStr != "" {
		endAt, err := strconv.ParseInt(endAtLeStr, 10, 64)
		if err != nil {
			log.Info(ctx, "getScheduleTimeView: end_at_le is empty or invalid", log.String("endAtLeStr", endAtLeStr))
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
			return err
		}
		condition.EndAtLe = sql.NullInt64{
			Int64: endAt,
			Valid: true,
		}
	}
	return nil
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
// @Success 200 {object} IDResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules/{schedule_id}/status [put]
func (s *Server) updateScheduleStatus(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status")
	ctx := c.Request.Context()
	op := s.getOperator(c)
	scheduleStatus := entity.ScheduleStatus(status)
	if !scheduleStatus.Valid() {
		log.Warn(ctx, "schedule status error", log.String("id", id), log.String("status", status))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err := model.GetScheduleModel().UpdateScheduleStatus(ctx, dbo.MustGetDB(ctx), op, id, scheduleStatus)
	log.Info(ctx, "schedule status error", log.String("id", id), log.String("status", status))
	switch err {
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(ScheduleMessageEditOverlap))
	case nil:
		c.JSON(http.StatusOK, IDResponse{ID: id})
	default:
		c.JSON(http.StatusInternalServerError, err.Error())
	}
}

// @Summary get lessonPlans by teacher and class
// @Description get lessonPlans by teacher and class
// @Tags schedule
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
		StartAtLt: sql.NullInt64{
			Int64: time.Now().Add(constant.ScheduleAllowGoLiveTime).Unix(),
			Valid: true,
		},
		RelationID: sql.NullString{
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
		c.JSON(http.StatusNotFound, L(ScheduleMessageEditOverlap))
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get schedule filter schools
// @Description get get schedule filter schools
// @Tags schedule
// @ID getScheduleFilterSchool
// @Accept json
// @Produce json
// @Success 200 {array} entity.ScheduleFilterSchool
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_filter/schools [get]
func (s Server) getSchoolInScheduleFilter(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	result, err := model.GetSchedulePermissionModel().GetSchoolsByOperator(ctx, op)
	switch err {
	case constant.ErrForbidden:
		c.JSON(http.StatusOK, []*entity.ScheduleFilterSchool{})
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get schedule filter classes
// @Description get schedule filter classes
// @Tags schedule
// @ID getScheduleFilterClasses
// @Accept json
// @Produce json
// @Param school_id query string false "school id,if "-1",return classes without school"
// @Success 200 {array} entity.ScheduleFilterClass
// @Failure 400 {object} BadRequestResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_filter/classes [get]
func (s Server) getClassesInScheduleFilter(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	schoolID := c.Query("school_id")
	result, err := model.GetSchedulePermissionModel().GetClassesByOperator(ctx, op, schoolID)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrForbidden, constant.ErrRecordNotFound:
		c.JSON(http.StatusOK, []*entity.ScheduleFilterClass{})
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary updateScheduleShowOption
// @ID updateScheduleShowOption
// @Description update schedule show option
// @Accept json
// @Produce json
// @Param schedule_id path string true "schedule id"
// @Param show_option query string false "hidden properties" enums(hidden,visible)
// @Tags schedule
// @Success 200 {object} IDResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules/{schedule_id}/show_option [put]
func (s *Server) updateScheduleShowOption(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	id := c.Param("id")
	option := c.Query("show_option")
	id, err := model.GetScheduleModel().UpdateScheduleShowOption(ctx, op, id, entity.ScheduleShowOption(option))
	switch err {
	case nil:
		c.JSON(http.StatusOK, IDResponse{ID: id})
	case constant.ErrForbidden:
		c.JSON(http.StatusForbidden, L(ScheduleMessageNoPermission))
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(ScheduleMessageEditOverlap))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary getScheduleNewestFeedbackByOperator
// @ID getScheduleNewestFeedbackByOperator
// @Description get schedule newest feedback by operator
// @Accept json
// @Produce json
// @Param schedule_id path string true "schedule id"
// @Tags schedule
// @Success 200 {object} entity.ScheduleFeedbackView
// @Failure 404 {object} NotFoundResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules/{schedule_id}/operator/newest_feedback [get]
func (s *Server) getScheduleNewestFeedbackByOperator(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	scheduleID := c.Param("id")

	result, err := model.GetScheduleFeedbackModel().GetNewest(ctx, op, op.UserID, scheduleID)

	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case model.ErrFeedbackNotGenerateAssessment:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	case constant.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get schedule filter programs
// @Description get schedule filter programs
// @Tags schedule
// @ID getProgramsInScheduleFilter
// @Accept json
// @Produce json
// @Success 200 {array} entity.ScheduleShortInfo
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_filter/programs [get]
func (s *Server) getProgramsInScheduleFilter(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()

	programs, err := model.GetScheduleModel().GetPrograms(ctx, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, programs)
	case constant.ErrForbidden:
		c.JSON(http.StatusOK, []*entity.ScheduleShortInfo{})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get schedule filter subjects
// @Description get schedule filter subjects
// @Tags schedule
// @ID getSubjectsInScheduleFilter
// @Accept json
// @Produce json
// @Param program_id query string true "program id"
// @Success 200 {array} entity.ScheduleShortInfo
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_filter/subjects [get]
func (s *Server) getSubjectsInScheduleFilter(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()
	programID := c.Query("program_id")

	subjects, err := model.GetScheduleModel().GetSubjects(ctx, op, programID)
	switch err {
	case nil:
		c.JSON(http.StatusOK, subjects)
	case constant.ErrForbidden:
		c.JSON(http.StatusOK, []*entity.ScheduleShortInfo{})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary get schedule filter classTypes
// @Description get schedule filter classTypes
// @Tags schedule
// @ID getClassTypesInScheduleFilter
// @Accept json
// @Produce json
// @Success 200 {array} entity.ScheduleShortInfo
// @Failure 500 {object} InternalServerErrorResponse
// @Router /schedules_filter/class_types [get]
func (s *Server) getClassTypesInScheduleFilter(c *gin.Context) {
	op := s.getOperator(c)
	ctx := c.Request.Context()

	classTypes, err := model.GetScheduleModel().GetClassTypes(ctx, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, classTypes)
	case constant.ErrForbidden:
		c.JSON(http.StatusOK, []string{})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
