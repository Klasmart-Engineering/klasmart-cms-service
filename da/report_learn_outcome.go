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
	sbAssessmentStudent := NewSqlBuilder(ctx, `select a.id as assessment_id,sr.relation_id as student_id,a.schedule_id from assessments a 
	left join schedules_relations sr on  sr.relation_type ='class_roster_student' and sr.schedule_id =a.schedule_id 
	where a.complete_time >= ? and a.complete_time < ?
	union 
	select hfs.id as assessment_id,hfs.student_id,hfs.schedule_id  from home_fun_studies hfs 
	where hfs.complete_at >= ? and hfs.complete_at < ? `, from, to, from, to)

	sbScheduleID := NewSqlBuilder(ctx, `select sr.schedule_id from schedules_relations sr
where sr.relation_type ='class_roster_teacher'
and sr.relation_id in(?)`, teacherIDs)
	sql := `
select t.student_id ,count(1) as total_achieved_outcome_count,sum(t.student_achieved) as achieved_outcome_count from (
	select ao.outcome_id , sa.student_id,IF(oa.id is null,0,1) as student_achieved
	from assessments_outcomes ao
	left join (
		{{.sbAssessmentStudent}}
	) sa  on sa.assessment_id=ao.assessment_id 
	left join outcomes_attendances oa on ao.assessment_id =oa.assessment_id and oa.assessment_id =ao.assessment_id 
	where ao.skip =0 and oa.id is NULL 
	and sa.assessment_id is not  null 
	and sa.schedule_id in(
		{{.sbScheduleID}}
	)
) t 
group by t.student_id 
 `
	sb := NewSqlBuilder(ctx, sql).
		Replace(ctx, "sbAssessmentStudent", sbAssessmentStudent).
		Replace(ctx, "sbScheduleID", sbScheduleID)
	sql, args, err := sb.Build(ctx)
	if err != nil {
		return
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
select count(distinct sr.relation_id ) as cnt from schedules_relations sr  
inner join schedules_relations sr2 
	on sr2.schedule_id =sr.schedule_id 
	and sr2.relation_type ='class_roster_teacher'
	and sr2.relation_id in(?)
where sr.relation_type ='learning_outcome' and sr.schedule_id in
(
	select schedule_id from home_fun_studies hfs where hfs.complete_at between ? and  ? 
	union all 
	select schedule_id  from assessments a where a.complete_time   between ? and  ? 
)
`
	args := []interface{}{
		teacherIDs,
		from,
		to,
		from,
		to,
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
