package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
	"sync"
)

type IReportDA interface {
	DataAccessor
	ITeacherLoadAssessment
	ITeacherLoadLesson
}
type ReportDA struct {
	BaseDA
}

var _reportDA *ReportDA
var _reportDAOnce sync.Once

func GetReportDA() IReportDA {
	_reportDAOnce.Do(func() {
		_reportDA = new(ReportDA)
	})
	return _reportDA
}

type ITeacherLoadLesson interface {
	MissedLessonsListInfo(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (model []*entity.TeacherLoadMissedLesson, err error)
	MissedLessonsListTotal(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (total int, err error)
}

func (r *ReportDA) MissedLessonsListInfo(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (model []*entity.TeacherLoadMissedLesson, err error) {
	sql := `
select * 
from 
(
	select 
		s.id,
		s.class_type, 
		s.title,
		(
            select relation_id from schedules_relations sel 
            where sel.schedule_id=s.id 
            and sel.relation_type='class_roster_class'
        ) as class_id,
        (
			select 
				count(*) 
			from schedules_relations sls 
			where  sls.schedule_id=s.id 
           	and sls.relation_type='class_roster_student'
		) as no_of_student,
		s.start_at as start_date,
		s.end_at as end_date
	from schedules s
	inner join schedules_relations sl
	on s.id=sl.schedule_id
	where sl.relation_id=@teacherId 
	and s.end_at between @startDt and @endDt
	and s.class_id in (@class_ids)
	group by sl.schedule_id
	order by s.end_at desc
) sc  
where !EXISTS(select id from assessments ass where ass.schedule_id=sc.id)
LIMIT @pageSize OFFSET @limitNumber`
	startAt, endAt, err := request.Duration.Value(ctx)
	if err != nil {
		return
	}
	params := map[string]interface{}{"teacherId": request.TeacherId, "startDt": startAt, "endDt": endAt,
		"class_ids": strings.Join(request.ClassIDs, ","),
		"pageSize":  request.PageNumber, "limitNumber": (request.PageSize - 1) * request.PageNumber}
	err = r.QueryRawSQL(ctx, &model, sql, params)
	if err != nil {
		log.Error(ctx, "exec missedLessonsListInfo sql failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("params", params))
		return
	}
	return
}
func (r *ReportDA) MissedLessonsListTotal(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (total int, err error) {
	sql := `
select count(*) from 
	(
		select s.id
        from schedules s
        inner join schedules_relations sl
        on s.id=sl.schedule_id
        where sl.relation_id=@teacherId 
        and s.end_at between @startDt and @endDt
        and s.class_id in (@class_ids)
        group by sl.schedule_id
        order by s.end_at desc
    ) sc  
where !EXISTS(select id from assessments ass where ass.schedule_id=sc.id)`
	startAt, endAt, err := request.Duration.Value(ctx)
	if err != nil {
		return
	}
	params := map[string]interface{}{"teacherId": request.TeacherId, "startDt": startAt, "endDt": endAt,
		"class_ids": strings.Join(request.ClassIDs, ","),
		"pageSize":  request.PageNumber, "limitNumber": (request.PageSize - 1) * request.PageNumber}
	err = r.QueryRawSQL(ctx, &total, sql, params)
	if err != nil {
		log.Error(ctx, "exec missedLessonsListTotal sql failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("params", params))
		return
	}
	return
}
