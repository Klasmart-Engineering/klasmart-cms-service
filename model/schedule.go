package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	ErrScheduleEditMissTime       = errors.New("editable time has expired")
	ErrScheduleLessonPlanUnAuthed = errors.New("schedule content data unAuthed")
)

type IScheduleModel interface {
	Add(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error)
	AddTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error)
	Update(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleUpdateView) (string, error)
	Delete(ctx context.Context, op *entity.Operator, id string, editType entity.ScheduleEditType) error
	//DeleteTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, id string, editType entity.ScheduleEditType) error
	Query(ctx context.Context, condition *da.ScheduleCondition, loc *time.Location) ([]*entity.ScheduleListView, error)
	QueryScheduledDates(ctx context.Context, condition *da.ScheduleCondition, loc *time.Location) ([]string, error)
	Page(ctx context.Context, operator *entity.Operator, condition *da.ScheduleCondition) (int, []*entity.ScheduleSearchView, error)
	GetByID(ctx context.Context, operator *entity.Operator, id string) (*entity.ScheduleDetailsView, error)
	ConflictDetection(ctx context.Context, op *entity.Operator, input *entity.ScheduleConflictInput) (*entity.ScheduleConflictView, error)
	GetOrgClassIDsByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string, orgID string) ([]string, error)
	GetTeacherByName(ctx context.Context, operator *entity.Operator, OrgID, name string) ([]*external.Teacher, error)
	ExistScheduleAttachmentFile(ctx context.Context, attachmentPath string) bool
	ExistScheduleByLessonPlanID(ctx context.Context, lessonPlanID string) (bool, error)
	ExistScheduleByID(ctx context.Context, id string) (bool, error)
	GetPlainByID(ctx context.Context, id string) (*entity.SchedulePlain, error)
	UpdateScheduleStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.ScheduleStatus) error
	GetParticipateClass(ctx context.Context, operator *entity.Operator) ([]*external.Class, error)
	GetLessonPlanByCondition(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, condition *da.ScheduleCondition) ([]*entity.ScheduleShortInfo, error)
	GetScheduleIDsByCondition(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, condition *entity.ScheduleIDsCondition) ([]string, error)
	GetScheduleIDsByOrgID(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, orgID string) ([]string, error)
	VerifyLessonPlanAuthed(ctx context.Context, operator *entity.Operator, lessonPlanID string) error
	GetScheduleRealTimeStatus(ctx context.Context, op *entity.Operator, id string) (*entity.ScheduleRealTimeView, error)
}
type scheduleModel struct {
	testScheduleRepeatFlag bool
}

func (s *scheduleModel) GetOrgClassIDsByUserIDs(ctx context.Context, operator *entity.Operator, userIDs []string, orgID string) ([]string, error) {
	userClassInfos, err := external.GetClassServiceProvider().GetByUserIDs(ctx, operator, userIDs)
	if err != nil {
		log.Error(ctx, "GetMyOrgClassIDs:GetClassServiceProvider.GetByUserID error",
			log.Err(err),
			log.String("org", orgID),
			log.Strings("userIDs", userIDs))
		return nil, err
	}

	myClassIDs := make([]string, 0)
	for _, classInfos := range userClassInfos {
		for _, item := range classInfos {
			myClassIDs = append(myClassIDs, item.ID)
		}
	}
	orgClassInfoMap, err := external.GetClassServiceProvider().GetByOrganizationIDs(ctx, operator, []string{orgID})
	if err != nil {
		log.Error(ctx, "GetMyOrgClassIDs:GetClassServiceProvider.GetByOrganizationIDs error",
			log.Err(err),
			log.Strings("userIDs", userIDs),
			log.String("orgID", orgID),
		)
		return nil, err
	}
	orgClassInfos := orgClassInfoMap[orgID]
	orgClassIDs := make([]string, len(orgClassInfos))
	for i, item := range orgClassInfos {
		orgClassIDs[i] = item.ID
	}
	result := utils.IntersectAndDeduplicateStrSlice(myClassIDs, orgClassIDs)
	log.Debug(ctx, "my org class ids",
		log.String("orgID", orgID),
		log.Strings("userIDs", userIDs),
		log.Strings("myClassIDs", myClassIDs),
		log.Strings("orgClassIDs", orgClassIDs),
		log.Strings("result", result),
	)
	return result, nil
}

