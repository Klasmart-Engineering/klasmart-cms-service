package da

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IStudentProgressLearnOutcomeAchievement interface {
	GetStudentProgressLearnOutcomeCountByStudentAndSubject(ctx context.Context, req *entity.LearnOutcomeAchievementRequest) (res []*entity.StudentProgressLearnOutcomeCount, err error)
}

func (r *ReportDA) GetStudentProgressLearnOutcomeCountByStudentAndSubject(ctx context.Context, req *entity.LearnOutcomeAchievementRequest) (res []*entity.StudentProgressLearnOutcomeCount, err error) {
	res = []*entity.StudentProgressLearnOutcomeCount{}

	sbStudentAchieveOutcome := NewSqlBuilder(ctx, `
select
    t1.assessment_id,
    t1.outcome_id,
    t2.student_id,
    vss.class_id ,
    a.complete_time ,
    vss.subject_id ,    
    if(oa.id is null,0,1) as is_student_achieved
from
    (
        SELECT
            DISTINCT assessment_id,
                     outcome_id
        FROM
            assessments_outcomes
    ) t1
INNER JOIN (
        SELECT
            assessment_id,
            attendance_id AS student_id
        FROM
            assessments_attendances
        WHERE
                checked = 1
          AND origin = 'class_roaster'
          AND role = 'student'
          and attendance_id=?
    ) t2 
   ON	t1.assessment_id = t2.assessment_id
left join outcomes_attendances oa
	on t1.assessment_id = oa.assessment_id
    and t1.outcome_id = oa.outcome_id
    and t2.student_id = oa.attendance_id 
inner join assessments a on
            a.id = t1.assessment_id and a.status ='complete'  
left join v_schedules_subjects vss on vss.schedule_id =a.schedule_id 

union all 

select 
	ao.assessment_id ,
	ao.outcome_id ,
	hfs.student_id ,	
	vss.class_id ,
	hfs.complete_at as complete_time,
	vss.subject_id ,
	IF (EXISTS (select * from outcomes_attendances oa where oa.assessment_id =ao.assessment_id and oa.outcome_id=ao.outcome_id and oa.attendance_id=hfs.student_id),1,0) as is_student_achieved
from home_fun_studies hfs 
inner join assessments_outcomes ao on ao.assessment_id =hfs.id 
left join v_schedules_subjects vss on vss.schedule_id =hfs.schedule_id 
where hfs.student_id =?`, req.StudentID, req.StudentID)
	sbStudentFirstAchieveOutcome := NewSqlBuilder(ctx, `
select 
	t0.assessment_id,
	t0.outcome_id,
	t0.subject_id,
	t0.student_id,
	t0.class_id,
	t0.complete_time,
	t2.first_achieve_time,
	is_student_achieved,
	{{.sbPlDuration}}	
from ({{.sbStudentAchieveOutcome}}) t0 
left join (
	select 
		t1.outcome_id ,
		t1.student_id,
		min(complete_time) as first_achieve_time  
		from 
	({{.sbStudentAchieveOutcome}}) t1
	where t1.is_student_achieved =1 
	group by t1.outcome_id ,t1.student_id 
) t2 on t2.outcome_id = t0.outcome_id  and t2.student_id=t0.student_id
where ({{.sbCondDuration}})
`)
	plDuration := ``
	condDurations := []string{}
	for _, duration := range req.Durations {
		min, max, _ := duration.Value(ctx)
		plDuration = fmt.Sprintf(`
		when t0.complete_time >= %d and t0.complete_time < %d THEN '%s'`, min, max, duration) + plDuration
		condDurations = append(condDurations, fmt.Sprintf("(t0.complete_time >= %d and t0.complete_time < %d)", min, max))
	}

	sbPlDuration := NewSqlBuilder(ctx, fmt.Sprintf(`case %s end as duration`, plDuration))
	sbCondDuration := NewSqlBuilder(ctx, strings.Join(condDurations, " or "))
	sbStudentFirstAchieveOutcome.
		Replace(ctx, "sbStudentAchieveOutcome", sbStudentAchieveOutcome).
		Replace(ctx, "sbPlDuration", sbPlDuration).
		Replace(ctx, "sbCondDuration", sbCondDuration)

	sb := NewSqlBuilder(ctx, `
select 
	t.student_id,
	t.subject_id,
	t.duration,
	count(1) as completed_count,
	sum(t.is_student_achieved) as achieved_count,
	sum(if(t.complete_time=t.first_achieve_time,1,0)) as first_achieved_count
from ({{.sbStudentFirstAchieveOutcome}}) t 

{{.sbCondition}}
group by t.student_id,t.subject_id,t.duration
`)

	sbCondition := NewSqlBuilder(ctx, `
	where
 t.class_id=?
 and t.subject_id in (?)
`, req.ClassID, append(req.SelectedSubjectIDList, req.UnSelectedSubjectIDList...))
	sb.Replace(ctx, "sbStudentFirstAchieveOutcome", sbStudentFirstAchieveOutcome).
		Replace(ctx, "sbCondition", sbCondition)
	sql, args, err := sb.Build(ctx)
	if err != nil {
		return
	}
	err = r.QueryRawSQL(ctx, &res, sql, args...)
	if err != nil {
		return
	}
	return
}
