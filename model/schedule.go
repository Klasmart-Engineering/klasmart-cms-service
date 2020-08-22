package model

import (
	"context"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"sync"
)

type IScheduleModel interface {
	Add(ctx context.Context, op *entity.Operator, viewdata *entity.ScheduleAddView) (string, error)
	Update(ctx context.Context, op *entity.Operator, viewdata *entity.ScheduleUpdateView) error
	Delete(ctx context.Context, op *entity.Operator, id string, editType entity.ScheduleEditType) error

	Page(ctx context.Context, condition *dynamodbhelper.Condition) (string, []*entity.ScheduleListView, error)
	GetByID(ctx context.Context, id string) (*entity.ScheduleDetailsView, error)
}
type scheduleModel struct {
	testScheduleRepeatFlag bool
}

func (s *scheduleModel) Add(ctx context.Context, op *entity.Operator, viewdata *entity.ScheduleAddView) (string, error) {
	// TODO:
	// 1.verify data

	// convert to schedule
	schedule := viewdata.Convert()
	schedule.CreatedID = op.UserID
	scheduleList, err := s.RepeatSchedule(ctx, schedule)
	if err != nil {
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
	// add to schedules
	err = da.GetScheduleDA().BatchInsert(ctx, scheduleList)
	if err != nil {
		return "", err
	}

	// add to teachers_schedules
	err = da.GetTeacherScheduleDA().BatchAdd(ctx, teacherSchedules)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (s *scheduleModel) Update(ctx context.Context, op *entity.Operator, viewdata *entity.ScheduleUpdateView) error {
	// TODO: check permission
	if !viewdata.EditType.Valid() {
		err := errors.New("update schedule: invalid type")
		log.Error(ctx, err.Error(), log.String("edit_type", string(viewdata.EditType)))
		return err
	}
	if err := s.Delete(ctx, op, viewdata.ID, viewdata.EditType); err != nil {
		log.Error(ctx, "update schedule: delete failed",
			log.Err(err),
			log.String("id", viewdata.ID),
			log.String("edit_type", string(viewdata.EditType)),
		)
		return err
	}
	if _, err := s.Add(ctx, op, &viewdata.ScheduleAddView); err != nil {
		log.Error(ctx, "update schedule: delete failed",
			log.Err(err),
			log.Any("schedule_add_view", viewdata.ScheduleAddView),
		)
		return err
	}
	return nil
}

func (s *scheduleModel) Delete(ctx context.Context, op *entity.Operator, id string, editType entity.ScheduleEditType) error {
	// TODO: check permission
	switch editType {
	case entity.ScheduleEditOnlyCurrent:
		if err := da.GetScheduleDA().Delete(ctx, id); err != nil {
			log.Error(ctx, "delete schedule: delete failed",
				log.String("id", id), log.String("edit_type", string(editType)))
			return err
		}
	case entity.ScheduleEditWithFollowing:
		item, err := da.GetScheduleDA().GetByID(ctx, id)
		if err != nil {
			log.Error(ctx, "delete schedule: get by id failed", log.String("id", id))
			return err
		}
		cond := dynamodbhelper.Condition{
			PrimaryKey:  dynamodbhelper.KeyValue{Key: "repeat_id", Value: item.RepeatID},
			SortKey:     dynamodbhelper.KeyValue{Key: "start_at", Value: item.StartAt},
			CompareType: dynamodbhelper.SortKeyGreaterThanEqual,
			IndexName:   entity.Schedule{}.IndexNameRepeatIDAndStartAt(),
		}
		items, err := da.GetScheduleDA().Query(ctx, &cond)
		if err != nil {
			log.Error(ctx, "delete schedule: query failed", log.Any("cond", cond))
			return err
		}
		var ids []string
		for _, item := range items {
			ids = append(ids, item.ID)
		}
		if err = da.GetScheduleDA().BatchDelete(ctx, ids); err != nil {
			log.Error(ctx, "delete schedule: batch delete failed", log.Err(err))
			return err
		}
	default:
		err := fmt.Errorf("delete schedule: invalid edit type")
		log.Error(ctx, err.Error(), log.String("edit_type", string(editType)))
		return err
	}
	return nil
}

func (s *scheduleModel) Page(ctx context.Context, condition *dynamodbhelper.Condition) (string, []*entity.ScheduleListView, error) {
	panic("implement me")
}

func (s *scheduleModel) GetByID(ctx context.Context, id string) (*entity.ScheduleDetailsView, error) {
	schedule, err := da.GetScheduleDA().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &entity.ScheduleDetailsView{
		ID:    schedule.ID,
		Title: schedule.Title,
		//ClassID: entity.ShortInfo{
		//	ID:   schedule.ClassID,
		//	Name: external.ClassServiceProvider(),
		//},
		//LessonPlanID: entity.ShortInfo{},
		//TeacherIDs:   make([]entity.ShortInfo, len(schedule.TeacherIDs)),
		//SubjectID:    entity.ShortInfo{},
		//ProgramID:    entity.ShortInfo{},
		//AttachmentID: entity.ShortInfo{},
		OrgID:       schedule.OrgID,
		StartAt:     schedule.StartAt,
		EndAt:       schedule.EndAt,
		ModeType:    schedule.ModeType,
		ClassType:   schedule.ClassType,
		DueAt:       schedule.DueAt,
		Description: schedule.Description,
		Version:     schedule.Version,
		RepeatID:    schedule.RepeatID,
		Repeat:      schedule.Repeat,
	}
	classService, err := external.GetClassServiceProvider()
	if err != nil {
		return nil, err
	}
	classInfos, err := classService.BatchGet(ctx, []string{schedule.ClassID})
	if err != nil {
		return nil, err
	}
	if len(classInfos) > 0 {
		result.Class = entity.ShortInfo{
			ID:   classInfos[0].ID,
			Name: classInfos[0].Name,
		}
	}
	result.Teachers = make([]entity.ShortInfo, len(schedule.TeacherIDs))
	teacherService, err := external.GetTeacherServiceProvider()
	if err != nil {
		return nil, err
	}
	teacherInfos, err := teacherService.BatchGet(ctx, schedule.TeacherIDs)
	if err != nil {
		return nil, err
	}
	for i, item := range teacherInfos {
		result.Teachers[i] = entity.ShortInfo{
			ID:   item.ID,
			Name: item.Name,
		}
	}
	subjectService, err := external.GetSubjectServiceProvider()
	if err != nil {
		return nil, err
	}
	subjectInfos, err := subjectService.BatchGet(ctx, []string{schedule.SubjectID})
	if err != nil {
		return nil, err
	}
	if len(subjectInfos) > 0 {
		result.Subject = entity.ShortInfo{
			ID:   subjectInfos[0].ID,
			Name: subjectInfos[0].Name,
		}
	}
	programService, err := external.GetProgramServiceProvider()
	if err != nil {
		return nil, err
	}
	programInfos, err := programService.BatchGet(ctx, []string{schedule.ProgramID})
	if err != nil {
		return nil, err
	}
	if len(programInfos) > 0 {
		result.Program = entity.ShortInfo{
			ID:   programInfos[0].ID,
			Name: programInfos[0].Name,
		}
	}
	// TODO LessonPlan Attachment

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
