package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

const ()

type IScheduleModel interface {
	Add(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleAddView, location *time.Location) (string, error)
	Update(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleUpdateView, location *time.Location) (string, error)
	Delete(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, id string, editType entity.ScheduleEditType) error
	Query(ctx context.Context, tx *dbo.DBContext, condition *da.ScheduleCondition) ([]*entity.ScheduleListView, error)
	Page(ctx context.Context, tx *dbo.DBContext, condition *da.ScheduleCondition) (int, []*entity.ScheduleSeachView, error)
	GetByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.ScheduleDetailsView, error)
	IsScheduleConflict(ctx context.Context, op *entity.Operator, startAt int64, endAt int64) (bool, error)
	GetTeacherByName(ctx context.Context, name string) ([]*external.Teacher, error)
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

func (s *scheduleModel) addRepeatSchedule(ctx context.Context, op *entity.Operator, viewData *entity.ScheduleAddView, options *entity.RepeatOptions, location *time.Location) (string, error) {
	schedule := viewData.Convert()
	schedule.CreatedID = op.UserID
	scheduleList, err := s.RepeatSchedule(ctx, schedule, options, location)
	if err != nil {
		log.Error(ctx, "schedule repeat error", log.Err(err), log.Any("schedule", schedule), log.Any("options", options))
		return "", err
	}
	scheduleTeachers := make([]*entity.ScheduleTeacher, len(scheduleList)*len(viewData.TeacherIDs))
	index := 0
	for _, item := range scheduleList {
		item.ID = utils.NewID()
		for _, teacherID := range viewData.TeacherIDs {
			tsItem := &entity.ScheduleTeacher{
				TeacherID:  teacherID,
				ScheduleID: schedule.ID,
			}
			scheduleTeachers[index] = tsItem
			index++
		}
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// add to schedules
		_, err = da.GetScheduleDA().BatchInsert(ctx, tx, scheduleList)
		if err != nil {
			log.Error(ctx, "schedule batchInsert error", log.Err(err), log.Any("scheduleList", scheduleList))
			return err
		}

		// add to teachers_schedules
		_, err = da.GetScheduleTeacherDA().BatchInsert(ctx, tx, scheduleTeachers)
		if err != nil {
			log.Error(ctx, "teachers_schedules batchInsert error", log.Err(err), log.Any("scheduleTeachers", scheduleTeachers))
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(scheduleList) <= 0 {
		log.Error(ctx, "schedules batchInsert error,schedules is empty", log.Any("schedule", schedule), log.Any("options", options))
		return "", errors.New("schedules is empty")
	}
	return scheduleList[0].ID, nil
}
func (s *scheduleModel) Add(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleAddView, location *time.Location) (string, error) {
	// validate attachment
	if viewData.Attachment != "" {
		_, exist := storage.DefaultStorage().ExistFile(ctx, ScheduleAttachment_Storage_Partition, viewData.Attachment)
		if !exist {
			log.Info(ctx, "add schedule: attachment is not exits", log.Any("requestData", viewData))
			return "", constant.ErrFileNotFound
		}
	}

	// is force add
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
	if viewData.IsRepeat {
		return s.addRepeatSchedule(ctx, op, viewData, &viewData.Repeat, location)
	} else {
		schedule := viewData.Convert()
		schedule.CreatedID = op.UserID
		schedule.ID = utils.NewID()
		err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
			_, err := da.GetScheduleDA().InsertTx(ctx, tx, schedule)
			if err != nil {
				return err
			}
			scheduleTeachers := make([]*entity.ScheduleTeacher, len(viewData.TeacherIDs))
			for i, item := range viewData.TeacherIDs {
				scheduleTeacher := &entity.ScheduleTeacher{
					ID:         utils.NewID(),
					TeacherID:  item,
					ScheduleID: schedule.ID,
					DeletedAt:  0,
				}
				scheduleTeachers[i] = scheduleTeacher
			}
			// add to teachers_schedules
			_, err = da.GetScheduleTeacherDA().BatchInsert(ctx, tx, scheduleTeachers)
			if err != nil {
				log.Error(ctx, "schedules_teachers batchInsert error", log.Err(err), log.Any("scheduleTeachers", scheduleTeachers))
				return err
			}
			return nil
		})
		if err != nil {
			log.Error(ctx, "add schedule error", log.Err(err), log.Any("schedule", schedule))
			return "", err
		}
		return schedule.ID, nil
	}
}