func (s *scheduleModel) ConflictDetection(ctx context.Context, op *entity.Operator, input *entity.ScheduleConflictInput) (*entity.ScheduleConflictView, error) {
	//var scheduleList []*entity.Schedule
	//StartAndEndRange := make([]sql.NullInt64, 2)
	//StartAndEndRange[0] = sql.NullInt64{
	//	Valid: startAt <= 0,
	//	Int64: startAt,
	//}
	//StartAndEndRange[1] = sql.NullInt64{
	//	Valid: endAt <= 0,
	//	Int64: endAt,
	//}
	//err := da.GetScheduleDA().Query(ctx, &da.ScheduleCondition{
	//	OrgID: sql.NullString{
	//		String: op.OrgID,
	//		Valid:  op.OrgID != "",
	//	},
	//	StartAndEndRange: StartAndEndRange,
	//}, &scheduleList)
	//if err != nil {
	//	return false, err
	//}
	//if len(scheduleList) > 0 {
	//	log.Debug(ctx, "conflict schedule data", log.Any("scheduleList", scheduleList))
	//	return true, nil
	//}
	return nil, nil
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
	err = da.GetScheduleRedisDA().Clean(ctx, nil)
	if err != nil {
		log.Info(ctx, "Add:GetScheduleRedisDA.Clean error", log.Err(err))
	}
	return id.(string), nil
}
func (s *scheduleModel) AddTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error) {
	// verify data
	err := s.verifyData(ctx, op, &entity.ScheduleVerify{
		ClassID:      viewData.ClassID,
		SubjectID:    viewData.SubjectID,
		ProgramID:    viewData.ProgramID,
		LessonPlanID: viewData.LessonPlanID,
		ClassType:    viewData.ClassType,
	})
	if err != nil {
		log.Error(ctx, "add schedule: verify data error",
			log.Err(err),
			log.Any("viewData", viewData))
		return "", constant.ErrInvalidArgs
	}
	if viewData.ClassType == entity.ScheduleClassTypeTask {
		viewData.LessonPlanID = ""
		viewData.ProgramID = ""
		viewData.SubjectID = ""
	}

	schedule, err := viewData.ToSchedule(ctx)
	schedule.CreatedID = op.UserID
	scheduleID, err := s.addSchedule(ctx, tx, schedule, &viewData.Repeat, viewData.Location)
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

func (s *scheduleModel) addSchedule(ctx context.Context, tx *dbo.DBContext, schedule *entity.Schedule, options *entity.RepeatOptions, location *time.Location) (string, error) {
	scheduleList, err := s.StartScheduleRepeat(ctx, schedule, options, location)
	if err != nil {
		log.Error(ctx, "schedule repeat error", log.Err(err), log.Any("schedule", schedule), log.Any("options", options))
		return "", err
	}
	// add to schedules
	_, err = da.GetScheduleDA().BatchInsert(ctx, tx, scheduleList)
	if err != nil {
		log.Error(ctx, "schedule batchInsert error", log.Err(err), log.Any("scheduleList", scheduleList))
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
		return nil, constant.ErrOperateNotAllowed
	}
	switch schedule.ClassType {
	case entity.ScheduleClassTypeHomework, entity.ScheduleClassTypeTask:

	case entity.ScheduleClassTypeOnlineClass, entity.ScheduleClassTypeOfflineClass:
		diff := utils.TimeStampDiff(schedule.StartAt, time.Now().Unix())
		if diff <= constant.ScheduleAllowEditTime {
			log.Warn(ctx, "checkScheduleStatus: GetDiffToMinutesByTimeStamp warn",
				log.Any("schedule", schedule),
				log.Int64("schedule.StartAt", schedule.StartAt),
				log.Any("diff", diff),
				log.Any("ScheduleAllowEditTime", constant.ScheduleAllowEditTime),
			)
			return nil, ErrScheduleEditMissTime
		}
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
	err = s.verifyData(ctx, operator, &entity.ScheduleVerify{
		ClassID:      viewData.ClassID,
		SubjectID:    viewData.SubjectID,
		ProgramID:    viewData.ProgramID,
		LessonPlanID: viewData.LessonPlanID,
		ClassType:    viewData.ClassType,
	})
	if err != nil {
		log.Error(ctx, "update schedule: verify data error",
			log.Err(err),
			log.Any("viewData", viewData))
		return "", constant.ErrInvalidArgs
	}

	if viewData.ClassType == entity.ScheduleClassTypeTask {
		viewData.LessonPlanID = ""
		viewData.ProgramID = ""
		viewData.SubjectID = ""
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
			if schedule.RepeatID == "" {
				schedule.RepeatID = utils.NewID()
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

		id, err = s.addSchedule(ctx, tx, schedule, repeatOptions, viewData.Location)
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
	err = da.GetScheduleRedisDA().Clean(ctx, []string{viewData.ID})
	if err != nil {
		log.Info(ctx, "Update:GetScheduleRedisDA.Clean error", log.Err(err))
	}
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
	err = da.GetScheduleRedisDA().Clean(ctx, []string{id})
	if err != nil {
		log.Info(ctx, "Delete:GetScheduleRedisDA.Clean error", log.Err(err))
	}

	return nil
}

func (s *scheduleModel) deleteScheduleTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, schedule *entity.Schedule, editType entity.ScheduleEditType) error {
	scheduleIDs := make([]string, 0)

	switch editType {
	case entity.ScheduleEditOnlyCurrent:
		if err := da.GetScheduleDA().SoftDelete(ctx, tx, schedule.ID, op); err != nil {
			log.Error(ctx, "delete schedule: soft delete failed",
				log.Any("schedule", schedule),
				log.String("edit_type", string(editType)),
			)
			return err
		}
		scheduleIDs = append(scheduleIDs, schedule.ID)

	case entity.ScheduleEditWithFollowing:
		if schedule.RepeatID == "" {
			if err := da.GetScheduleDA().SoftDelete(ctx, tx, schedule.ID, op); err != nil {
				log.Error(ctx, "delete schedule: soft delete failed",
					log.Any("schedule", schedule),
					log.String("edit_type", string(editType)),
				)
				return err
			}
			return nil
		}
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
	return nil
}

func (s *scheduleModel) Page(ctx context.Context, operator *entity.Operator, condition *da.ScheduleCondition) (int, []*entity.ScheduleSearchView, error) {
	var scheduleList []*entity.Schedule
	total, err := da.GetScheduleDA().Page(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "Page: schedule query error", log.Err(err), log.Any("condition", condition))
		return 0, nil, err
	}

	result := make([]*entity.ScheduleSearchView, 0, len(scheduleList))
	basicInfo, err := s.getBasicInfo(ctx, operator, scheduleList)
	if err != nil {
		log.Error(ctx, "Page: get basic info error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("scheduleList", scheduleList))
		return 0, nil, err
	}
	for _, item := range scheduleList {
		if item.ClassType == entity.ScheduleClassTypeHomework && item.DueAt <= 0 {
			log.Info(ctx, "schedule type is homework", log.Any("schedule", item))
			continue
		}
		viewData := &entity.ScheduleSearchView{
			ID:      item.ID,
			StartAt: item.StartAt,
			Title:   item.Title,
			EndAt:   item.EndAt,
		}
		if v, ok := basicInfo[item.ID]; ok {
			viewData.ScheduleBasic = *v
		}

		result = append(result, viewData)
	}

	return total, result, nil
}

func (s *scheduleModel) Query(ctx context.Context, condition *da.ScheduleCondition, loc *time.Location) ([]*entity.ScheduleListView, error) {
	cacheData, err := da.GetScheduleRedisDA().SearchToListView(ctx, condition)
	if err == nil && len(cacheData) > 0 {
		log.Debug(ctx, "Query:using cache",
			log.Any("condition", condition),
			log.Any("cacheData", cacheData),
		)
		for _, item := range cacheData {
			item.Status = item.Status.GetScheduleStatus(item.EndAt)
		}
		return cacheData, nil
	}
	var scheduleList []*entity.Schedule
	err = da.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "schedule query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	result := make([]*entity.ScheduleListView, 0, len(scheduleList))
	for _, item := range scheduleList {
		//if item.ClassType == entity.ScheduleClassTypeHomework && item.DueAt > 0 {
		//	item.StartAt = utils.TodayZeroByTimeStamp(item.DueAt, loc).Unix()
		//	item.EndAt = utils.TodayEndByTimeStamp(item.DueAt, loc).Unix()
		//}
		temp := &entity.ScheduleListView{
			ID:           item.ID,
			Title:        item.Title,
			StartAt:      item.StartAt,
			EndAt:        item.EndAt,
			IsRepeat:     item.RepeatID != "",
			LessonPlanID: item.LessonPlanID,
			Status:       item.Status.GetScheduleStatus(item.EndAt),
			ClassType:    item.ClassType,
			ClassID:      item.ClassID,
			DueAt:        item.DueAt,
		}
		result = append(result, temp)
	}
	err = da.GetScheduleRedisDA().Add(ctx, condition, result)
	if err != nil {
		log.Info(ctx, "schedule Query:GetScheduleRedisDA.AddList error", log.Err(err))
	}

	return result, nil
}

func (s *scheduleModel) getBasicInfo(ctx context.Context, operator *entity.Operator, schedules []*entity.Schedule) (map[string]*entity.ScheduleBasic, error) {
	scheduleBasicMap := make(map[string]*entity.ScheduleBasic)
	if len(schedules) == 0 {
		return scheduleBasicMap, nil
	}
	var (
		classIDs      []string
		classMap      map[string]*entity.ScheduleShortInfo
		subjectIDs    []string
		subjectMap    map[string]*entity.ScheduleShortInfo
		programIDs    []string
		programMap    map[string]*entity.ScheduleShortInfo
		scheduleIDs   []string
		lessonPlanIDs []string
		lessonPlanMap map[string]*entity.ScheduleShortInfo
	)
	for _, item := range schedules {
		classIDs = append(classIDs, item.ClassID)
		subjectIDs = append(subjectIDs, item.SubjectID)
		programIDs = append(programIDs, item.ProgramID)
		scheduleIDs = append(scheduleIDs, item.ID)
		lessonPlanIDs = append(lessonPlanIDs, item.LessonPlanID)
	}
	classIDs = utils.SliceDeduplication(classIDs)
	classMap, err := s.getClassInfoMapByClassIDs(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "getBasicInfo:get class info error", log.Err(err), log.Strings("classIDs", classIDs))
		return nil, err
	}
	classTeachers, err := external.GetTeacherServiceProvider().GetByClasses(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "getBasicInfo:GetTeacherServiceProvider.GetByClasses error", log.Err(err), log.Strings("classIDs", classIDs))
		return nil, err
	}
	classStudents, err := external.GetStudentServiceProvider().GetByClassIDs(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "getBasicInfo:GetTeacherServiceProvider.GetByClasses error", log.Err(err), log.Strings("classIDs", classIDs))
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
	lessonPlanMap, err = s.getLessonPlanMapByLessonPlanIDs(ctx, dbo.MustGetDB(ctx), lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "getBasicInfo:get lesson plan info error", log.Err(err), log.Any("lessonPlanIDs", lessonPlanIDs))
		return nil, err
	}

	for _, item := range schedules {
		scheduleBasic := &entity.ScheduleBasic{}
		if v, ok := classMap[item.ClassID]; ok {
			scheduleBasic.Class = v
		}
		if v, ok := subjectMap[item.SubjectID]; ok {
			scheduleBasic.Subject = v
		}
		if v, ok := programMap[item.ProgramID]; ok {
			scheduleBasic.Program = v
		}
		if v, ok := lessonPlanMap[item.LessonPlanID]; ok {
			scheduleBasic.LessonPlan = v
		}

		if v, ok := classTeachers[item.ClassID]; ok {
			scheduleBasic.MemberTeachers = make([]*entity.ScheduleShortInfo, len(v))
			for i, t := range v {
				scheduleBasic.MemberTeachers[i] = &entity.ScheduleShortInfo{
					ID:   t.ID,
					Name: t.Name,
				}
			}
		}
		scheduleBasic.Members = scheduleBasic.MemberTeachers

		if v, ok := classStudents[item.ClassID]; ok {
			scheduleBasic.StudentCount = len(v)
			for _, t := range v {
				scheduleBasic.Members = append(scheduleBasic.Members, &entity.ScheduleShortInfo{
					ID:   t.ID,
					Name: t.Name,
				})
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

func (s *scheduleModel) getClassInfoMapByClassIDs(ctx context.Context, operator *entity.Operator, classIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	var classMap = make(map[string]*entity.ScheduleShortInfo)
	if len(classIDs) != 0 {
		classService := external.GetClassServiceProvider()
		classInfos, err := classService.BatchGet(ctx, operator, classIDs)
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetClassServiceProvider BatchGet error", log.Err(err), log.Strings("classIDs", classIDs))
			return nil, err
		}
		for _, item := range classInfos {
			if item != nil {
				classMap[item.ID] = &entity.ScheduleShortInfo{
					ID:   item.ID,
					Name: item.Name,
				}
			}
		}
	}
	return classMap, nil
}

func (s *scheduleModel) geSubjectInfoMapBySubjectIDs(ctx context.Context, subjectIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	var subjectMap = make(map[string]*entity.ScheduleShortInfo)
	if len(subjectIDs) != 0 {
		subjectIDs = utils.SliceDeduplication(subjectIDs)
		subjectInfos, err := GetSubjectModel().Query(ctx, &da.SubjectCondition{
			IDs: entity.NullStrings{
				Strings: subjectIDs,
				Valid:   len(subjectIDs) != 0,
			},
		})
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
		programInfos, err := GetProgramModel().Query(ctx, &da.ProgramCondition{
			IDs: entity.NullStrings{
				Strings: programIDs,
				Valid:   len(programIDs) != 0,
			},
		})
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

func (s *scheduleModel) GetByID(ctx context.Context, operator *entity.Operator, id string) (*entity.ScheduleDetailsView, error) {
	realTimeData, err := s.GetScheduleRealTimeStatus(ctx, operator, id)
	if err != nil {
		log.Error(ctx, "GetByID using cache:GetScheduleRealTimeStatus error",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("scheduleID", id),
		)
		return nil, err
	}
	cacheData, err := da.GetScheduleRedisDA().GetByIDs(ctx, []string{id})
	if err == nil && len(cacheData) > 0 {
		log.Debug(ctx, "GetByID:using cache",
			log.Any("id", id),
			log.Any("cacheData", cacheData),
		)
		data := cacheData[0]
		data.Status = data.Status.GetScheduleStatus(data.EndAt)
		data.RealTimeStatus = *realTimeData
		return data, nil
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
		ID:             schedule.ID,
		Title:          schedule.Title,
		OrgID:          schedule.OrgID,
		StartAt:        schedule.StartAt,
		EndAt:          schedule.EndAt,
		IsAllDay:       schedule.IsAllDay,
		ClassType:      schedule.ClassType,
		DueAt:          schedule.DueAt,
		Description:    schedule.Description,
		Version:        schedule.ScheduleVersion,
		IsRepeat:       schedule.RepeatID != "",
		Status:         schedule.Status.GetScheduleStatus(schedule.EndAt),
		RealTimeStatus: *realTimeData,
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
	basicInfo, err := s.getBasicInfo(ctx, operator, []*entity.Schedule{schedule})
	if err != nil {
		log.Error(ctx, "getBasicInfo error", log.Err(err))
		return nil, err
	}
	if v, ok := basicInfo[result.ID]; ok {
		result.ScheduleBasic = *v
	}
	err = da.GetScheduleRedisDA().BatchAdd(ctx, []*entity.ScheduleDetailsView{result})
	if err != nil {
		log.Info(ctx, "GetByID:GetScheduleRedisDA.BatchAdd error", log.Err(err))
	}
	return result, nil
}

func (s *scheduleModel) GetTeacherByName(ctx context.Context, operator *entity.Operator, orgID, name string) ([]*external.Teacher, error) {
	teacherService := external.GetTeacherServiceProvider()
	teachers, err := teacherService.Query(ctx, operator, orgID, name)
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
	lessonPlanPastIDs, err := GetContentModel().GetPastContentIDByID(ctx, dbo.MustGetDB(ctx), lessonPlanID)
	if err != nil {
		logger.Error(ctx, "ExistScheduleByLessonPlanID:GetContentModel.GetPastContentIDByID error",
			log.Err(err),
			log.String("lessonPlanID", lessonPlanID),
		)
		return false, err
	}
	condition := da.ScheduleCondition{
		LessonPlanIDs: entity.NullStrings{
			Strings: lessonPlanPastIDs,
			Valid:   true,
		},
		EndAtGe: sql.NullInt64{
			Int64: time.Now().Unix(),
			Valid: true,
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
	result.Schedule = schedule

	return result, nil
}

func (s *scheduleModel) verifyData(ctx context.Context, operator *entity.Operator, v *entity.ScheduleVerify) error {
	// class
	classService := external.GetClassServiceProvider()
	classInfos, err := classService.BatchGet(ctx, operator, []string{v.ClassID})
	if err != nil {
		log.Error(ctx, "verifyData:GetClassServiceProvider BatchGet error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}
	for _, item := range classInfos {
		if item == nil {
			log.Error(ctx, "verifyData:GetClassServiceProvider class info not found", log.Any("ScheduleVerify", v))
			return constant.ErrRecordNotFound
		}
	}

	if v.ClassType == entity.ScheduleClassTypeTask {
		return nil
	}
	// subject
	subjectIDs := []string{v.SubjectID}
	_, err = GetSubjectModel().Query(ctx, &da.SubjectCondition{
		IDs: entity.NullStrings{
			Strings: subjectIDs,
			Valid:   len(subjectIDs) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "verifyData:GetSubjectServiceProvider BatchGet error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}
	// program
	programIDs := []string{v.ProgramID}
	_, err = GetProgramModel().Query(ctx, &da.ProgramCondition{
		IDs: entity.NullStrings{
			Strings: programIDs,
			Valid:   len(programIDs) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "verifyData:GetProgramServiceProvider BatchGet error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}

	// verify lessPlan type
	lessonPlanInfo, err := GetContentModel().GetContentNameByID(ctx, dbo.MustGetDB(ctx), v.LessonPlanID)
	if err != nil {
		log.Error(ctx, "verifyData:get lessPlan info error", log.Err(err), log.Any("ScheduleVerify", v))
		return err
	}
	if lessonPlanInfo.ContentType != entity.ContentTypePlan {
		log.Error(ctx, "verifyData:content type is not lesson", log.Any("lessonPlanInfo", lessonPlanInfo), log.Any("ScheduleVerify", v))
		return constant.ErrInvalidArgs
	}
	// verify lessPlan is valid
	err = s.VerifyLessonPlanAuthed(ctx, operator, v.LessonPlanID)

	return err
}

func (s *scheduleModel) VerifyLessonPlanAuthed(ctx context.Context, operator *entity.Operator, lessonPlanID string) error {
	// verify lessPlan is valid
	contentMap, err := GetAuthedContentRecordsModel().GetContentAuthByIDList(ctx, []string{lessonPlanID}, operator)
	if err != nil {
		log.Error(ctx, "verifyLessonPlanAuthed:GetAuthedContentRecordsModel.GetContentAuthByIDList error", log.Err(err),
			log.String("lessonPlanID", lessonPlanID),
			log.Any("operator", operator),
		)
		return err
	}
	item, ok := contentMap[lessonPlanID]
	if !ok {
		log.Error(ctx, "verifyLessonPlanAuthed:lesson plan not found", log.Any("operator", operator), log.String("lessonPlanID", lessonPlanID))
		return ErrScheduleLessonPlanUnAuthed
	}
	if item == entity.ContentUnauthed {
		log.Info(ctx, "verifyLessonPlanAuthed:lesson plan unAuthed", log.Any("operator", operator),
			log.Any("lessonInfo", item),
			log.String("lessonPlanID", lessonPlanID),
		)
		return ErrScheduleLessonPlanUnAuthed
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
	err = da.GetScheduleRedisDA().Clean(ctx, []string{id})
	if err != nil {
		log.Info(ctx, "UpdateScheduleStatus:GetScheduleRedisDA.Clean error", log.Err(err))
	}
	return nil
}

func (s *scheduleModel) GetParticipateClass(ctx context.Context, operator *entity.Operator) ([]*external.Class, error) {
	classIDs, err := da.GetScheduleDA().GetParticipateClass(ctx, dbo.MustGetDB(ctx), operator.UserID)
	if err == constant.ErrRecordNotFound {
		log.Error(ctx, "GetParticipateClass:get participate class not found", log.Err(err), log.Any("op", operator))
		return []*external.Class{}, nil
	}
	if err != nil {
		log.Error(ctx, "GetParticipateClass:get participate  class from db error", log.Err(err), log.Any("op", operator))
		return nil, err
	}
	result, err := external.GetClassServiceProvider().BatchGet(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "GetParticipateClass:batch get class from ClassServiceProvider error",
			log.Err(err),
			log.Any("op", operator),
			log.Strings("classIDs", classIDs),
		)
		return nil, err
	}
	var classes []*external.Class
	for i := range result {
		if result[i].Valid {
			classes = append(classes, &result[i].Class)
		} else {
			log.Warn(ctx, "invalid value", log.Strings("class_ids", classIDs), log.Int("index", i))
		}
	}
	return classes, nil
}

func (s *scheduleModel) GetLessonPlanByCondition(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, condition *da.ScheduleCondition) ([]*entity.ScheduleShortInfo, error) {
	lessonPlanIDs, err := da.GetScheduleDA().GetLessonPlanIDsByCondition(ctx, tx, condition)
	if err != nil {
		logger.Error(ctx, "GetLessonPlanByCondition:get lessonPlanIDs error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("operator", operator),
		)
		return nil, err
	}
	latestIDs, err := GetContentModel().GetLatestContentIDByIDList(ctx, tx, lessonPlanIDs)
	if err != nil {
		logger.Error(ctx, "GetLessonPlanByCondition:get latest lessonPlanIDs error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("operator", operator),
			log.Strings("lessonPlanIDs", lessonPlanIDs),
			log.Strings("latestIDs", latestIDs),
		)
		return nil, err
	}
	latestIDs = utils.SliceDeduplication(latestIDs)
	lessonPlanInfos, err := GetContentModel().GetContentNameByIDList(ctx, tx, latestIDs)
	if err != nil {
		logger.Error(ctx, "GetLessonPlanByCondition:get lessonPlan info error",
			log.Err(err),
			log.Strings("lessonPlanIDs", lessonPlanIDs),
			log.Strings("latestIDs", latestIDs),
			log.Any("condition", condition),
			log.Any("operator", operator),
		)
	}
	result := make([]*entity.ScheduleShortInfo, len(lessonPlanInfos))
	for i, item := range lessonPlanInfos {
		result[i] = &entity.ScheduleShortInfo{
			ID:   item.ID,
			Name: item.Name,
		}
	}
	return result, nil
}

func (s *scheduleModel) GetScheduleIDsByCondition(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, condition *entity.ScheduleIDsCondition) ([]string, error) {
	lessonPlanPastIDs, err := GetContentModel().GetPastContentIDByID(ctx, tx, condition.LessonPlanID)
	if err != nil {
		logger.Error(ctx, "GetScheduleIDsByCondition:get past lessonPlan id error",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("operator", operator),
		)
		return nil, err
	}
	daCondition := &da.ScheduleCondition{
		ClassID: sql.NullString{
			String: condition.ClassID,
			Valid:  true,
		},
		LessonPlanIDs: entity.NullStrings{
			Strings: lessonPlanPastIDs,
			Valid:   true,
		},
		StartLt: sql.NullInt64{
			Int64: condition.StartAt,
			Valid: true,
		},
	}
	var scheduleList []*entity.Schedule
	err = da.GetScheduleDA().Query(ctx, daCondition, &scheduleList)
	if err != nil {
		log.Error(ctx, "schedule query error", log.Err(err), log.Any("daCondition", daCondition))
		return nil, err
	}
	var result = make([]string, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = item.ID
	}
	return result, nil
}

func (s *scheduleModel) GetScheduleIDsByOrgID(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, orgID string) ([]string, error) {
	condition := &da.ScheduleCondition{
		OrgID: sql.NullString{
			String: orgID,
			Valid:  true,
		},
	}
	var scheduleList []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "GetScheduleIDsByOrgID:GetScheduleDA.Query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	var result = make([]string, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = item.ID
	}
	return result, nil
}

func (s *scheduleModel) StartScheduleRepeat(ctx context.Context, template *entity.Schedule, options *entity.RepeatOptions, location *time.Location) ([]*entity.Schedule, error) {
	if options == nil || !options.Type.Valid() {
		return []*entity.Schedule{template}, nil
	}

	cfg := NewRepeatConfig(options, location)
	plan, err := NewRepeatCyclePlan(ctx, template.StartAt, template.EndAt, cfg)
	if err != nil {
		log.Error(ctx, "StartScheduleRepeat:NewRepeatCyclePlan error", log.Err(err), log.Any("template", template), log.Any("cfg", cfg))
		return nil, err
	}
	endRule, err := NewEndRepeatCycleRule(options)
	if err != nil {
		log.Error(ctx, "StartScheduleRepeat:NewEndRepeatCycleRule error", log.Err(err), log.Any("template", template), log.Any("options", options))
		return nil, err
	}
	planResult, err := plan.GenerateTimeByEndRule(endRule)
	if err != nil {
		log.Error(ctx, "StartScheduleRepeat:GenerateTimeByEndRule error", log.Err(err),
			log.Any("template", template),
			log.Any("plan", plan),
			log.Any("endRule", endRule),
		)
		return nil, err
	}
	result := make([]*entity.Schedule, len(planResult))
	for i, item := range planResult {
		temp := template.Clone()
		temp.StartAt = item.Start
		temp.EndAt = item.End
		temp.ID = utils.NewID()
		result[i] = &temp
	}
	return result, nil
}

func (s *scheduleModel) QueryScheduledDates(ctx context.Context, condition *da.ScheduleCondition, loc *time.Location) ([]string, error) {
	cacheData, err := da.GetScheduleRedisDA().SearchToStrings(ctx, condition)
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
		log.Error(ctx, "GetHasScheduleDate:GetScheduleDA.Query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	dateList := make([]string, 0)
	for _, item := range scheduleList {
		if item.ClassType == entity.ScheduleClassTypeHomework && item.DueAt <= 0 {
			continue
		}
		betweenTimes := utils.DateBetweenTimeAndFormat(item.StartAt, item.EndAt, loc)
		dateList = append(dateList, betweenTimes...)
	}
	result := utils.SliceDeduplication(dateList)
	err = da.GetScheduleRedisDA().Add(ctx, condition, result)
	if err != nil {
		log.Info(ctx, "QueryScheduledDates:GetScheduleRedisDA.AddDates error", log.Err(err))
	}

	return result, nil
}

func (s *scheduleModel) GetScheduleRealTimeStatus(ctx context.Context, op *entity.Operator, id string) (*entity.ScheduleRealTimeView, error) {
	var schedule = new(entity.Schedule)
	err := da.GetScheduleDA().Get(ctx, id, schedule)
	if err == dbo.ErrRecordNotFound {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetScheduleRealTimeStatus error",
			log.Err(err),
			log.String("id", id),
			log.Any("op", op),
		)
		return nil, err
	}
	if schedule.DeleteAt != 0 {
		return nil, constant.ErrRecordNotFound
	}
	result := new(entity.ScheduleRealTimeView)
	result.ID = schedule.ID
	if schedule.ClassType != entity.ScheduleClassTypeTask {
		// lesson plan real time info
		err = s.VerifyLessonPlanAuthed(ctx, op, schedule.LessonPlanID)
		switch err {
		case nil:
			result.LessonPlanIsAuth = true
		case ErrScheduleLessonPlanUnAuthed:
			log.Info(ctx, "GetScheduleRealTimeStatus:lesson plan unAuthed",
				log.String("LessonPlanID", schedule.LessonPlanID),
				log.Any("op", op),
			)
			result.LessonPlanIsAuth = false
		default:
			log.Error(ctx, "GetScheduleRealTimeStatus:lesson plan unAuthed",
				log.String("LessonPlanID", schedule.LessonPlanID),
				log.Any("op", op),
				log.Err(err),
			)
			return nil, err
		}
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
