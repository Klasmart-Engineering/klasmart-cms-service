package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IScheduleModel interface {
	Add(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error)
	AddTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error)
	Update(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleUpdateView) (string, error)
	Delete(ctx context.Context, op *entity.Operator, id string, editType entity.ScheduleEditType) error
	//DeleteTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, id string, editType entity.ScheduleEditType) error
	Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.ScheduleListView, error)
	Page(ctx context.Context, condition *da.ScheduleCondition) (int, []*entity.ScheduleSearchView, error)
	GetByID(ctx context.Context, id string) (*entity.ScheduleDetailsView, error)
	IsScheduleConflict(ctx context.Context, op *entity.Operator, startAt int64, endAt int64) (bool, error)
	GetTeacherByName(ctx context.Context, name string) ([]*external.Teacher, error)
	ExistScheduleAttachmentFile(ctx context.Context, attachmentPath string) bool
	ExistScheduleByLessonPlanID(ctx context.Context, lessonPlanID string) (bool, error)
	ExistScheduleByID(ctx context.Context, id string) (bool, error)
	GetPlainByID(ctx context.Context, id string) (*entity.SchedulePlain, error)
	UpdateScheduleStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.ScheduleStatus) error
	GetParticipateClass(ctx context.Context, operator *entity.Operator) ([]*external.Class, error)
}
type scheduleModel struct {
	testScheduleRepeatFlag bool
}

func (s *scheduleModel) IsScheduleConflict(ctx context.Context, op *entity.Operator, startAt int64, endAt int64) (bool, error) {
	var scheduleList []*entity.Schedule
	StartAndEndRange := make([]sql.NullInt64, 2)
	StartAndEndRange[0] = sql.NullInt64{
		Valid: startAt <= 0,
		Int64: startAt,
	}
	StartAndEndRange[1] = sql.NullInt64{
		Valid: endAt <= 0,
		Int64: endAt,
	}
	err := da.GetScheduleDA().Query(ctx, &da.ScheduleCondition{
		OrgID: sql.NullString{
			String: op.OrgID,
			Valid:  op.OrgID != "",
		},
		StartAndEndRange: StartAndEndRange,
	}, &scheduleList)
	if err != nil {
		return false, err
	}
	if len(scheduleList) > 0 {
		log.Debug(ctx, "conflict schedule data", log.Any("scheduleList", scheduleList))
		return true, nil
	}
	return false, nil
}

func (s *scheduleModel) ExistScheduleAttachmentFile(ctx context.Context, attachmentPath string) bool {
	_, exist := storage.DefaultStorage().ExistFile(ctx, storage.ScheduleAttachmentStoragePartition, attachmentPath)
	if !exist {
		log.Info(ctx, "add schedule: attachment is not exits", log.Any("attachmentPath", attachmentPath))
		return false
	}
	return true
}

func (s *scheduleModel) Add(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error) {
	id, err := dbo.GetTransResult(ctx, func(ctx context.Context, tx *dbo.DBContext) (interface{}, error) {
		return s.AddTx(ctx, tx, op, viewData)
	})
	if err != nil {
		log.Error(ctx, "add schedule error",
			log.Err(err),
			log.Any("viewData", viewData),
		)
		return "", err
	}
	da.GetScheduleRedisDA().Clean(ctx, nil)
	return id.(string), nil
}
func (s *scheduleModel) AddTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error) {
	// verify data
	err := s.verifyData(ctx, &entity.ScheduleVerify{
		ClassID:      viewData.ClassID,
		SubjectID:    viewData.SubjectID,
		ProgramID:    viewData.ProgramID,
		TeacherIDs:   viewData.TeacherIDs,
		LessonPlanID: viewData.LessonPlanID,
		ClassType:    viewData.ClassType,
	})
	if err != nil {
		log.Error(ctx, "add schedule: verify data error",
			log.Err(err),
			log.Any("viewData", viewData))
		return "", constant.ErrInvalidArgs
	}

	// not force add need conflict detection
	if !viewData.IsForce {
		conflict, err := GetScheduleModel().IsScheduleConflict(ctx, op, viewData.StartAt, viewData.EndAt)
		if err != nil {
			log.Error(ctx, "add schedule: check conflict failed",
				log.Int64("start_at", viewData.StartAt),
				log.Int64("end_at", viewData.EndAt),
			)
			return "", err
		}
		if conflict {
			log.Warn(ctx, "add schedule: time conflict",
				log.Int64("start_at", viewData.StartAt),
				log.Int64("end_at", viewData.EndAt),
			)
			return "", constant.ErrConflict
		}
	}
	schedule, err := viewData.ToSchedule(ctx)
	schedule.CreatedID = op.UserID
	scheduleID, err := s.addSchedule(ctx, tx, schedule, viewData.TeacherIDs, &viewData.Repeat, viewData.Location)
	if err != nil {
		log.Error(ctx, "add schedule: error",
			log.Err(err),
			log.Any("viewData", viewData),
			log.Any("schedule", schedule),
		)
		return "", err
	}
	return scheduleID, nil
}

