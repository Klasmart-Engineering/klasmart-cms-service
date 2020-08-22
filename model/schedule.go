package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"sync"
)

type IScheduleModel interface {
	Add(ctx context.Context, op *entity.Operator, viewdata *entity.ScheduleAddView) (string, error)
	Update(ctx context.Context, op *entity.Operator, viewdata *entity.ScheduleUpdateView) error
	Delete(ctx context.Context, op *entity.Operator, id string, editType entity.ScheduleEditType) error

	Page(ctx context.Context, condition *dynamodbhelper.Condition) (int64, []*entity.ScheduleListView, error)
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

func (s *scheduleModel) Page(ctx context.Context, condition *dynamodbhelper.Condition) (int64, []*entity.ScheduleListView, error) {
	panic("implement me")
}

func (s *scheduleModel) GetByID(ctx context.Context, id string) (*entity.ScheduleDetailsView, error) {
	panic("implement me")
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
