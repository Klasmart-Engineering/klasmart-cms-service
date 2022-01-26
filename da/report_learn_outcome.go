package da

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

	"gitlab.badanamu.com.cn/calmisland/dbo"
)

type ILearnOutcome interface {
	GetCompleteLearnOutcomeCount(ctx context.Context, tx *dbo.DBContext, from, to int64, teacherIDs []string) (cnt int, err error)
	GetStudentAchievedOutcome(ctx context.Context, tx *dbo.DBContext, from, to int64, teacherIDs []string) (studentOutcomeAchievedCounts []*entity.StudentOutcomeAchievedCount, err error)
}

func (r *ReportDA) GetStudentAchievedOutcome(ctx context.Context, tx *dbo.DBContext, from, to int64, teacherIDs []string) (studentOutcomeAchievedCounts []*entity.StudentOutcomeAchievedCount, err error) {
	studentOutcomeAchievedCounts = []*entity.StudentOutcomeAchievedCount{}
	if len(teacherIDs) == 0 {
		return
	}

	sql := `
select ass.student_id,count(1) as total_achieved_outcome_count,SUM(IF(oa.id is null,0,1)) as achieved_outcome_count from (
select a.id as assessment_id,sr.relation_id as student_id,a.schedule_id from assessments a
inner join schedules_relations sr on sr.relation_type ='class_roster_student' and sr.schedule_id =a.schedule_id
where a.complete_time >= ? and a.complete_time < ?
union
select hfs.id as assessment_id,hfs.student_id,hfs.schedule_id from home_fun_studies hfs
where hfs.complete_at >= ? and hfs.complete_at < ?

) ass 
inner join assessments_outcomes ao on ao.assessment_id = ass.assessment_id
left join outcomes_attendances  oa on oa.assessment_id = ass.assessment_id and oa.attendance_id =ass.student_id and oa.outcome_id  = ao.outcome_id 
where EXISTS (
select relation_id from schedules_relations sr where sr.relation_type ='class_roster_teacher' and sr.schedule_id =ass.schedule_id and sr.relation_id  in(?)
)
and ao.skip=0

group by ass.student_id
`
	args := []interface{}{
		from,
		to,
		from,
		to,
		teacherIDs,
	}
	err = r.QueryRawSQL(ctx, &studentOutcomeAchievedCounts, sql, args...)
	if err != nil {
		return
	}
	return
}
func (r *ReportDA) GetCompleteLearnOutcomeCount(ctx context.Context, tx *dbo.DBContext, from, to int64, teacherIDs []string) (cnt int, err error) {
	if len(teacherIDs) == 0 {
		return
	}
	sql := `
select count(distinct ao.outcome_id) as cnt from (
select hfs.id as assessment_id,hfs.schedule_id from home_fun_studies hfs where hfs.complete_at>= ? and hfs.complete_at<?
union all
select a.id  as assessment_id,a.schedule_id from assessments a where a.complete_time>= ? and a.complete_time<?
) sa 
inner join assessments_outcomes ao on ao.assessment_id = sa.assessment_id
where  EXISTS (
select sr.id from schedules_relations sr  
where sr.relation_type ='class_roster_teacher' 
and sr.schedule_id = sa.schedule_id 
and sr.relation_id in(?)
)

`
	args := []interface{}{
		from,
		to,
		from,
		to,
		teacherIDs,
	}
	res := struct {
		Cnt int `json:"cnt" gorm:"column:cnt" `
	}{}
	err = r.QueryRawSQL(ctx, &res, sql, args...)
	if err != nil {
		return
	}
	cnt = res.Cnt
	return
}