func (s *scheduleModel) addSchedule(ctx context.Context, tx *dbo.DBContext, schedule *entity.Schedule, teacherIDs []string, options *entity.RepeatOptions, location *time.Location) (string, error) {
	scheduleList, err := s.RepeatSchedule(ctx, schedule, options, location)
	if err != nil {
		log.Error(ctx, "schedule repeat error", log.Err(err), log.Any("schedule", schedule), log.Any("options", options))
		return "", err
	}
	scheduleTeachers := make([]*entity.ScheduleTeacher, len(scheduleList)*len(teacherIDs))
	index := 0
	for _, item := range scheduleList {
		item.ID = utils.NewID()
		for _, teacherID := range teacherIDs {
			tsItem := &entity.ScheduleTeacher{
				ID:         utils.NewID(),
				TeacherID:  teacherID,
				ScheduleID: item.ID,
			}
			scheduleTeachers[index] = tsItem
			index++
		}
	}

	// add to schedules
	_, err = da.GetScheduleDA().BatchInsert(ctx, tx, scheduleList)
	if err != nil {
		log.Error(ctx, "schedule batchInsert error", log.Err(err), log.Any("scheduleList", scheduleList))
		return "", err
	}

	// add to teachers_schedules
	_, err = da.GetScheduleTeacherDA().BatchInsert(ctx, tx, scheduleTeachers)
	if err != nil {
		log.Error(ctx, "teachers_schedules batchInsert error", log.Err(err), log.Any("scheduleTeachers", scheduleTeachers))
		return "", err
	}
	if len(scheduleList) <= 0 {
		log.Error(ctx, "schedules batchInsert error,schedules is empty", log.Any("schedule", schedule), log.Any("options", options))
		return "", constant.ErrRecordNotFound
	}

	return scheduleList[0].ID, nil
}
func (s *scheduleModel) checkScheduleStatus(ctx context.Context, id string) (*entity.Schedule, error) {
	// get old schedule by id
	var schedule = new(entity.Schedule)
	err := da.GetScheduleDA().Get(ctx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "checkScheduleStatus: get schedule by id failed, schedule not found", log.Err(err), log.String("id", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "checkScheduleStatus: get schedule by id failed",
			log.Err(err),
			log.String("id", id),
		)
		return nil, err
	}
	if schedule.DeleteAt != 0 {
		log.Error(ctx, "checkScheduleStatus: get schedule by id failed, schedule not found",
			log.String("id", id),
		)
		return nil, constant.ErrRecordNotFound
	}
	if schedule.Status != entity.ScheduleStatusNotStart {
		log.Warn(ctx, "checkScheduleStatus: schedule status error",
			log.String("id", id),
			log.Any("schedule", schedule),
		)
		return schedule, constant.ErrOperateNotAllowed
	}
	return schedule, nil
}
func (s *scheduleModel) Update(ctx context.Context, operator *entity.Operator, viewData *entity.ScheduleUpdateView) (string, error) {
	schedule, err := s.checkScheduleStatus(ctx, viewData.ID)
	if err != nil {
		log.Error(ctx, "update schedule: get schedule by id error",
			log.Any("viewData", viewData),
			log.Err(err),
		)
		return "", err
	}
	// verify data
	err = s.verifyData(ctx, &entity.ScheduleVerify{
		ClassID:      viewData.ClassID,
		SubjectID:    viewData.SubjectID,
		ProgramID:    viewData.ProgramID,
		TeacherIDs:   viewData.TeacherIDs,
		LessonPlanID: viewData.LessonPlanID,
	})
	if err != nil {
		log.Error(ctx, "update schedule: verify data error",
			log.Err(err),
			log.Any("viewData", viewData))
		return "", constant.ErrInvalidArgs
	}

	// not force add need conflict detection
	if !viewData.IsForce {
		conflict, err := s.IsScheduleConflict(ctx, operator, viewData.StartAt, viewData.EndAt)
		if err != nil {
			log.Error(ctx, "update schedule: check time conflict failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("viewData", viewData),
			)
			return "", err
		}

		if conflict {
			log.Info(ctx, "update schedule: time conflict",
				log.Any("operator", operator),
				log.Any("viewData", viewData),
			)
			return "", constant.ErrConflict
		}
	}

	// update schedule
	var id string
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// delete schedule
		var err error
		if err = s.deleteScheduleTx(ctx, tx, operator, schedule, viewData.EditType); err != nil {
			log.Error(ctx, "update schedule: delete failed",
				log.Err(err),
				log.String("id", viewData.ID),
				log.String("edit_type", string(viewData.EditType)),
			)
			return err
		}
		// add schedule,update old schedule fields that need to be updated
		schedule.ID = utils.NewID()
		schedule.LessonPlanID = viewData.LessonPlanID
		schedule.ProgramID = viewData.ProgramID
		schedule.SubjectID = viewData.SubjectID
		schedule.ClassID = viewData.ClassID
		schedule.StartAt = viewData.StartAt
		schedule.EndAt = viewData.EndAt
		schedule.Title = viewData.Title
		schedule.IsAllDay = viewData.IsAllDay
		schedule.Description = viewData.Description
		schedule.DueAt = viewData.DueAt
		schedule.ClassType = viewData.ClassType
		schedule.CreatedID = operator.UserID
		schedule.CreatedAt = time.Now().Unix()
		schedule.UpdatedID = operator.UserID
		schedule.UpdatedAt = time.Now().Unix()
		schedule.DeletedID = ""
		schedule.DeleteAt = 0
		// attachment
		b, err := json.Marshal(viewData.Attachment)
		if err != nil {
			log.Warn(ctx, "update schedule:marshal attachment error", log.Any("attachment", viewData.Attachment))
			return err
		}
		schedule.Attachment = string(b)

		// update repeat rule
		var repeatOptions *entity.RepeatOptions
		// if repeat selected, use repeat rule
		if viewData.IsRepeat {
			b, err := json.Marshal(viewData.Repeat)
			if err != nil {
				return err
			}
			schedule.RepeatJson = string(b)
			// if following selected, set repeat rule
			if viewData.EditType == entity.ScheduleEditWithFollowing {
				repeatOptions = &viewData.Repeat
			}
		} else {
			// if repeat not selected,but need to update follow schedule, use old schedule repeat rule
			if viewData.EditType == entity.ScheduleEditWithFollowing {
				var repeat = new(entity.RepeatOptions)
				if err := json.Unmarshal([]byte(schedule.RepeatJson), repeat); err != nil {
					log.Error(ctx, "update schedule:unmarshal schedule repeatJson error",
						log.Err(err),
						log.Any("viewData", viewData),
						log.Any("schedule", schedule),
					)
					return err
				}
				repeatOptions = repeat
			}
		}

		id, err = s.addSchedule(ctx, tx, schedule, viewData.TeacherIDs, repeatOptions, viewData.Location)
		if err != nil {
			log.Error(ctx, "update schedule: add failed",
				log.Err(err),
				log.Any("schedule", schedule),
				log.Any("viewData", viewData),
			)
			return err
		}
		return nil
	}); err != nil {
		log.Error(ctx, "update schedule: tx failed", log.Err(err))
		return "", err
	}
	da.GetScheduleRedisDA().Clean(ctx, []string{viewData.ID})
	return id, nil
}
func (s *scheduleModel) Delete(ctx context.Context, op *entity.Operator, id string, editType entity.ScheduleEditType) error {
	schedule, err := s.checkScheduleStatus(ctx, id)
	if err == constant.ErrRecordNotFound {
		log.Warn(ctx, "DeleteTx:schedule not found",
			log.Err(err),
			log.String("id", id),
			log.String("edit_type", string(editType)),
		)
		return nil
	}
	if err != nil {
		log.Error(ctx, "DeleteTx:delete schedule by id error",
			log.Err(err),
			log.String("id", id),
			log.String("edit_type", string(editType)),
		)
		return err
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		return s.deleteScheduleTx(ctx, tx, op, schedule, editType)
	})
	if err != nil {
		log.Error(ctx, "delete schedule error",
			log.Err(err),
			log.String("id", id),
			log.String("edit_type", string(editType)),
		)
		return err
	}
	da.GetScheduleRedisDA().Clean(ctx, []string{id})
	return nil
}