func (s *scheduleModel) Update(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, viewdata *entity.ScheduleUpdateView, location *time.Location) (string, error) {
	// TODO: check permission
	if !viewdata.IsForce {
		conflict, err := s.IsScheduleConflict(ctx, operator, viewdata.StartAt, viewdata.EndAt)
		if err != nil {
			log.Error(ctx, "update schedule: check time conflict failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("viewData", viewdata),
			)
			return "", err
		}
		if conflict {
			log.Info(ctx, "update schedule: time conflict",
				log.Any("operator", operator),
				log.Any("viewData", viewdata),
			)
			return "", constant.ErrConflict
		}
	}
	var schedule entity.Schedule
	if err := da.GetScheduleDA().Get(ctx, viewdata.ID, &schedule); err != nil {
		log.Error(ctx, "update schedule: get schedule by id failed",
			log.Err(err),
			log.String("id", viewdata.ID),
			log.String("edit_type", string(viewdata.EditType)),
		)
		return "", err
	}
	var id string
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
		if err = s.Delete(ctx, tx, operator, viewdata.ID, viewdata.EditType); err != nil {
			log.Error(ctx, "update schedule: delete failed",
				log.Err(err),
				log.String("id", viewdata.ID),
				log.String("edit_type", string(viewdata.EditType)),
			)
			return err
		}
		viewdata.RepeatID = schedule.RepeatID
		id, err = s.Add(ctx, tx, operator, &viewdata.ScheduleAddView, location)
		if err != nil {
			log.Error(ctx, "update schedule: delete failed",
				log.Err(err),
				log.Any("schedule_add_view", viewdata.ScheduleAddView),
			)
			return err
		}
		return nil
	}); err != nil {
		log.Error(ctx, "update schedule: tx failed", log.Err(err))
		return "", err
	}
	return id, nil
}

func (s *scheduleModel) Delete(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, id string, editType entity.ScheduleEditType) error {
	// TODO: check permission
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		switch editType {
		case entity.ScheduleEditOnlyCurrent:
			if err := da.GetScheduleDA().SoftDelete(ctx, tx, id, op); err != nil {
				log.Error(ctx, "delete schedule: soft delete failed",
					log.String("id", id),
					log.String("edit_type", string(editType)),
				)
				return err
			}
		case entity.ScheduleEditWithFollowing:
			var schedule entity.Schedule
			if err := da.GetScheduleDA().Get(ctx, id, &schedule); err != nil {
				if err == dbo.ErrRecordNotFound {
					log.Info(ctx, "delete schedule: get schedule by id failed",
						log.Err(err),
						log.String("id", id),
						log.String("edit_type", string(editType)),
					)
					return nil
				}
				log.Error(ctx, "delete schedule: get schedule by id failed",
					log.Err(err),
					log.String("id", id),
					log.String("edit_type", string(editType)),
				)
				return err
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
		if err := da.GetScheduleTeacherDA().DeleteByScheduleID(ctx, tx, id); err != nil {
			log.Error(ctx, "delete schedule: delete by schedule id failed",
				log.Err(err),
				log.String("id", id),
			)
			return err
		}
		return nil
	}); err != nil {
		log.Error(ctx, "delete schedule: tx failed", log.Err(err))
	}
	return nil
}

func (s *scheduleModel) Page(ctx context.Context, tx *dbo.DBContext, condition *da.ScheduleCondition) (int, []*entity.ScheduleSeachView, error) {
	var scheduleList []*entity.Schedule
	total, err := da.GetScheduleDA().Page(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "Page: schedule query error", log.Err(err), log.Any("condition", condition))
		return 0, nil, err
	}

	result := make([]*entity.ScheduleSeachView, len(scheduleList))
	basicInfo, err := s.getBasicInfo(ctx, tx, scheduleList)
	for i, item := range scheduleList {
		viewData := &entity.ScheduleSeachView{
			ID:      item.ID,
			StartAt: item.StartAt,
			EndAt:   item.EndAt,
		}
		if v, ok := basicInfo[item.ID]; ok {
			viewData.ScheduleBasic = *v
		}
		result[i] = viewData
	}

	return total, result, nil
}

