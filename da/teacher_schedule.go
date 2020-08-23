package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ITeacherScheduleDA interface {
	Add(ctx context.Context, data *entity.TeacherSchedule) error
	BatchAdd(ctx context.Context, datalist []*entity.TeacherSchedule) error

	Delete(ctx context.Context, teacherID string, scheduleID string) error
	BatchDelete(ctx context.Context, pks [][2]string) error

	Page(ctx context.Context, condition TeacherScheduleCondition) (string, []*entity.TeacherSchedule, error)
}