func (s *scheduleModel) deleteScheduleTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, schedule *entity.Schedule, editType entity.ScheduleEditType) error {
	switch editType {
	case entity.ScheduleEditOnlyCurrent:
		if err := da.GetScheduleDA().SoftDelete(ctx, tx, schedule.ID, op); err != nil {
			log.Error(ctx, "delete schedule: soft delete failed",
				log.Any("schedule", schedule),
				log.String("edit_type", string(editType)),
			)
			return err
		}
	case entity.ScheduleEditWithFollowing:
		if err := da.GetScheduleDA().DeleteWithFollowing(ctx, tx, schedule.RepeatID, schedule.StartAt); err != nil {
			log.Error(ctx, "delete schedule: delete with following failed",
				log.Err(err),
				log.String("repeat_id", schedule.RepeatID),
				log.Int64("start_at", schedule.StartAt),
				log.String("edit_type", string(editType)),
			)
			return err
		}
	}
	// delete schedules_teachers data
	var scheduleList []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, &da.ScheduleCondition{
		RepeatID: sql.NullString{
			String: schedule.RepeatID,
			Valid:  true,
		},
		Status: sql.NullString{
			String: string(entity.ScheduleStatusNotStart),
			Valid:  true,
		},
	}, &scheduleList)
	if err != nil {
		log.Error(ctx, "delete schedule: delete with following failed",
			log.Err(err),
			log.String("repeat_id", schedule.RepeatID),
			log.String("edit_type", string(editType)),
		)
		return err
	}
	scheduleIDs := make([]string, len(scheduleList))
	for i, item := range scheduleList {
		scheduleIDs[i] = item.ID
	}
	if err := da.GetScheduleTeacherDA().BatchDelByScheduleIDs(ctx, tx, scheduleIDs); err != nil {
		log.Error(ctx, "delete schedule: batch delete  by schedule ids failed",
			log.Err(err),
			log.Strings("scheduleIDs", scheduleIDs),
		)
		return err
	}
	return nil
}