func (s *scheduleModel) Query(ctx context.Context, tx *dbo.DBContext, condition *da.ScheduleCondition) ([]*entity.ScheduleListView, error) {
	var scheduleList []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "schedule query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	result := make([]*entity.ScheduleListView, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = &entity.ScheduleListView{
			ID:      item.ID,
			Title:   item.Title,
			StartAt: item.StartAt,
			EndAt:   item.EndAt,
		}
	}
	return result, nil
}

func (s *scheduleModel) getBasicInfo(ctx context.Context, tx *dbo.DBContext, schedules []*entity.Schedule) (map[string]*entity.ScheduleBasic, error) {
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
		scheduleTeacherMap = make(map[string][]string)
		for i, item := range scheduleTeacherList {
			teacherIDs[i] = item.TeacherID
			if _, ok := scheduleTeacherMap[item.ScheduleID]; !ok {
				scheduleTeacherMap[item.ScheduleID] = []string{}
			}
			scheduleTeacherMap[item.ScheduleID] = append(scheduleTeacherMap[item.ScheduleID], item.TeacherID)
		}

		teacherService, err := external.GetTeacherServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetTeacherServiceProvider error", log.Err(err), log.Any("schedules", schedules))
			return nil, err
		}
		teacherInfos, err := teacherService.BatchGet(ctx, teacherIDs)
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetTeacherServiceProvider BatchGet error", log.Err(err), log.Any("schedules", schedules))
			return nil, err
		}
		teacherMap = make(map[string]*entity.ScheduleShortInfo)
		for _, item := range teacherInfos {
			teacherMap[item.ID] = &entity.ScheduleShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	lessonPlans, err := GetContentModel().GetContentNameByIdList(ctx, tx, lessonPlanIDs)
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

func (s *scheduleModel) getClassInfoMapByClassIDs(ctx context.Context, classIDs []string) (map[string]*entity.ScheduleShortInfo, error) {
	var classMap = make(map[string]*entity.ScheduleShortInfo)
	if len(classIDs) != 0 {
		classService, err := external.GetClassServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetClassServiceProvider error", log.Err(err), log.Strings("classIDs", classIDs))
			return nil, err
		}
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
		subjectService, err := external.GetSubjectServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetSubjectServiceProvider error", log.Err(err), log.Strings("subjectIDs", subjectIDs))
			return nil, err
		}
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
		programService, err := external.GetProgramServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetProgramServiceProvider error", log.Err(err), log.Strings("programIDs", programIDs))
			return nil, err
		}
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

func (s *scheduleModel) GetByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.ScheduleDetailsView, error) {
	var schedule *entity.Schedule
	err := da.GetScheduleDA().Get(ctx, id, schedule)
	if err != nil {
		log.Error(ctx, "GetByID error", log.Err(err), log.String("id", id))
		return nil, err
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
		Version:     schedule.Version,
		RepeatID:    schedule.RepeatID,
	}
	if schedule.RepeatJson != "" {
		var repeat entity.RepeatOptions
		err := json.Unmarshal([]byte(schedule.RepeatJson), repeat)
		if err != nil {
			log.Error(ctx, "Unmarshal schedule.RepeatJson error", log.Err(err), log.String("schedule.RepeatJson", schedule.RepeatJson))
			return nil, err
		}
		result.Repeat = repeat
	}
	basicInfo, err := s.getBasicInfo(ctx, tx, []*entity.Schedule{schedule})
	if err != nil {
		log.Error(ctx, "getBasicInfo error", log.Err(err))
		return nil, err
	}
	if v, ok := basicInfo[result.ID]; ok {
		result.ScheduleBasic = *v
	}
	return result, nil
}

func (s *scheduleModel) GetTeacherByName(ctx context.Context, name string) ([]*external.Teacher, error) {
	teacherService, err := external.GetTeacherServiceProvider()
	if err != nil {
		log.Error(ctx, "querySchedule:get teacher service provider error", log.Err(err), log.String("name", name))
		return nil, err
	}
	teachers, err := teacherService.Query(ctx, name)
	if err != nil {
		log.Error(ctx, "querySchedule:query teacher info error", log.Err(err), log.String("name", name))
		return nil, err
	}

	return teachers, nil
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
