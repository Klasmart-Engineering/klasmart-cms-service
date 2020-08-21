package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"sync"
)

type IScheduleModel interface {
	Add(ctx context.Context, op *entity.Operator, viewdata *entity.ScheduleAddView) (string, error)
	Update(ctx context.Context, op *entity.Operator, viewdata *entity.ScheduleUpdateView) error
	Delete(ctx context.Context, op *entity.Operator, id string) error

	Page(ctx context.Context, condition *dynamodbhelper.Condition) (int64, []*entity.ScheduleListView, error)
	GetByID(ctx context.Context, id string) (*entity.ScheduleDetailsView, error)
}
type scheduleModel struct{}

func (s *scheduleModel) Add(ctx context.Context, op *entity.Operator, viewdata *entity.ScheduleAddView) (string, error) {
	// TODO: verify data

	// convert to schedule
	schedule := viewdata.Convert()
	schedule.CreatedID = op.UserID
	scheduleList, err := RepeatSchedule(ctx, schedule)
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
	// add to teachers_schedules
	err = da.GetTeacherScheduleDA().BatchAdd(ctx, teacherSchedules)
	if err != nil {
		return "", err
	}
	// add to schedules
	err = da.GetScheduleDA().BatchInsert(ctx, scheduleList)
	if err != nil {
		return "", err
	}

	return "", nil
}

func (s *scheduleModel) Update(ctx context.Context, op *entity.Operator, viewdata *entity.ScheduleUpdateView) error {
	panic("implement me")
}

func (s *scheduleModel) Delete(ctx context.Context, op *entity.Operator, id string) error {
	panic("implement me")
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