func (s *scheduleModel) Page(ctx context.Context, condition *da.ScheduleCondition) (int, []*entity.ScheduleSearchView, error) {
	var scheduleList []*entity.Schedule
	total, err := da.GetScheduleDA().Page(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "Page: schedule query error", log.Err(err), log.Any("condition", condition))
		return 0, nil, err
	}

	result := make([]*entity.ScheduleSearchView, len(scheduleList))
	basicInfo, err := s.getBasicInfo(ctx, scheduleList)
	if err != nil {
		log.Error(ctx, "Page: get basic info error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("scheduleList", scheduleList))
		return 0, nil, err
	}
	for i, item := range scheduleList {
		viewData := &entity.ScheduleSearchView{
			ID:      item.ID,
			StartAt: item.StartAt,
			Title:   item.Title,
			EndAt:   item.EndAt,
		}
		if v, ok := basicInfo[item.ID]; ok {
			viewData.ScheduleBasic = *v
		}
		result[i] = viewData
	}

	return total, result, nil
}

func (s *scheduleModel) Query(ctx context.Context, condition *da.ScheduleCondition) ([]*entity.ScheduleListView, error) {
	cacheData, err := da.GetScheduleRedisDA().GetScheduleCacheByCondition(ctx, condition)
	if err == nil && len(cacheData) > 0 {
		log.Debug(ctx, "Query:using cache",
			log.Any("condition", condition),
			log.Any("cacheData", cacheData),
		)
		return cacheData, nil
	}
	var scheduleList []*entity.Schedule
	err = da.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "schedule query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	result := make([]*entity.ScheduleListView, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = &entity.ScheduleListView{
			ID:           item.ID,
			Title:        item.Title,
			StartAt:      item.StartAt,
			EndAt:        item.EndAt,
			IsRepeat:     item.RepeatID != "",
			LessonPlanID: item.LessonPlanID,
			Status:       item.Status,
			ClassType:    item.ClassType,
		}
	}
	da.GetScheduleRedisDA().AddScheduleByCondition(ctx, condition, result)

	return result, nil
}

