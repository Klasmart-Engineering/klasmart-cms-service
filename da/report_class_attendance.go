package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
)

type IClassAttendance interface {
	GetClassAttendance(ctx context.Context, request *entity.ClassAttendanceQueryParameters) (model []*entity.ClassAttendance, err error)
}

func (r *ReportDA) GetClassAttendance(ctx context.Context, request *entity.ClassAttendanceQueryParameters) (model []*entity.ClassAttendance, err error) {
	sql := `
select s1.class_id,s1.subject_id,s2.student_id,if(s2.finish_counts > 0,true,false) as is_attendance
from 
    (
     select sr.schedule_id,
            sr.relation_id                                as class_id,
            if(sr1.relation_id is null, '', sr1.relation_id) as subject_id,
            sh.start_at,sh.class_type
            from   schedules_relations sr
            left join schedules_relations sr1 
            on sr.schedule_id = sr1.schedule_id
            inner join schedules sh
            on sr.schedule_id = sh.id
            where sr.relation_type = '${class_roster_class}' and sr1.relation_type = '${subject}'
    )
as s1
inner join
   (
    select sr.relation_id as student_id,car.class_id,car.schedule_id,if(car.finish_counts is null, 0, car.finish_counts) as finish_counts
    from schedules_relations sr 
    left join classes_assignments_records  car
    on sr.schedule_id=car.schedule_id
    and sr.relation_id=car.attendance_id
    where sr.relation_type='${class_roster_student}'
   )
as s2
on s1.schedule_id=s2.schedule_id
and s1.class_id=s2.class_id
and s1.class_type='${OnlineClass}'
where  s1.class_id=? and s1.start_at>=? and s1.start_at<? and s1.subject_id in (?)
order by s1.class_id,s1.subject_id
`
	sql = strings.Replace(sql, "${subject}", entity.ScheduleClassTypeSubject.String(), -1)
	sql = strings.Replace(sql, "${OnlineClass}", entity.ScheduleClassTypeOnlineClass.String(), -1)
	sql = strings.Replace(sql, "${class_roster_student}", entity.ScheduleRelationTypeClassRosterStudent.String(), -1)
	sql = strings.Replace(sql, "${class_roster_class}", entity.ScheduleRelationTypeClassRosterClass.String(), -1)
	startAt, endAt, err := request.Duration.Value(ctx)
	if err != nil {
		return
	}
	err = r.QueryRawSQL(ctx, &model, sql, request.ClassID, startAt, endAt, request.SubjectIDS)
	if err != nil {
		log.Error(ctx, "exec GetClassAttendance sql failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("params", request))
		return
	}
	return
}
