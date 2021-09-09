package da

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IStudentUsage interface {
	DataAccessor
	IStudentUsageMaterial
}
type IStudentUsageMaterial interface {
	GetMaterialViewCountUsages(ctx context.Context, req *entity.StudentUsageMaterialViewCountReportRequest) (usages []*entity.MaterialUsage, err error)
	GetMaterialUsages(ctx context.Context, req *entity.StudentUsageMaterialReportRequest) (usages []*entity.MaterialUsage, err error)
}

var _studentUsageOnce sync.Once

type studentUsageDA struct {
	BaseDA
}

func (s studentUsageDA) makePlaceHolderForStringSlice(slice []string) (pl string) {
	return strings.TrimRight(strings.Repeat("?,", len(slice)), ",")
}
func (s studentUsageDA) appendStringSlice(args []interface{}, slice []string) []interface{} {
	for _, s2 := range slice {
		args = append(args, s2)
	}
	return args
}
func (s studentUsageDA) GetMaterialUsages(ctx context.Context, req *entity.StudentUsageMaterialReportRequest) (usages []*entity.MaterialUsage, err error) {
	usages = make([]*entity.MaterialUsage, 0)
	if len(req.TimeRangeList) < 1 {
		return
	}

	var sqlArr []string
	var args []interface{}

	for _, timeRange := range req.TimeRangeList.Slice() {
		var min, max int64
		min, max, err = timeRange.Value(ctx)
		if err != nil {
			return
		}

		sqlArr = append(sqlArr, fmt.Sprintf(`
select class_id,content_type,count(1) as used_count,? as time_range from (
	select
	DISTINCT student_user_id,
	lesson_plan_id,
	lesson_material_id,
	class_id,
	content_type
FROM
	student_usage_records
where
	class_id in (%s)
	and content_type in (%s)
	and schedule_start_at BETWEEN ? and ?
) t group by t.class_id,t.content_type
`, s.makePlaceHolderForStringSlice(req.ClassIDList), s.makePlaceHolderForStringSlice(req.ContentTypeList)))
		args = append(args, timeRange)
		args = s.appendStringSlice(args, req.ClassIDList)
		args = s.appendStringSlice(args, req.ContentTypeList)
		args = append(args, min, max)
	}
	sql := strings.Join(sqlArr, " union all ")
	err = s.exec(ctx, sql, args, &usages)
	if err != nil {
		return
	}

	return
}
func (s studentUsageDA) GetMaterialViewCountUsages(ctx context.Context, req *entity.StudentUsageMaterialViewCountReportRequest) (usages []*entity.MaterialUsage, err error) {
	usages = make([]*entity.MaterialUsage, 0)
	if len(req.TimeRangeList) < 1 {
		return
	}

	var args []interface{}
	args = s.appendStringSlice(args, req.ClassIDList)
	args = s.appendStringSlice(args, req.ContentTypeList)

	var pls []string
	for _, timeRange := range req.TimeRangeList.Slice() {
		var min, max int64
		min, max, err = timeRange.Value(ctx)
		if err != nil {
			return
		}
		pls = append(pls, "schedule_start_at BETWEEN ? and ?")
		args = append(args, min, max)
	}
	sql := fmt.Sprintf(`
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
	class_id in (%s)
	and content_type in (%s)
	and (%s)
) t group by t.content_type
`, s.makePlaceHolderForStringSlice(req.ClassIDList), s.makePlaceHolderForStringSlice(req.ContentTypeList), strings.Join(pls, " or "))

	err = s.exec(ctx, sql, args, &usages)
	if err != nil {
		return
	}
	return
}

func (s studentUsageDA) exec(ctx context.Context, sql string, args []interface{}, result interface{}) (err error) {
	defer func() {
		log.Info(
			ctx,
			"exec",
			log.Any("sql", sql),
			log.Any("args", args),
			log.Any("result", result),
			log.Any("err", err),
		)
	}()
	db := dbo.MustGetDB(ctx)
	log.Info(ctx, "start execute sql", log.Any("sql", sql), log.Any("args", args))
	err = db.Raw(sql, args...).Scan(result).Error
	if err != nil {
		log.Error(ctx, "execute sql error", log.Err(err), log.Any("sql", sql), log.Any("args", args))
		return
	}
	return
}

var _studentUsageDA *studentUsageDA

func GetStudentUsageDA() IStudentUsage {
	_studentUsageOnce.Do(func() {
		_studentUsageDA = new(studentUsageDA)
	})
	return _studentUsageDA
}
