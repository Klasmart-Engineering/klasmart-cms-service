package da

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ITeacherLoadAssessment interface {
	GetTeacherLoadAssignmentScheduledCount(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error)
	GetTeacherLoadAssignmentCompletedCountOfHomeFun(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error)
	GetTeacherLoadAssignmentScheduleIDListForStudyComment(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (scheduleIDs []string, err error)
	GetTeacherLoadAssignmentPendingAssessment(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error)
}

func (r *ReportDA) GetTeacherLoadAssignmentPendingAssessment(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error) {
	res = entity.TeacherLoadAssignmentResponseItemSlice{}
	plHomeFun := ""
	var args []interface{}
	for _, teacherID := range req.TeacherIDList {
		plHomeFun += `
union
select 
	hfs.schedule_id ,
	? as teacher_id,
	sr.relation_id as class_id,
	1 as is_home_fun ,
	UNIX_TIMESTAMP() - IF(hfs.due_at <= 0,hfs.create_at,hfs.due_at) - 7*24*60*60 as pending_seconds	
from home_fun_studies hfs 
inner join schedules_relations sr on
		hfs.schedule_id = sr.schedule_id
where hfs.status=?
and sr.relation_type = ?
and JSON_contains(teacher_ids,?) 
`
		args = append(
			args,
			teacherID,
			entity.AssessmentStatusInProgress,
			entity.ScheduleRelationTypeClassRosterClass,
			fmt.Sprintf(`"%s"`, teacherID),
		)
	}
	sql := fmt.Sprintf(`
select 	 
	t.teacher_id,
	count(distinct t.schedule_id) as count_of_pending_assignment,
	avg(if(t.pending_seconds<0,0,t.pending_seconds))/(24*60*60) as  avg_days_of_pending_assignment
from 
(
	select
		s.id as schedule_id,
		sr.relation_id as teacher_id,
		sr2.relation_id as class_id,
		0 as is_home_fun,
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
	and sr2.relation_type = ?
	and sr3.relation_type = ?
	and a2.type=?
	and a2.status=?

%s

) t

where t.teacher_id in (%s)
and t.class_id in (%s)
and t.is_home_fun in (%s)
group by t.teacher_id
	`,
		plHomeFun,
		r.getPlaceHolder(len(req.TeacherIDList)),
		r.getPlaceHolder(len(req.ClassIDList)),
		r.getPlaceHolder(len(req.ClassTypeList)),
	)
	args = append(
		args,
		entity.ScheduleClassTypeHomework,
		entity.ScheduleRelationTypeClassRosterTeacher,
		entity.ScheduleRelationTypeClassRosterClass,
		entity.ScheduleRelationTypeClassRosterStudent,
		entity.AssessmentTypeStudyH5p,
		entity.AssessmentStatusInProgress,
	)

	for _, teacherID := range req.TeacherIDList {
		args = append(args, teacherID)
	}
	for _, classID := range req.ClassIDList {
		args = append(args, classID)
	}
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
	and	a2.type=? 
	and a2.status=?
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
		entity.AssessmentTypeStudyH5p,
		entity.AssessmentStatusComplete,
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