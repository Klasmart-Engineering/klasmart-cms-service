package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IStudentUsage interface {
	DataAccessor
	IStudentUsageMaterial
}
type IStudentUsageMaterial interface {
	GetMaterialUsages(ctx context.Context, req *entity.StudentUsageMaterialViewCountReportRequest) (usages []*entity.MaterialUsage, err error)
}

var _studentUsageOnce sync.Once

type studentUsageDA struct {
	BaseDA
}

func (s studentUsageDA) GetMaterialUsages(ctx context.Context, req *entity.StudentUsageMaterialViewCountReportRequest) (usages []*entity.MaterialUsage, err error) {
	sql := `
select content_type,count(1) as used_count from (
	select
	DISTINCT student_user_id,
	lesson_plan_id,
	lesson_material_id,
	class_id,
	content_type
FROM
	student_usage_records
where
	class_id in (?)
	and content_type in (?)
	and (%s)
) t group by t.content_type
`
	args := []interface{}{
		req.ClassIDList,
		req.ContentTypeList,
	}
	if len(req.TimeRangeList) < 1 {
		return
	}

	timeRangePlaceHolder := " 1=1 "
	for _, timeRange := range req.TimeRangeList {
		var min, max int64
		min, max, err = timeRange.Value(ctx)
		if err != nil {
			return
		}
		timeRangePlaceHolder += "or schedule_start_at BETWEEN ? and ?"
		args = append(args, min, max)
	}

	db := dbo.MustGetDB(ctx)
	db.Exec(sql, args)

	panic("implement me")
}

var _studentUsageDA *studentUsageDA

func GetStudentUsageDA() IStudentUsage {
	_studentUsageOnce.Do(func() {
		_studentUsageDA = new(studentUsageDA)
	})
	return _studentUsageDA
}