func (s *scheduleModel) getBasicInfo(ctx context.Context, schedules []*entity.Schedule) (map[string]*entity.ScheduleBasic, error) {
	var (
		classIDs           []string
		classMap           map[string]*entity.ScheduleShortInfo
		subjectIDs         []string
		subjectMap         map[string]*entity.ScheduleShortInfo
		programIDs         []string
		programMap         map[string]*entity.ScheduleShortInfo
		scheduleIDs        []string
		teacherIDs         []string
		teacherMap         map[string]*entity.ScheduleShortInfo
		scheduleTeacherMap map[string][]string
		lessonPlanIDs      []string
		lessonPlanMap      map[string]*entity.ScheduleShortInfo
	)
	for _, item := range schedules {
		classIDs = append(classIDs, item.ClassID)
		subjectIDs = append(subjectIDs, item.SubjectID)
		programIDs = append(programIDs, item.ProgramID)
		scheduleIDs = append(scheduleIDs, item.ID)
		lessonPlanIDs = append(lessonPlanIDs, item.LessonPlanID)
	}

	classMap, err := s.getClassInfoMapByClassIDs(ctx, classIDs)
	if err != nil {
		log.Error(ctx, "getBasicInfo:get class info error", log.Err(err), log.Strings("classIDs", classIDs))
		return nil, err
	}
	subjectMap, err = s.geSubjectInfoMapBySubjectIDs(ctx, subjectIDs)
	if err != nil {
		log.Error(ctx, "getBasicInfo:get subject info error", log.Err(err), log.Strings("subjectIDs", subjectIDs))
		return nil, err
	}

	programMap, err = s.getProgramInfoMapByProgramIDs(ctx, programIDs)
	if err != nil {
		log.Error(ctx, "getBasicInfo:get program info error", log.Err(err), log.Strings("programIDs", programIDs))
		return nil, err
	}
	teacherMap = make(map[string]*entity.ScheduleShortInfo)
	scheduleTeacherMap = make(map[string][]string)
	if len(scheduleIDs) > 0 {
		var scheduleTeacherList []*entity.ScheduleTeacher
		err := da.GetScheduleTeacherDA().Query(ctx, &da.ScheduleTeacherCondition{
			ScheduleIDs: entity.NullStrings{
				Strings: scheduleIDs,
				Valid:   true,
			},
		}, &scheduleTeacherList)
		if err != nil {
			return nil, err
		}
		teacherIDs = make([]string, len(scheduleTeacherList))
		for i, item := range scheduleTeacherList {
			teacherIDs[i] = item.TeacherID
			if _, ok := scheduleTeacherMap[item.ScheduleID]; !ok {
				scheduleTeacherMap[item.ScheduleID] = []string{}
			}
			scheduleTeacherMap[item.ScheduleID] = append(scheduleTeacherMap[item.ScheduleID], item.TeacherID)
		}
		teacherIDs = utils.SliceDeduplication(teacherIDs)
		teacherService := external.GetTeacherServiceProvider()
		teacherInfos, err := teacherService.BatchGet(ctx, teacherIDs)
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetTeacherServiceProvider BatchGet error", log.Err(err), log.Any("schedules", schedules))
			return nil, err
		}
		for _, item := range teacherInfos {
			teacherMap[item.ID] = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	lessonPlanMap, err = s.getLessonPlanMapByLessonPlanIDs(ctx, dbo.MustGetDB(ctx), lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "getBasicInfo:get lesson plan info error", log.Err(err), log.Any("lessonPlanIDs", lessonPlanIDs))
		return nil, err
	}
	scheduleBasicMap := make(map[string]*entity.ScheduleBasic)
	for _, item := range schedules {
		scheduleBasic := &entity.ScheduleBasic{}
		if v, ok := classMap[item.ClassID]; ok {
			scheduleBasic.Class = *v
		}
		if v, ok := subjectMap[item.SubjectID]; ok {
			scheduleBasic.Subject = *v
		}
		if v, ok := programMap[item.ProgramID]; ok {
			scheduleBasic.Program = *v
		}
		if v, ok := lessonPlanMap[item.LessonPlanID]; ok {
			scheduleBasic.LessonPlan = *v
		}
		tIDs := scheduleTeacherMap[item.ID]
		scheduleBasic.Teachers = make([]entity.ScheduleShortInfo, 0, len(tIDs))
		for _, tID := range tIDs {
			if v, ok := teacherMap[tID]; ok {
				scheduleBasic.Teachers = append(scheduleBasic.Teachers, *v)
			}
		}
		scheduleBasicMap[item.ID] = scheduleBasic
	}
	return scheduleBasicMap, nil
}

