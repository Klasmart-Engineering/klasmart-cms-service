package da

import (
	"context"
	"fmt"
	"strings"

	"github.com/KL-Engineering/common-log/log"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type ITeacherLoadAssessment interface {
	GetTeacherLoadAssignmentScheduledCount(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error)
	GetTeacherLoadAssignmentCompletedCountOfHomeFun(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error)
	GetTeacherLoadAssignmentFeedbackOfHomeFun(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentFeedbackSlice, err error)
	GetTeacherLoadAssignmentScheduleIDListForStudyComment(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (scheduleIDs []string, err error)
	GetTeacherLoadAssignmentPendingAssessment(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error)
}

func (r *ReportDA) GetTeacherLoadAssignmentPendingAssessment(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error) {
	res = entity.TeacherLoadAssignmentResponseItemSlice{}
	plHomeFun := ""
	var args []interface{}
	for _, teacherID := range req.TeacherIDList {
		plHomeFun += `
select 
	hfs.schedule_id ,
	? as teacher_id,
	sr.relation_id as class_id,
	1 as is_home_fun ,
	student_id,
	UNIX_TIMESTAMP() - IF(hfs.due_at <= 0,hfs.create_at,hfs.due_at) - 7*24*60*60 as pending_seconds	
from  home_fun_studies hfs 
inner join schedules_relations sr on
		hfs.schedule_id = sr.schedule_id
where hfs.status=?
and sr.relation_type = ?
and sr.relation_id in (?)
and JSON_contains(teacher_ids,?) 
union all
`
		args = append(
			args,
			teacherID,
			entity.AssessmentStatusInProgress,
			entity.ScheduleRelationTypeClassRosterClass,
			req.ClassIDList,
			fmt.Sprintf(`"%s"`, teacherID),
		)
	}
	sql := fmt.Sprintf(`
select 	 
	t.teacher_id,
	count(1) as count_of_pending_assignment,
	avg(if(t.pending_seconds<0,0,t.pending_seconds))/(24*60*60) as  avg_days_of_pending_assignment
from 
(
	%s

	select
		s.id as schedule_id,
		sr.relation_id as teacher_id,
		sr2.relation_id as class_id,
		0 as is_home_fun,
		sr3.relation_id as student_id,
		UNIX_TIMESTAMP() - IF(s.due_at <= 0,s.created_at,s.due_at) - 7*24*60*60  as pending_seconds				 			 
	from
		schedules s
	inner join schedules_relations sr on
		s.id = sr.schedule_id	
	inner join schedules_relations sr2 on
		s.id = sr2.schedule_id	
	inner join schedules_relations sr3 on
		s.id = sr3.schedule_id	
	inner join assessments a2 on
		s.id = a2.schedule_id 
	where s.is_home_fun = 0 
	and s.class_type=?
	and sr.relation_type =  ?
	and sr.relation_id in (?)
	and sr2.relation_type = ?
	and sr2.relation_id in (?)
	and sr3.relation_type = ?
	and a2.status=?
) t

where t.teacher_id in (?)
and t.class_id in (?)
and t.is_home_fun in (?)
group by t.teacher_id
	`,
		plHomeFun)
	args = append(
		args,
		entity.ScheduleClassTypeHomework,
		entity.ScheduleRelationTypeClassRosterTeacher,
		req.TeacherIDList,
		entity.ScheduleRelationTypeClassRosterClass,
		req.ClassIDList,
		entity.ScheduleRelationTypeClassRosterStudent,
		entity.AssessmentStatusInProgress,
		req.TeacherIDList,
		req.ClassIDList,
	)

	var classTypeValues []int
	for _, classType := range req.ClassTypeList {
		switch classType {
		case constant.ReportClassTypeStudy:
			classTypeValues = append(classTypeValues, 0)
		case constant.ReportClassTypeHomeFun:
			classTypeValues = append(classTypeValues, 1)
		default:
			log.Error(ctx, "invalid class_type", log.Any("request", req))
			err = constant.ErrInvalidArgs
			return
		}
	}
	args = append(args, classTypeValues)

	err = r.QueryRawSQL(ctx, &res, sql, args...)
	if err != nil {
		return
	}
	return
}
func (r *ReportDA) GetTeacherLoadAssignmentScheduleIDListForStudyComment(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (scheduleIDs []string, err error) {
	scheduleIDs = []string{}
	sql := fmt.Sprintf(`
select t.schedule_id from 
(
	select
		s.id as schedule_id,
		sr.relation_id as teacher_id,
		sr2.relation_id as class_id,
		s.class_type ,
		s.is_home_fun ,
		s.created_at ,
		a2.complete_time 
	from
		schedules s
	inner join schedules_relations sr on
		s.id = sr.schedule_id	
	inner join schedules_relations sr2 on
		s.id = sr2.schedule_id		 
	inner join assessments a2 on
		s.id = a2.schedule_id     
	where sr.relation_type = ?
	and sr2.relation_type = ?	
	and a2.status=?
	and s.class_type=?
	and s.is_home_fun=?
) t
 
where t.teacher_id in (%s) 
and t.class_id in(%s)
and t.complete_time between ? and ?
`,
		r.getPlaceHolder(len(req.TeacherIDList)),
		r.getPlaceHolder(len(req.ClassIDList)),
	)
	var args []interface{}
	args = append(
		args,
		entity.ScheduleRelationTypeClassRosterTeacher,
		entity.ScheduleRelationTypeClassRosterClass,
		entity.AssessmentStatusComplete,
		entity.ScheduleClassTypeHomework,
		0,
	)
	for _, teacherID := range req.TeacherIDList {
		args = append(args, teacherID)
	}
	for _, classID := range req.ClassIDList {
		args = append(args, classID)
	}
	startAt, endAt, err := req.Duration.Value(ctx)
	if err != nil {
		return
	}
	args = append(args, startAt, endAt)

	err = r.QueryRawSQL(ctx, &scheduleIDs, sql, args...)
	if err != nil {
		return
	}
	return
}
func (r *ReportDA) GetTeacherLoadAssignmentCompletedCountOfHomeFun(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error) {
	res = entity.TeacherLoadAssignmentResponseItemSlice{}
	var sqlArr []string
	var args []interface{}
	startAt, endAt, err := req.Duration.Value(ctx)
	if err != nil {
		return
	}
	for _, teacherID := range req.TeacherIDList {
		sqlArr = append(sqlArr, `
select 
	? as teacher_id,
	count(1) as count_of_completed_assignment,
	count(IF(LENGTH(assess_comment)=0,NULL,1)) as count_of_commented_assignment 
from  home_fun_studies 

where JSON_contains(teacher_ids,?) 
and complete_at between ? and ?
`)
		args = append(args, teacherID, fmt.Sprintf(`"%s"`, teacherID), startAt, endAt)
	}
	sql := strings.Join(sqlArr, "\n union \n")
	err = r.QueryRawSQL(ctx, &res, sql, args...)
	if err != nil {
		return
	}
	return
}

func (r *ReportDA) GetTeacherLoadAssignmentFeedbackOfHomeFun(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentFeedbackSlice, err error) {
	res = entity.TeacherLoadAssignmentFeedbackSlice{}
	var sqlArr []string
	var args []interface{}
	startAt, endAt, err := req.Duration.Value(ctx)
	if err != nil {
		return
	}
	for _, teacherID := range req.TeacherIDList {
		sqlArr = append(sqlArr, `
select 
	? as teacher_id,	
	schedule_id,
	count(IF(LENGTH(assess_comment)=0,NULL,1))/count(1) as feedback_of_assignment 
from  home_fun_studies 

where schedule_id in (
    select 	 	
        schedule_id  
    from  home_fun_studies 
    where JSON_contains(teacher_ids,?) 
    and complete_at between ? and  ?
)
group by schedule_id
`)
		args = append(args, teacherID, fmt.Sprintf(`"%s"`, teacherID), startAt, endAt)
	}
	sql := strings.Join(sqlArr, "\n union \n")
	err = r.QueryRawSQL(ctx, &res, sql, args...)
	if err != nil {
		return
	}
	return
}

func (r *ReportDA) GetTeacherLoadAssignmentScheduledCount(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error) {
	res = entity.TeacherLoadAssignmentResponseItemSlice{}
	sql := fmt.Sprintf(`
select 
	t.teacher_id,
	count(1) as count_of_scheduled_assignment 
from 
(
	select
		s.id as schedule_id,
		sr.relation_id as teacher_id,
		s.class_type ,
		s.is_home_fun ,
		s.created_at,
		sr2.relation_id  as class_id  
	from
		schedules s
	inner join schedules_relations sr on
		s.id = sr.schedule_id
	inner join schedules_relations sr2 on
		s.id = sr2.schedule_id
	where sr.relation_type =  ?	
	and sr2.relation_type = ?	 
) t 

where t.class_type= ? 
and t.is_home_fun in(%s)
and t.created_at between ? and ?
and t.teacher_id in(%s)
and t.class_id in(%s)
group by t.teacher_id
`,
		r.getPlaceHolder(len(req.ClassTypeList)),
		r.getPlaceHolder(len(req.TeacherIDList)),
		r.getPlaceHolder(len(req.ClassIDList)),
	)
	var args []interface{}
	args = append(
		args,
		entity.ScheduleRelationTypeClassRosterTeacher,
		entity.ScheduleRelationTypeClassRosterClass,
		entity.ScheduleClassTypeHomework,
	)
	for _, classType := range req.ClassTypeList {
		switch classType {
		case constant.ReportClassTypeStudy:
			args = append(args, 0)
		case constant.ReportClassTypeHomeFun:
			args = append(args, 1)
		default:
			log.Error(ctx, "invalid class_type", log.Any("request", req))
			err = constant.ErrInvalidArgs
			return
		}
	}
	startAt, endAt, err := req.Duration.Value(ctx)
	if err != nil {
		return
	}
	args = append(args, startAt, endAt)
	for _, teacherID := range req.TeacherIDList {
		args = append(args, teacherID)
	}
	for _, classID := range req.ClassIDList {
		args = append(args, classID)
	}
	err = r.QueryRawSQL(ctx, &res, sql, args...)
	if err != nil {
		return
	}
	return
}
