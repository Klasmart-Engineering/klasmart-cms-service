package da

import (
	"context"
	"fmt"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IStudentProgressLearnOutcomeAchievement interface {
	GetStudentProgressLearnOutcomeCountByStudentAndSubject(ctx context.Context, req *entity.LearnOutcomeAchievementRequest) (res []*entity.StudentProgressLearnOutcomeCount, err error)
}

func (r *ReportDA) GetStudentProgressLearnOutcomeCountByStudentAndSubject(ctx context.Context, req *entity.LearnOutcomeAchievementRequest) (res []*entity.StudentProgressLearnOutcomeCount, err error) {
	res = []*entity.StudentProgressLearnOutcomeCount{}

	plDuration := ``
	for _, duration := range req.Durations {
		min, max, _ := duration.Value(ctx)
		plDuration = fmt.Sprintf(`
when a.complete_time > %d and a.complete_time < %d THEN '%s' 
`, min, max, duration) + plDuration
	}
	plDuration = fmt.Sprintf(`case %s end as duration`, plDuration)

	sql := fmt.Sprintf(`
select 
	t.student_id,
	t.subject_id,
	count(1) as completed_count,
	sum(is_student_achieved) as achieved_count,
	sum(is_first_achieved) as first_achieved_count,
	t.duration	
from (
	select
	vss.schedule_id ,
	vss.subject_id ,
	vss.class_id,
	vaos.assessment_id ,
	vaos.outcome_id ,
	vaos.student_id ,
	vaos.is_student_achieved ,
	a.complete_time ,	
	a.status,
	if(vsofa.first_achieve_time=a.complete_time,1,0) as is_first_achieved,
	%s
from
	v_assessments_outcomes_students vaos
	left join assessments a on a.id=vaos.assessment_id 
	left join v_students_outcomes_first_achieve vsofa 
		on vsofa.student_id = vaos.student_id 
		and vsofa.outcome_id = vaos.outcome_id 
	left join v_schedules_subjects vss on vss.schedule_id = a.schedule_id 
	where a.delete_at=0
) t 

where 
 t.status=?
 and t.class_id=?

group by t.student_id,t.subject_id,t.duration
`, plDuration)
	args := []interface{}{entity.AssessmentStatusComplete, req.ClassID}
	err = r.QueryRawSQL(ctx, &res, sql, args...)
	if err != nil {
		return
	}
	return
}