func (s *scheduleModel) getLessonPlanMapByLessonPlanIDs(ctx context.Context, tx *dbo.DBContext, lessonPlanIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	lessonPlanMap := make(map[string]*entity.ScheduleShortInfo)
	if len(lessonPlanIDs) != 0 {
		lessonPlanIDs = utils.SliceDeduplication(lessonPlanIDs)
		lessonPlans, err := GetContentModel().GetContentNameByIDList(ctx, tx, lessonPlanIDs)
		if err != nil {
			log.Error(ctx, "getBasicInfo:get lesson plan info error", log.Err(err), log.Strings("lessonPlanIDs", lessonPlanIDs))
			return nil, err
		}

		for _, item := range lessonPlans {
			lessonPlanMap[item.ID] = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	return lessonPlanMap, nil
}

func (s *scheduleModel) getClassInfoMapByClassIDs(ctx context.Context, classIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	var classMap = make(map[string]*entity.ScheduleShortInfo)
	if len(classIDs) != 0 {
		classIDs = utils.SliceDeduplication(classIDs)
		classService := external.GetClassServiceProvider()
		classInfos, err := classService.BatchGet(ctx, classIDs)
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetClassServiceProvider BatchGet error", log.Err(err), log.Strings("classIDs", classIDs))
			return nil, err
		}
		for _, item := range classInfos {
			classMap[item.ID] = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	return classMap, nil
}

func (s *scheduleModel) geSubjectInfoMapBySubjectIDs(ctx context.Context, subjectIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	var subjectMap = make(map[string]*entity.ScheduleShortInfo)
	if len(subjectIDs) != 0 {
		subjectIDs = utils.SliceDeduplication(subjectIDs)
		subjectService := external.GetSubjectServiceProvider()
		subjectInfos, err := subjectService.BatchGet(ctx, subjectIDs)
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetSubjectServiceProvider BatchGet error", log.Err(err), log.Strings("subjectIDs", subjectIDs))
			return nil, err
		}

		for _, item := range subjectInfos {
			subjectMap[item.ID] = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	return subjectMap, nil
}

func (s *scheduleModel) getProgramInfoMapByProgramIDs(ctx context.Context, programIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	var programMap = make(map[string]*entity.ScheduleShortInfo)
	if len(programIDs) != 0 {
		programIDs = utils.SliceDeduplication(programIDs)
		programService := external.GetProgramServiceProvider()
		programInfos, err := programService.BatchGet(ctx, programIDs)
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetProgramServiceProvider BatchGet error", log.Err(err), log.Strings("programIDs", programIDs))
			return nil, err
		}

		for _, item := range programInfos {
			programMap[item.ID] = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	return programMap, nil
}

func (s *scheduleModel) GetByID(ctx context.Context, id string) (*entity.ScheduleDetailsView, error) {
	cacheData, err := da.GetScheduleRedisDA().GetScheduleCacheByIDs(ctx, []string{id})
	if err == nil && len(cacheData) > 0 {
		log.Debug(ctx, "GetByID:using cache",
			log.Any("id", id),
			log.Any("cacheData", cacheData),
		)
		return cacheData[0], nil
	}
	var schedule = new(entity.Schedule)
	err = da.GetScheduleDA().Get(ctx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetByID error", log.Err(err), log.String("id", id))
		return nil, err
	}
	if schedule.DeleteAt != 0 {
		return nil, constant.ErrRecordNotFound
	}

	result := &entity.ScheduleDetailsView{
		ID:          schedule.ID,
		Title:       schedule.Title,
		OrgID:       schedule.OrgID,
		StartAt:     schedule.StartAt,
		EndAt:       schedule.EndAt,
		IsAllDay:    schedule.IsAllDay,
		ClassType:   schedule.ClassType,
		DueAt:       schedule.DueAt,
		Description: schedule.Description,
		Version:     schedule.ScheduleVersion,
		IsRepeat:    schedule.RepeatID != "",
		Status:      schedule.Status,
	}
	if schedule.Attachment != "" {
		var attachment entity.ScheduleShortInfo
		err := json.Unmarshal([]byte(schedule.Attachment), &attachment)
		if err != nil {
			log.Error(ctx, "Unmarshal schedule.Attachment error", log.Err(err), log.String("schedule.Attachment", schedule.Attachment))
			return nil, err
		}
		result.Attachment = attachment
	}
	if schedule.RepeatJson != "" {
		var repeat entity.RepeatOptions
		err := json.Unmarshal([]byte(schedule.RepeatJson), &repeat)
		if err != nil {
			log.Error(ctx, "Unmarshal schedule.RepeatJson error", log.Err(err), log.String("schedule.RepeatJson", schedule.RepeatJson))
			return nil, err
		}
		result.Repeat = repeat
	}
	basicInfo, err := s.getBasicInfo(ctx, []*entity.Schedule{schedule})
	if err != nil {
		log.Error(ctx, "getBasicInfo error", log.Err(err))
		return nil, err
	}
	if v, ok := basicInfo[result.ID]; ok {
		result.ScheduleBasic = *v
	}
	da.GetScheduleRedisDA().BatchAddScheduleCache(ctx, []*entity.ScheduleDetailsView{result})
	return result, nil
}

func (s *scheduleModel) GetTeacherByName(ctx context.Context, name string) ([]*external.Teacher, error) {
	teacherService := external.GetTeacherServiceProvider()
	teachers, err := teacherService.Query(ctx, name)
	if err != nil {
		log.Error(ctx, "querySchedule:query teacher info error", log.Err(err), log.String("name", name))
		return nil, err
	}

	return teachers, nil
}

func (s *scheduleModel) ExistScheduleByLessonPlanID(ctx context.Context, lessonPlanID string) (bool, error) {
	if strings.TrimSpace(lessonPlanID) == "" {
		log.Info(ctx, "lessonPlanID is empty", log.String("lessonPlanID", lessonPlanID))
		return false, errors.New("lessonPlanID is empty")
	}
	condition := &da.ScheduleCondition{
		LessonPlanID: sql.NullString{
			String: lessonPlanID,
			Valid:  true,
		},
	}
	count, err := da.GetScheduleDA().Count(ctx, condition, &entity.Schedule{})
	if err != nil {
		log.Error(ctx, "get schedule count by condition error", log.Err(err), log.Any("condition", condition))
		return false, err
	}

	return count > 0, nil
}

func (s *scheduleModel) ExistScheduleByID(ctx context.Context, id string) (bool, error) {
	condition := &da.ScheduleCondition{
		ID: sql.NullString{
			String: id,
			Valid:  true,
		},
	}
	count, err := da.GetScheduleDA().Count(ctx, condition, &entity.Schedule{})
	if err != nil {
		log.Error(ctx, "get schedule count by condition error", log.Err(err), log.Any("condition", condition))
		return false, err
	}

	return count > 0, nil
}

func (s *scheduleModel) GetPlainByID(ctx context.Context, id string) (*entity.SchedulePlain, error) {
	var schedule = new(entity.Schedule)
	err := da.GetScheduleDA().Get(ctx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "GetPlainByID not found", log.Err(err), log.String("id", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetPlainByID error", log.Err(err), log.String("id", id))
		return nil, err
	}
	if schedule.DeleteAt != 0 {
		log.Error(ctx, "GetPlainByID deleted", log.Err(err), log.Any("schedule", schedule))
		return nil, constant.ErrRecordNotFound
	}
	result := new(entity.SchedulePlain)
	result.Schedule = *schedule

	var scheduleTeacherList []*entity.ScheduleTeacher
	err = da.GetScheduleTeacherDA().Query(ctx, &da.ScheduleTeacherCondition{
		ScheduleID: sql.NullString{
			String: schedule.ID,
			Valid:  true,
		},
	}, &scheduleTeacherList)
	if err != nil {
		log.Error(ctx, "GetPlainByID:get schedule_teacher error", log.Err(err), log.Any("schedule", schedule))
		return nil, err
	}
	result.TeacherIDs = make([]string, len(scheduleTeacherList))
	for i, item := range scheduleTeacherList {
		result.TeacherIDs[i] = item.TeacherID
	}
	return result, nil
}

func (s *scheduleModel) verifyData(ctx context.Context, v *entity.ScheduleVerify) error {
	// class
	classService := external.GetClassServiceProvider()
	_, err := classService.BatchGet(ctx, []string{v.ClassID})
	if err != nil {
		log.Error(ctx, "getBasicInfo:GetClassServiceProvider BatchGet error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}
	// teacher
	teacherIDs := utils.SliceDeduplication(v.TeacherIDs)
	teacherService := external.GetTeacherServiceProvider()
	_, err = teacherService.BatchGet(ctx, teacherIDs)
	if err != nil {
		log.Error(ctx, "getBasicInfo:GetProgramServiceProvider BatchGet error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}

	if v.ClassType == entity.ScheduleClassTypeTask {
		if v.LessonPlanID != "" || v.ProgramID != "" || v.SubjectID != "" {
			return constant.ErrInvalidArgs
		}
		return nil
	}
	// subject
	subjectService := external.GetSubjectServiceProvider()
	_, err = subjectService.BatchGet(ctx, []string{v.SubjectID})
	if err != nil {
		log.Error(ctx, "getBasicInfo:GetSubjectServiceProvider BatchGet error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}
	// program
	programService := external.GetProgramServiceProvider()
	_, err = programService.BatchGet(ctx, []string{v.ProgramID})
	if err != nil {
		log.Error(ctx, "getBasicInfo:GetProgramServiceProvider BatchGet error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}

	// lessPlan
	lessonPlanInfo, err := GetContentModel().GetContentNameByID(ctx, dbo.MustGetDB(ctx), v.LessonPlanID)
	if err != nil {
		log.Error(ctx, "getBasicInfo:get lessPlan info error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}
	if lessonPlanInfo.ContentType != entity.ContentTypeLesson {
		log.Error(ctx, "getBasicInfo:content type is not lesson", log.Any("lessonPlanInfo", lessonPlanInfo), log.Any("ScheduleVerify", v))
		return constant.ErrInvalidArgs
	}
	return nil
}

func (s *scheduleModel) UpdateScheduleStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.ScheduleStatus) error {
	var schedule = new(entity.Schedule)
	err := da.GetScheduleDA().GetTx(ctx, tx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "UpdateScheduleStatus: get schedule by id failed, schedule not found", log.Err(err), log.String("id", id))
		return constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "UpdateScheduleStatus: get schedule by id failed",
			log.Err(err),
			log.String("id", id),
		)
		return err
	}
	if schedule.DeleteAt != 0 {
		log.Error(ctx, "UpdateScheduleStatus: get schedule by id failed, schedule not found", log.String("id", id))
		return constant.ErrRecordNotFound
	}

	schedule.Status = status
	_, err = da.GetScheduleDA().UpdateTx(ctx, tx, schedule)
	if err != nil {
		log.Error(ctx, "UpdateScheduleStatus: update schedule status error",
			log.String("id", id),
			log.Any("schedule", schedule),
			log.Err(err),
		)
		return err
	}
	da.GetScheduleRedisDA().Clean(ctx, []string{id})
	return nil
}

func (s *scheduleModel) GetParticipateClass(ctx context.Context, operator *entity.Operator) ([]*external.Class, error) {
	// user is admin
	if operator.Role == string(constant.RoleAdmin) {
		result, err := external.GetClassServiceProvider().BatchGet(ctx, nil)
		if err != nil {
			log.Error(ctx, "GetParticipateClass:batch get class from ClassServiceProvider error", log.Err(err), log.Any("op", operator))
			return nil, err
		}
		return result, nil
	}
	// user is not admin
	classIDs, err := da.GetScheduleDA().GetParticipateClass(ctx, dbo.MustGetDB(ctx), operator.UserID)
	if err != nil {
		log.Error(ctx, "GetParticipateClass:get participate  class from db error", log.Err(err), log.Any("op", operator))
		return nil, err
	}
	result, err := external.GetClassServiceProvider().BatchGet(ctx, classIDs)
	if err != nil {
		log.Error(ctx, "GetParticipateClass:batch get class from ClassServiceProvider error",
			log.Err(err),
			log.Any("op", operator),
			log.Strings("classIDs", classIDs),
		)
		return nil, err
	}
	return result, nil
}

var (
	_scheduleOnce  sync.Once
	_scheduleModel IScheduleModel
)

func GetScheduleModel() IScheduleModel {
	_scheduleOnce.Do(func() {
		_scheduleModel = &scheduleModel{}
	})
	return _scheduleModel
}
