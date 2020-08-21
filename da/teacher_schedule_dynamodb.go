package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"sync"
)

type teacherScheduleDA struct{}

func (t teacherScheduleDA) Add(ctx context.Context, data *entity.TeacherSchedule) error {
	panic("implement me")
}

func (t teacherScheduleDA) BatchAdd(ctx context.Context, datalist []*entity.TeacherSchedule) error {
	panic("implement me")
}

func (t teacherScheduleDA) Update(ctx context.Context, data *entity.TeacherSchedule) error {
	panic("implement me")
}

func (t teacherScheduleDA) BatchUpdate(ctx context.Context, data []*entity.TeacherSchedule) error {
	panic("implement me")
}

func (t teacherScheduleDA) Delete(ctx context.Context, id string) error {
	panic("implement me")
}

func (t teacherScheduleDA) BatchDelete(ctx context.Context, id string) error {
	panic("implement me")
}

func (t teacherScheduleDA) Page(ctx context.Context, condition dynamodbhelper.Condition) ([]*entity.TeacherSchedule, error) {
	panic("implement me")
}

var (
	_teacherScheduleOnce sync.Once
	_teacherScheduleDA   ITeacherScheduleDA
)

func GetTeacherScheduleDA() ITeacherScheduleDA {
	_teacherScheduleOnce.Do(func() {
		_teacherScheduleDA = teacherScheduleDA{}
	})
	return _teacherScheduleDA
}
