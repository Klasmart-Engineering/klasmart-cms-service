package da

import (
	"context"
	"fmt"
	"strings"

	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IStudentProgressLearnOutcomeAchievement interface {
	GetStudentProgressLearnOutcomeCountByStudentAndSubject(ctx context.Context, req *entity.LearnOutcomeAchievementRequest) (res []*entity.StudentProgressLearnOutcomeCount, err error)
}

func (r *ReportDA) GetStudentProgressLearnOutcomeCountByStudentAndSubject(ctx context.Context, req *entity.LearnOutcomeAchievementRequest) (res []*entity.StudentProgressLearnOutcomeCount, err error) {
	res = []*entity.StudentProgressLearnOutcomeCount{}
	sbStudentAchieveOutcome := NewSqlBuilder(ctx, `
select  
	auv.assessment_id ,
	auov.outcome_id,	
	auv.user_id as student_id,
	srClass.relation_id as class_id,
	av.complete_at as complete_time,
	if(srSubject.relation_id is null,'',srSubject.relation_id) as subject_id,	
	IF(auov.status=?,1,0) as is_student_achieved
	
from (
	select 
		DISTINCT assessment_user_id ,
		outcome_id,
		status,
		delete_at 
	from assessments_users_outcomes_v2
) auov 
inner join assessments_users_v2 auv on auov.assessment_user_id =auv.id  and auv.user_type =?
inner join assessments_v2 av on auv.assessment_id =av.id 
left join schedules_relations srSubject on srSubject.schedule_id =av.schedule_id and srSubject.relation_type =?
inner join schedules_relations srClass on srClass.schedule_id =av.schedule_id and srClass.relation_type=?

where 
	auov.delete_at=0
	and auv.delete_at =0
	and av.delete_at =0

	and av.status = ?
	and auv.status_by_system= ?
 	and srClass.relation_id = ?	
	and EXISTS (
		select 1 from schedules_relations sr where sr.schedule_id =av.schedule_id and sr.relation_id =auv.user_id and sr.relation_type=?
	)`,
		v2.AssessmentUserOutcomeStatusAchieved,
		v2.AssessmentUserTypeStudent,
		entity.ScheduleRelationTypeSubject,
		entity.ScheduleRelationTypeClassRosterClass,
		v2.AssessmentStatusComplete,
		v2.AssessmentUserStatusParticipate,
		req.ClassID,
		entity.ScheduleRelationTypeClassRosterStudent,
	)
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
from (
	{{.sbStudentAchieveOutcome}}
) t0 
left join (
	select 
		t1.outcome_id ,
		t1.student_id,
		min(complete_time) as first_achieve_time  
	from (
			{{.sbStudentAchieveOutcome}}
	) t1
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
from (
	{{.sbStudentFirstAchieveOutcome}}
) t 
{{.sbCondition}}
group by t.student_id,t.subject_id,t.duration
`)

	sbCondition := NewSqlBuilder(ctx, `
where t.class_id=?
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
