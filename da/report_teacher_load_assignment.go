package da

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ITeacherLoadAssessment interface {
	GetTeacherLoadAssignmentScheduledCount(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error)
	GetTeacherLoadAssignmentCompletedCountOfHomeFun(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error)
	GetTeacherLoadAssignmentScheduleIDListForStudyComment(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (scheduleIDs []string, err error)
}

func (r *ReportDA) GetTeacherLoadAssignmentPendingAssessment(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (scheduleIDs []string, err error) {

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
	s.class_type ,
	s.is_home_fun ,
	s.created_at ,
	a2.complete_time 
from
	schedules s
inner join schedules_relations sr on
	s.id = sr.schedule_id
	and sr.relation_type = 'class_roster_teacher'
inner join assessments a2 on
	s.id = a2.schedule_id  and  a2.type='study_h5p' and a2.status='complete'
) t
 
where t.teacher_id in (%s) 
and t.complete_time between ? and ?
`, r.getPlaceHolder(len(req.TeacherIDList)))
	var args []interface{}
	for _, teacherID := range req.TeacherIDList {
		args = append(args, teacherID)
	}
	startAt, endAt, err := req.Duration.Value(ctx)
	if err != nil {
		return
	}
	args = append(args, startAt, endAt)

	dbo.MustGetDB(ctx).Raw(sql, args...)
	panic("111")
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
select ? as teacher_id,count(1) as count_complete,count(IF(LENGTH(assess_comment)=0,NULL,1)) as count_comment 
from  home_fun_studies 
where JSON_contains(teacher_ids,?) 
and complete_at between ? and ?
`)
		args = append(args, teacherID, `"`+teacherID+`""`, startAt, endAt)
	}
	sql := strings.Join(sqlArr, "\n union \n")
	dbo.MustGetDB(ctx).Raw(sql, args...)

	panic("implement me")
}

func (r *ReportDA) GetTeacherLoadAssignmentScheduledCount(ctx context.Context, req *entity.TeacherLoadAssignmentRequest) (res entity.TeacherLoadAssignmentResponseItemSlice, err error) {
	res = entity.TeacherLoadAssignmentResponseItemSlice{}
	sql := fmt.Sprintf(`
select t.teacher_id,count(1) as count_of_scheduled_assignment from (
	select
	s.id as schedule_id,
	sr.relation_id as teacher_id,
	s.class_type ,
	s.is_home_fun ,
	s.created_at 
from
	schedules s
inner join schedules_relations sr on
	s.id = sr.schedule_id
	and sr.relation_type = 'class_roster_teacher'
	  
) t 

where t.class_type='Homework' 
and t.is_home_fun in(%s)
and t.created_at between ? and ?
and t.teacher_id in(%s)
group by t.teacher_id
`, r.getPlaceHolder(len(req.ClassTypeList)), r.getPlaceHolder(len(req.TeacherIDList)))
	var args []interface{}
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
	dbo.MustGetDB(ctx).Raw(sql, args...)

	panic("implement me")
}
