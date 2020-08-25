package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IScheduleModel interface {
	Add(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error)
	Update(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleUpdateView) (string, error)
	Delete(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, id string, editType entity.ScheduleEditType) error
	Query(ctx context.Context, tx *dbo.DBContext, condition *da.ScheduleCondition) ([]*entity.ScheduleListView, error)
	Page(ctx context.Context, tx *dbo.DBContext, condition *da.ScheduleCondition) (int, []*entity.ScheduleSeachView, error)
	//PageByTeacherID(ctx context.Context, tx *dbo.DBContext, condition *da.ScheduleCondition) (int, []*entity.ScheduleSeachView, error)
	GetByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.ScheduleDetailsView, error)
	IsScheduleConflict(ctx context.Context, op *entity.Operator, startAt int64, endAt int64) (bool, error)
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
		return true, nil
	}
	return false, nil
}

func (s *scheduleModel) addRepeatSchedule(ctx context.Context, tx *dbo.DBContext, schedule *entity.Schedule, options entity.RepeatOptions) (string, error) {
	scheduleList, err := s.RepeatSchedule(ctx, schedule, options)
	if err != nil {
		log.Error(ctx, "schedule repeat error", log.Err(err), log.Any("schedule", schedule))
		return "", err
	}
	teacherSchedules := make([]*entity.TeacherSchedule, len(scheduleList)*len(schedule.TeacherIDs))
	index := 0
	for _, item := range scheduleList {
		item.ID = utils.NewID()
		for _, teacherID := range item.TeacherIDs {
			tsItem := &entity.TeacherSchedule{
				TeacherID:  teacherID,
				ScheduleID: schedule.ID,
				StartAt:    schedule.StartAt,
			}
			teacherSchedules[index] = tsItem
			index++
		}
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// add to schedules
		_, err = da.GetScheduleDA().BatchInsert(ctx, tx, scheduleList)
		if err != nil {
			log.Error(ctx, "schedule batchInsert error", log.Err(err))
			return err
		}

		// add to teachers_schedules
		_, err = da.GetScheduleTeacherDA().BatchInsert(ctx, tx, teacherSchedules)
		if err != nil {
			log.Error(ctx, "teachers_schedules batchInsert error", log.Err(err), log.Any("teacherSchedules", teacherSchedules))
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(scheduleList) > 0 {
		return scheduleList[0].ID, errors.New("")
	}
	return "", errors.New("")
}
func (s *scheduleModel) Add(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewData *entity.ScheduleAddView) (string, error) {
	schedule := viewData.Convert()
	schedule.CreatedID = op.UserID
	if viewData.ModeType == entity.ModeTypeRepeat {
		return s.addRepeatSchedule(ctx, tx, schedule, viewData.Repeat)
	} else {
		schedule.ID = utils.NewID()
		err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
			_, err := da.GetScheduleDA().InsertTx(ctx, tx, schedule)
			if err != nil {
				return err
			}
			teacherSchedules := make([]*entity.TeacherSchedule, len(schedule.TeacherIDs))
			for i, item := range viewData.TeacherIDs {
				teacherSchedule := &entity.TeacherSchedule{
					ID:         utils.NewID(),
					TeacherID:  item,
					ScheduleID: schedule.ID,
					DeletedAt:  0,
				}
				teacherSchedules[i] = teacherSchedule
			}
			// add to teachers_schedules
			_, err = da.GetScheduleTeacherDA().BatchInsert(ctx, tx, teacherSchedules)
			if err != nil {
				log.Error(ctx, "teachers_schedules batchInsert error", log.Err(err), log.Any("teacherSchedules", teacherSchedules))
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

func (s *scheduleModel) Update(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewdata *entity.ScheduleUpdateView) (string, error) {
	// TODO: check permission
	if !viewdata.EditType.Valid() {
		err := errors.New("update schedule: invalid type")
		log.Info(ctx, err.Error(), log.String("edit_type", string(viewdata.EditType)))
		return "", entity.ErrInvalidArgs(err)
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
		if err = s.Delete(ctx, tx, op, viewdata.ID, viewdata.EditType); err != nil {
			log.Error(ctx, "update schedule: delete failed",
				log.Err(err),
				log.String("id", viewdata.ID),
				log.String("edit_type", string(viewdata.EditType)),
			)
			return err
		}
		id, err = s.Add(ctx, tx, op, &viewdata.ScheduleAddView)
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
					log.Warn(ctx, "delete schedule: get schedule by id failed",
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
			}
		default:
			err := errors.New("delete schedule: invalid edit type")
			log.Info(ctx, err.Error(), log.String("edit_type", string(editType)))
			return err
		}
		if err := da.GetScheduleTeacherDA().DeleteByScheduleID(ctx, tx, id); err != nil {
			log.Error(ctx, "delete schedule: delete by schedule id failed",
				log.Err(err),
				log.String("id", id),
			)
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
	for i, item := range scheduleList {
		viewData := &entity.ScheduleSeachView{
			ID:      item.ID,
			StartAt: item.StartAt,
			EndAt:   item.EndAt,
		}
		basicInfo, err := s.getBasicInfo(ctx, tx, item)
		if err != nil {
			log.Error(ctx, "Page:getBasicInfo error", log.Err(err), log.Any("scheduleItem", item))
			return 0, nil, err
		}
		viewData.ScheduleBasic = *basicInfo
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

func (s *scheduleModel) getBasicInfo(ctx context.Context, tx *dbo.DBContext, schedule *entity.Schedule) (*entity.ScheduleBasic, error) {
	result := &entity.ScheduleBasic{}
	if schedule.ClassID != "" {
		classService, err := external.GetClassServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetClassServiceProvider error", log.Err(err), log.Any("schedule", schedule))
			return nil, err
		}
		classInfos, err := classService.BatchGet(ctx, []string{schedule.ClassID})
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetClassServiceProvider BatchGet error", log.Err(err), log.Any("schedule", schedule))
			return nil, err
		}
		if len(classInfos) > 0 {
			result.Class = entity.ShortInfo{
				ID:   classInfos[0].ID,
				Name: classInfos[0].Name,
			}
		}
	}

	if schedule.SubjectID != "" {
		subjectService, err := external.GetSubjectServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetSubjectServiceProvider error", log.Err(err), log.Any("schedule", schedule))
			return nil, err
		}
		subjectInfos, err := subjectService.BatchGet(ctx, []string{schedule.SubjectID})
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetSubjectServiceProvider BatchGet error", log.Err(err), log.Any("schedule", schedule))
			return nil, err
		}
		if len(subjectInfos) > 0 {
			result.Subject = entity.ShortInfo{
				ID:   subjectInfos[0].ID,
				Name: subjectInfos[0].Name,
			}
		}
	}
	if schedule.ProgramID != "" {
		programService, err := external.GetProgramServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetProgramServiceProvider error", log.Err(err), log.Any("schedule", schedule))
			return nil, err
		}
		programInfos, err := programService.BatchGet(ctx, []string{schedule.ProgramID})
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetProgramServiceProvider BatchGet error", log.Err(err), log.Any("schedule", schedule))
			return nil, err
		}
		if len(programInfos) > 0 {
			result.Program = entity.ShortInfo{
				ID:   programInfos[0].ID,
				Name: programInfos[0].Name,
			}
		}
	}
	var scheduleTeacherList []*entity.TeacherSchedule
	err := da.GetScheduleTeacherDA().Query(ctx, &da.ScheduleTeacherCondition{
		ScheduleID: sql.NullString{
			String: schedule.ID,
			Valid:  true,
		},
	}, &scheduleTeacherList)
	if err != nil {
		return nil, err
	}
	if len(scheduleTeacherList) != 0 {
		teacherIDs := make([]string, len(scheduleTeacherList))
		for _, item := range scheduleTeacherList {
			teacherIDs = append(teacherIDs, item.TeacherID)
		}
		result.Teachers = make([]entity.ShortInfo, len(teacherIDs))
		teacherService, err := external.GetTeacherServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetTeacherServiceProvider error", log.Err(err), log.Any("schedule", schedule))
			return nil, err
		}
		teacherInfos, err := teacherService.BatchGet(ctx, teacherIDs)
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetTeacherServiceProvider BatchGet error", log.Err(err), log.Any("schedule", schedule))
			return nil, err
		}
		for i, item := range teacherInfos {
			result.Teachers[i] = entity.ShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	// TODO LessonPlan

	return result, nil
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
		ModeType:    schedule.ModeType,
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
	basicInfo, err := s.getBasicInfo(ctx, tx, schedule)
	if err != nil {
		log.Error(ctx, "getBasicInfo error", log.Err(err))
		return nil, err
	}
	result.ScheduleBasic = *basicInfo
	return result, nil
}

//func (s *scheduleModel) Page(ctx context.Context, tx *dbo.DBContext, condition *da.ScheduleCondition) (int, []*entity.ScheduleSeachView, error) {
//	var scheduleList []*entity.Schedule
//	total, err := da.GetScheduleDA().Page(ctx, condition, &scheduleList)
//	if err != nil {
//		return 0, nil, err
//	}
//	var result = make([]*entity.ScheduleSeachView, len(scheduleList))
//	for i, item := range scheduleList {
//		baseInfo, err := s.getBasicInfo(ctx, tx, item)
//		if err != nil {
//			return 0, nil, err
//		}
//		result[i] = &entity.ScheduleSeachView{
//			ID:            item.ID,
//			StartAt:       item.StartAt,
//			EndAt:         item.EndAt,
//			ScheduleBasic: *baseInfo,
//		}
//	}
//	return total, result, nil
//}
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
