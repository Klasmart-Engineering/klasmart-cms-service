package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

type ITeacherLoadLessonsDA interface {
	QueryTx(ctx context.Context, tx *dbo.DBContext, condition *TeacherLoadLessonsCondition) (interface{}, error)
}

var (
	_teacherLoadLessonsDA *TeacherLoadLessonsSQLDA

	_teacherLoadLessonsOnce sync.Once
)

func GetTeacherLoadLessonsDA() ITeacherLoadLessonsDA {
	_teacherLoadLessonsOnce.Do(func() {
		_teacherLoadLessonsDA = new(TeacherLoadLessonsSQLDA)
	})
	return _teacherLoadLessonsDA
}
