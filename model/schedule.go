package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"strings"
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

	// add to teachers_schedules
	teacherSchedules := make([]*entity.TeacherSchedule, len(schedule.TeacherIDs))
	for i, teacherID := range schedule.TeacherIDs {
		if strings.TrimSpace(teacherID) == "" {
			continue
		}
		tsItem := &entity.TeacherSchedule{
			TeacherID:  teacherID,
			ScheduleID: schedule.ID,
			StartAt:    schedule.StartAt,
		}
		teacherSchedules[i] = tsItem
	}
	err := da.GetTeacherScheduleDA().BatchAdd(ctx, teacherSchedules)
	if err != nil {
		return "", err
	}
	// add to schedules

	//

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
