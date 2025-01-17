package da

import (
	"context"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
)

type ITeacherLoadLesson interface {
	ListTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error)
	SummaryTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummaryFields, error)
	MissedLessonsListInfo(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (model []*entity.TeacherLoadMissedLesson, err error)
	MissedLessonsListTotal(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (total int, err error)
	GetTeacherLoadItems(ctx context.Context, op *entity.Operator, tr entity.TimeRange, teacherIDs []string, classIDs []string) (res []*entity.TeacherLoadItem, err error)
}

func (r *ReportDA) ListTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error) {
	sql := `
select
       teacher_id,
       count(if(is_attended=1 and class_type=?, 1, null)) as live_completed_count,
       count(if(is_attended=1 and class_type=?, 1, null)) as in_class_completed_count,
       count(if(is_attended=0 and class_type=?, 1, null)) as live_missed_count,
       count(if(is_attended=0 and class_type=?, 1, null)) as in_class_missed_count,
       count(1) as total_schedule
from (
     	select DISTINCT 
			auv.user_id as teacher_id,
			av.schedule_id ,
			s.end_at -s.start_at as duratopn,
			s.class_type ,
			if(auv.status_by_system != ?,1,0) as is_attended 
		from assessments_users_v2 auv 
		inner join assessments_v2 av on av.id = auv.assessment_id 
		inner join schedules s on s.id = av.schedule_id 
		inner join schedules_relations sr 
			on sr.schedule_id = s.id 
			and sr.relation_type = ?  
			and sr.relation_id in (?)
		where auv.user_type= ?
			and auv.user_id in (?)
			and s.class_type in (?,?)
			and s.end_at >=? and s.end_at <?
) tl
group by teacher_id;
`
	start, end, err := args.Duration.Value(ctx)
	if err != nil {
		log.Error(ctx, "ListTeacherLoadLessons: time range value failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("args", args))
		return nil, err
	}
	sqlArgs := []interface{}{
		entity.ScheduleClassTypeOnlineClass,
		entity.ScheduleClassTypeOfflineClass,
		entity.ScheduleClassTypeOnlineClass,
		entity.ScheduleClassTypeOfflineClass,
		v2.AssessmentUserSystemStatusNotStarted,
		entity.ScheduleRelationTypeClassRosterClass,
		args.ClassIDs,
		v2.AssessmentUserTypeTeacher,
		args.TeacherIDs,
		entity.ScheduleClassTypeOnlineClass,
		entity.ScheduleClassTypeOfflineClass,
		start,
		end,
	}

	result := []*entity.TeacherLoadLesson{}
	err = r.QueryRawSQLTx(ctx, tx, &result, sql, sqlArgs...)
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
    count(if(is_attended=1 and class_type=?, 1, null)) as live_completed_count,
    sum(if(is_attended=1 and class_type=?, duration, 0)) as live_completed_duration,
    count(if(is_attended=1 and class_type=?, 1, null)) as in_class_completed_count,
    sum(if(is_attended=1 and class_type=?, duration, 0)) as in_class_completed_duration,
    count(if(is_attended=0 and class_type=?, 1, null)) as live_missed_count,
    sum(if(is_attended=0 and class_type=?, duration, 0)) as live_missed_duration,
    count(if(is_attended=0 and class_type=?, 1, null)) as in_class_missed_count,
    sum(if(is_attended=0 and class_type=?, duration, 0)) as in_class_missed_duration
from (
     select DISTINCT 
			auv.user_id as teacher_id,
			av.schedule_id ,
			s.end_at -s.start_at as duration,
			s.class_type ,
			if(auv.status_by_system != ?,1,0) as is_attended 
	from assessments_users_v2 auv 
	inner join assessments_v2 av on av.id = auv.assessment_id 
	inner join schedules s on s.id = av.schedule_id 
	inner join schedules_relations sr 
		on sr.schedule_id = s.id 
		and sr.relation_type = ?  
		and sr.relation_id in (?)
	where auv.user_type= ?
		and auv.user_id in (?)
		and s.class_type in (?,?)
		and s.end_at >=? and s.end_at <?
) tl;
`
	start, end, err := args.Duration.Value(ctx)
	if err != nil {
		log.Error(ctx, "SummaryTeacherLoadLessons: time range value failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("args", args))
		return nil, err
	}
	sqlArgs := []interface{}{
		entity.ScheduleClassTypeOnlineClass,
		entity.ScheduleClassTypeOnlineClass,
		entity.ScheduleClassTypeOfflineClass,
		entity.ScheduleClassTypeOfflineClass,
		entity.ScheduleClassTypeOnlineClass,
		entity.ScheduleClassTypeOnlineClass,
		entity.ScheduleClassTypeOfflineClass,
		entity.ScheduleClassTypeOfflineClass,
		v2.AssessmentUserSystemStatusNotStarted,
		entity.ScheduleRelationTypeClassRosterClass,
		args.ClassIDs,
		v2.AssessmentUserTypeTeacher,
		args.TeacherIDs,
		entity.ScheduleClassTypeOnlineClass,
		entity.ScheduleClassTypeOfflineClass,
		start,
		end,
	}

	var result entity.TeacherLoadLessonSummaryFields
	err = r.QueryRawSQLTx(ctx, tx, &result, sql, sqlArgs...)
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
	select sr.* from 
	(
		select 
		s.id,
		sl.relation_id as teacher_id ,
		s.class_type,s.end_at,
		s.title,
		(
			select relation_id from schedules_relations sel
            where sel.schedule_id=s.id
            and sel.relation_type=?
        ) as class_id,
		(
			select count(*)
			from schedules_relations sls
			where  sls.schedule_id=s.id
           	and sls.relation_type=?
		) as no_of_student,
		s.start_at as start_date,
		s.end_at as end_date
		from 
			(
				select * from schedules 
				where class_type in (?, ?)
				and delete_at = 0
				and end_at >= ? and end_at <?
				and class_id in  (?) 
			) s 
		inner join schedules_relations sl 
		on s.id=sl.schedule_id
		where sl.relation_id =?
		and sl.relation_type in (?, ?)
	) sr
	left join assessments ass 
	on sr.id=ass.schedule_id
	where not exists
	( 
		select attendance_id from assessments_attendances ast 
		where ast.assessment_id = ass.id and ast.attendance_id=sr.teacher_id
	)
	order by sr.id,sr.end_at desc
	LIMIT ? OFFSET ?`
	startAt, endAt, err := request.Duration.Value(ctx)
	if err != nil {
		return
	}
	err = r.QueryRawSQL(ctx, &model, sql,
		entity.ScheduleRelationTypeClassRosterClass.String(),
		entity.ScheduleRelationTypeClassRosterStudent.String(),
		entity.ScheduleClassTypeOnlineClass.String(),
		entity.ScheduleClassTypeOfflineClass.String(),
		startAt,
		endAt,
		request.ClassIDs,
		request.TeacherId,
		entity.ScheduleRelationTypeClassRosterTeacher.String(),
		entity.ScheduleRelationTypeParticipantTeacher.String(),
		request.PageSize,
		(request.Page-1)*request.PageSize)
	if err != nil {
		log.Error(ctx, "exec missedLessonsListInfo sql failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("params", request))
		return
	}
	return
}
func (r *ReportDA) MissedLessonsListTotal(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (total int, err error) {
	sql := `
	select count(*) from 
	(
		select s.id,sl.relation_id,s.end_at from 
			(
				select * from schedules 
				where class_type in (?, ?) 
				and delete_at = 0
				and end_at >= ? and end_at <?
				and class_id in (?) 
			) s 
		inner join schedules_relations sl 
		on s.id=sl.schedule_id
		where sl.relation_id =?
		and sl.relation_type in (?, ?)
	) sr
	left join assessments ass 
	on sr.id=ass.schedule_id
	where not exists
	( 
		select attendance_id from assessments_attendances ast 
		where ast.assessment_id = ass.id and ast.attendance_id=sr.relation_id
	)`
	startAt, endAt, err := request.Duration.Value(ctx)
	if err != nil {
		return
	}
	err = r.QueryRawSQL(ctx, &total, sql,
		entity.ScheduleClassTypeOnlineClass.String(),
		entity.ScheduleClassTypeOfflineClass.String(),
		startAt,
		endAt,
		request.ClassIDs,
		request.TeacherId,
		entity.ScheduleRelationTypeClassRosterTeacher.String(),
		entity.ScheduleRelationTypeParticipantTeacher.String())
	if err != nil {
		log.Error(ctx, "exec missedLessonsListTotal sql failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("params", request))
		return
	}
	return
}

func (r *ReportDA) GetTeacherLoadItems(ctx context.Context, op *entity.Operator, tr entity.TimeRange, teacherIDs []string, classIDs []string) ([]*entity.TeacherLoadItem, error) {
	if !config.Get().RedisConfig.OpenCache {
		return r.getTeacherLoadItemsFromMySQL(ctx, op, tr, teacherIDs, classIDs)
	}

	request := &getTeacherUsageOverviewQueryCondition{
		Operator:   op,
		TimeRange:  tr,
		TeacherIDs: teacherIDs,
		ClassIDs:   classIDs,
	}

	response := []*entity.TeacherLoadItem{}
	err := r.teacherUsageOverviewCache.Get(ctx, request, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

type getTeacherUsageOverviewQueryCondition struct {
	Operator   *entity.Operator `json:"operator"`
	TimeRange  entity.TimeRange `json:"time_range"`
	TeacherIDs []string         `json:"teacher_ids"`
	ClassIDs   []string         `json:"class_ids"`
}

func (r *ReportDA) getTeacherUsageOverview(ctx context.Context, condition interface{}) (interface{}, error) {
	request, ok := condition.(*getTeacherUsageOverviewQueryCondition)
	if !ok {
		log.Error(ctx, "invalid request", log.Any("condition", condition))
		return nil, constant.ErrInvalidArgs
	}

	return r.getTeacherLoadItemsFromMySQL(ctx, request.Operator, request.TimeRange, request.TeacherIDs, request.ClassIDs)
}

func (r *ReportDA) getTeacherLoadItemsFromMySQL(ctx context.Context, op *entity.Operator, tr entity.TimeRange, teacherIDs []string, classIDs []string) (res []*entity.TeacherLoadItem, err error) {
	start, end, err := tr.Value(ctx)
	if err != nil {
		return
	}
	sql := `
select 
	auv.user_id as teacher_id,
	count(1) as total_lessons, 
	sum(if(auv.status_by_system=?,1,0)) as missed_lessons
from assessments_users_v2 auv  
inner join assessments_v2 av on av.id =auv.assessment_id  
inner join schedules s on s.id =av.schedule_id 
inner join schedules_relations sr on sr.schedule_id = s.id  and sr.relation_type = ? and sr.relation_id in (?)
where auv.user_type = ? 
and auv.user_id in (?)
and s.class_type in (?,?) 
and s.end_at >= ? and s.end_at < ?
and av.org_id = ?
group by auv.user_id 
`
	args := []interface{}{
		v2.AssessmentUserSystemStatusNotStarted,
		entity.ScheduleRelationTypeClassRosterClass,
		classIDs,
		v2.AssessmentUserTypeTeacher,
		teacherIDs,
		entity.ScheduleClassTypeOnlineClass,
		entity.ScheduleClassTypeOfflineClass,
		start,
		end,
		op.OrgID,
	}
	res = []*entity.TeacherLoadItem{}
	err = r.QueryRawSQLTx(ctx, dbo.MustGetDB(ctx), &res, sql, args...)
	if err != nil {
		return
	}
	return
}
