package da

import (
	"bytes"
	"context"
	"text/template"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IClassAttendance interface {
	GetClassAttendance(ctx context.Context, request *entity.ClassAttendanceQueryParameters) (model []*entity.ClassAttendance, err error)
}

func (r *ReportDA) GetClassAttendance(ctx context.Context, request *entity.ClassAttendanceQueryParameters) (model []*entity.ClassAttendance, err error) {
	sql := `
select s3.student_id,s3.class_id,s3.subject_id,if(car.finish_counts > 0,true,false) as is_attendance from 
(
	select s1.student_id,s1.schedule_id,s2.class_id,s2.subject_id,s2.start_at,s2.Title,s2.class_type from 
	(
   	select sr.relation_id as student_id,sr.schedule_id from schedules_relations sr
   	where sr.relation_type='{{.ClassRosterStudent}}' 
	) s1 left join 
	(
		select sr.schedule_id,
		sr.relation_id as class_id,
		if(sr1.relation_id is null, '', sr1.relation_id) as subject_id,
		sh.start_at,sh.class_type,sh.Title
		from schedules_relations sr
		left join schedules_relations sr1
		on sr.schedule_id = sr1.schedule_id
		inner join schedules sh
		on sr.schedule_id = sh.id
		where sr.relation_type = '{{.ClassRosterClass}}' and sr1.relation_type = '{{.Subject}}' 
	) s2
	on s1.schedule_id=s2.schedule_id 
) s3 left join classes_assignments_records car
on s3.schedule_id=car.schedule_id
and s3.student_id=car.attendance_id
and s3.class_id=car.class_id
where s3.class_type='{{.OnlineClass}}' and s3.class_id=?
and s3.start_at>=? and s3.start_at<?
and s3.subject_id in (?)
order by s3.class_id,s3.subject_id
`
	sqlTmplParameter := struct {
		ClassRosterClass   string
		Subject            string
		ClassRosterStudent string
		OnlineClass        string
	}{
		ClassRosterClass:   entity.ScheduleRelationTypeClassRosterClass.String(),
		Subject:            entity.ScheduleRelationTypeSubject.String(),
		ClassRosterStudent: entity.ScheduleRelationTypeClassRosterStudent.String(),
		OnlineClass:        entity.ScheduleClassTypeOnlineClass.String(),
	}
	sqlTmpl, err := template.New("sql").Parse(sql)
	if err != nil {
		log.Error(ctx, "create GetClassAttendance sql tmpl failed",
			log.Err(err),
			log.String("sql", sql))
		return
	}
	buffer := new(bytes.Buffer)
	err = sqlTmpl.Execute(buffer, sqlTmplParameter)
	if err != nil {
		log.Error(ctx, " GetClassAttendance replace sqlTmpl failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("params", sqlTmplParameter))
		return
	}
	startAt, endAt, err := request.Duration.Value(ctx)
	if err != nil {
		return
	}
	err = r.QueryRawSQL(ctx, &model, buffer.String(), request.ClassID, startAt, endAt, request.SubjectIDS)
	if err != nil {
		log.Error(ctx, "exec GetClassAttendance sql failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("params", request))
		return
	}
	return
}
