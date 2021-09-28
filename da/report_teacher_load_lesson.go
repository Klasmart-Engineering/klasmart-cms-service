package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
)

type ITeacherLoadLesson interface {
	ListTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error)
	SummaryTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummaryFields, error)
	MissedLessonsListInfo(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (model []*entity.TeacherLoadMissedLesson, err error)
	MissedLessonsListTotal(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (total int, err error)
}

func (r *ReportDA) ListTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error) {
	sql := `
select
       teacher_id,
       count(if(is_attended=1 and class_type='${OnlineClass}', 1, null)) as live_completed_count,
       count(if(is_attended=1 and class_type='${OfflineClass}', 1, null)) as in_class_completed_count,
       count(if(is_attended=0 and class_type='${OnlineClass}', 1, null)) as live_missed_count,
       count(if(is_attended=0 and class_type='${OfflineClass}', 1, null)) as in_class_missed_count,
       count(1) as total_schedule
from (
     select distinct tsa.teacher_id, tsa.schedule_id, tsa.duration, tsa.class_type, if(assessments_attendances.attendance_id is null, 0, 1) as is_attended from
         (select teacher_schedule.teacher_id as teacher_id, teacher_schedule.schedule_id, teacher_schedule.duration, teacher_schedule.class_type, assessments.id as assessment_id from
             (select teacher.id as teacher_id, s.id as schedule_id, s.duration, s.class_type from
                 (select relation_id as id, schedule_id from schedules_relations
                  where relation_type in ('${class_roster_teacher}', '${participant_teacher}')
                    and relation_id in (?)   -- teacher_id params
                 ) as teacher
                     join
                 (select id, end_at-start_at as duration, class_type from schedules
                  where class_type in ('${OnlineClass}', '${OfflineClass}') and delete_at = 0
                    and end_at >= ? and end_at < ?  -- params
                    and exists (
                          select schedule_id
                          from schedules_relations
                          where schedules.id = schedule_id
                            and relation_type = '${class_roster_class}'
                            and relation_id in (?) -- class_id params
                      )
                 ) as s
                 on teacher.schedule_id = s.id
             ) as teacher_schedule
                 left join assessments on teacher_schedule.schedule_id=assessments.schedule_id
         ) as tsa
             left join assessments_attendances on tsa.teacher_id=assessments_attendances.attendance_id and tsa.assessment_id=assessments_attendances.assessment_id
) tl
group by teacher_id;
`
	sql = strings.Replace(sql, "${OnlineClass}", entity.ScheduleClassTypeOnlineClass.String(), -1)
	sql = strings.Replace(sql, "${OfflineClass}", entity.ScheduleClassTypeOfflineClass.String(), -1)
	sql = strings.Replace(sql, "${class_roster_teacher}", entity.ScheduleRelationTypeClassRosterTeacher.String(), -1)
	sql = strings.Replace(sql, "${participant_teacher}", entity.ScheduleRelationTypeParticipantTeacher.String(), -1)
	sql = strings.Replace(sql, "${class_roster_class}", entity.ScheduleRelationTypeClassRosterClass.String(), -1)

	start, end, err := args.Duration.Value(ctx)
	if err != nil {
		log.Error(ctx, "ListTeacherLoadLessons: time range value failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("args", args))
		return nil, err
	}
	var result []*entity.TeacherLoadLesson
	err = r.QueryRawSQLTx(ctx, tx, &result, sql, args.TeacherIDs, start, end, args.ClassIDs)
	if err != nil {
		log.Error(ctx, "ListTeacherLoadLessons: sql exe failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("args", args))
		return nil, err
	}
	return result, nil
}

func (r *ReportDA) SummaryTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummaryFields, error) {
	sql := `
select
    count(if(is_attended=1 and class_type='${OnlineClass}', 1, null)) as live_completed_count,
    sum(if(is_attended=1 and class_type='${OnlineClass}', duration, 0)) as live_completed_duration,
    count(if(is_attended=1 and class_type='${OfflineClass}', 1, null)) as in_class_completed_count,
    sum(if(is_attended=1 and class_type='${OfflineClass}', duration, 0)) as in_class_completed_duration,
    count(if(is_attended=0 and class_type='${OnlineClass}', 1, null)) as live_missed_count,
    sum(if(is_attended=0 and class_type='${OnlineClass}', duration, 0)) as live_missed_duration,
    count(if(is_attended=0 and class_type='${OfflineClass}', 1, null)) as in_class_missed_count,
    sum(if(is_attended=0 and class_type='${OfflineClass}', duration, 0)) as in_class_missed_duration
from (
     select distinct tsa.teacher_id, tsa.schedule_id, tsa.duration, tsa.class_type, if(assessments_attendances.attendance_id is null, 0, 1) as is_attended from
         (select teacher_schedule.teacher_id as teacher_id, teacher_schedule.schedule_id, teacher_schedule.duration, teacher_schedule.class_type, assessments.id as assessment_id from
             (select teacher.id as teacher_id, s.id as schedule_id, s.duration, s.class_type from
                 (select relation_id as id, schedule_id from schedules_relations
                  where relation_type in ('${class_roster_teacher}', '${participant_teacher}')
                    and relation_id in (?)   -- teacher_id params
                 ) as teacher
                     join
                 (select id, end_at-start_at as duration, class_type from schedules
                  where class_type in ('${OnlineClass}', '${OfflineClass}') and delete_at = 0
                    and end_at >= ? and end_at < ?  -- params
                    and exists (
                          select schedule_id
                          from schedules_relations
                          where schedules.id = schedule_id
                            and relation_type = '${class_roster_class}'
                            and relation_id in (?) -- class_id params
                      )
                 ) as s
                 on teacher.schedule_id = s.id
             ) as teacher_schedule
                 left join assessments on teacher_schedule.schedule_id=assessments.schedule_id
         ) as tsa
             left join assessments_attendances on tsa.teacher_id=assessments_attendances.attendance_id and tsa.assessment_id=assessments_attendances.assessment_id
) tl;
`
	sql = strings.Replace(sql, "${OnlineClass}", entity.ScheduleClassTypeOnlineClass.String(), -1)
	sql = strings.Replace(sql, "${OfflineClass}", entity.ScheduleClassTypeOfflineClass.String(), -1)
	sql = strings.Replace(sql, "${class_roster_teacher}", entity.ScheduleRelationTypeClassRosterTeacher.String(), -1)
	sql = strings.Replace(sql, "${participant_teacher}", entity.ScheduleRelationTypeParticipantTeacher.String(), -1)
	sql = strings.Replace(sql, "${class_roster_class}", entity.ScheduleRelationTypeClassRosterClass.String(), -1)

	start, end, err := args.Duration.Value(ctx)
	if err != nil {
		log.Error(ctx, "SummaryTeacherLoadLessons: time range value failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("args", args))
		return nil, err
	}
	var result entity.TeacherLoadLessonSummaryFields
	err = r.QueryRawSQLTx(ctx, tx, &result, sql, args.TeacherIDs, start, end, args.ClassIDs)
	if err != nil {
		log.Error(ctx, "SummaryTeacherLoadLessons: sql exe failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("args", args))
		return nil, err
	}
	return &result, nil
}

func (r *ReportDA) MissedLessonsListInfo(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (model []*entity.TeacherLoadMissedLesson, err error) {
	sql := `
select sc.* from
    (
   	select 
		s.id,
		sl.relation_id as teacher_id,
		s.class_type, 
		s.title,
		(
            select relation_id from schedules_relations sel 
            where sel.schedule_id=s.id 
            and sel.relation_type='${class_roster_class}'
        ) as class_id,
        (
			select 
				count(*) 
			from schedules_relations sls 
			where  sls.schedule_id=s.id 
           	and sls.relation_type='${class_roster_student}'
		) as no_of_student,
		s.start_at as start_date,
		s.end_at as end_date
   	from schedules s
   	inner join schedules_relations sl
    on s.id=sl.schedule_id
   	where sl.relation_id=@teacherId -- teacherId params
    and s.class_type in ('${OnlineClass}', '${OfflineClass}') 
    and s.delete_at = 0
  	and s.end_at >= @startDt and s.end_at <@endDt
    and class_id in (@class_ids)
   	group by sl.schedule_id
   	order by s.end_at desc
	)
   sc left join  assessments ass on sc.id=ass.schedule_id
   where !exists
   ( 
	  select attendance_id from assessments_attendances ast 
     where ast.assessment_id = ass.id and ast.attendance_id=sc.teacher_id
    )
	LIMIT @pageSize OFFSET @offsetNumber`
	if request.Page < 0 {
		request.Page = constant.DefaultPageIndex
	}
	if request.PageSize < 0 {
		request.PageSize = constant.DefaultPageSize
	}
	sql = strings.Replace(sql, "${OnlineClass}", entity.ScheduleClassTypeOnlineClass.String(), -1)
	sql = strings.Replace(sql, "${OfflineClass}", entity.ScheduleClassTypeOfflineClass.String(), -1)
	sql = strings.Replace(sql, "${class_roster_student}", entity.ScheduleRelationTypeClassRosterStudent.String(), -1)
	sql = strings.Replace(sql, "${class_roster_class}", entity.ScheduleRelationTypeClassRosterClass.String(), -1)
	startAt, endAt, err := request.Duration.Value(ctx)
	if err != nil {
		return
	}
	params := map[string]interface{}{"teacherId": request.TeacherId, "startDt": startAt, "endDt": endAt,
		"class_ids": strings.Join(request.ClassIDs, ","),
		"pageSize":  request.PageSize, "offsetNumber": (request.Page - 1) * request.PageSize}
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
   	select 
		s.id,
		sl.relation_id as teacher_id
   	from schedules s
   	inner join schedules_relations sl
    on s.id=sl.schedule_id
   	where sl.relation_id=@teacherId -- teacherId params
    and s.class_type in ('${OnlineClass}', '${OfflineClass}') 
    and s.delete_at = 0
   	and s.end_at >= @startDt and s.end_at <@endDt
    and class_id in (@class_ids)
   	group by sl.schedule_id
   	order by s.end_at desc
	)
   sc left join  assessments ass on sc.id=ass.schedule_id
   where !exists
   ( 
	  select attendance_id from assessments_attendances ast 
     where ast.assessment_id = ass.id and ast.attendance_id=sc.teacher_id
    )`
	sql = strings.Replace(sql, "${OnlineClass}", entity.ScheduleClassTypeOnlineClass.String(), -1)
	sql = strings.Replace(sql, "${OfflineClass}", entity.ScheduleClassTypeOfflineClass.String(), -1)
	startAt, endAt, err := request.Duration.Value(ctx)
	if err != nil {
		return
	}
	params := map[string]interface{}{"teacherId": request.TeacherId, "startDt": startAt, "endDt": endAt,
		"class_ids": strings.Join(request.ClassIDs, ","),
		"pageSize":  request.Page, "limitNumber": (request.PageSize - 1) * request.Page}
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
