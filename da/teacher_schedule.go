package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ITeacherScheduleDA interface {
	Add(ctx context.Context, data *entity.TeacherSchedule) error
	BatchAdd(ctx context.Context, datalist []*entity.TeacherSchedule) error

	Update(ctx context.Context, data *entity.TeacherSchedule) error
	BatchUpdate(ctx context.Context, data []*entity.TeacherSchedule) error

	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, id []string) error

	Page(ctx context.Context, condition TeacherScheduleCondition) ([]*entity.TeacherSchedule, error)
}
