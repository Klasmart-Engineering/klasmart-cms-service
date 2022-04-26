package da

import (
	"context"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

	"gitlab.badanamu.com.cn/calmisland/dbo"
)

func (r *ReportDA) GetLessonPlanFilter(ctx context.Context, op *entity.Operator, classID string) (items []*entity.ScheduleShortInfo, err error) {
	items = []*entity.ScheduleShortInfo{}
	sql := `
select 
	DISTINCT cc.id,
	cc.content_name as name
from schedules s 
inner join assessments_v2 av on s.id = av.schedule_id 
inner join cms_contents cc on cc.id = s.lesson_plan_id 
where s.delete_at =0
and s.org_id = ?
and av.status = ?
and EXISTS (
	select 
		1 
	from schedules_relations sr 
	where sr.schedule_id =s.id  
	and sr.relation_type = ?
	and sr.relation_id = ?
)
`
	args := []interface{}{
		op.OrgID,
		v2.AssessmentStatusComplete,
		entity.ScheduleRelationTypeClassRosterClass,
		classID,
	}
	err = r.QueryRawSQL(ctx, &items, sql, args...)
	if err != nil {
		return
	}
	return
}

type LearningOutcomeOverviewQueryCondition struct {
	From       int64    `json:"to"`
	To         int64    `json:"from"`
	TeacherIDs []string `json:"teacher_ids"`
}

type LearningOutcomeOverviewResult struct {
	Covered  int                                   `json:"covered"`
	Achieved []*entity.StudentOutcomeAchievedCount `json:"achieved"`
}

func (r *ReportDA) GetLearnerOutcomeOverview(ctx context.Context, condition *LearningOutcomeOverviewQueryCondition) (int, []*entity.StudentOutcomeAchievedCount, error) {
	if !config.Get().RedisConfig.OpenCache {
		tx := dbo.MustGetDB(ctx)
		covered, err := r.getCompleteLearnOutcomeCount(ctx, tx, condition)
		if err != nil {
			return 0, nil, err
		}

		achieved, err := r.getStudentAchievedOutcome(ctx, tx, condition)
		if err != nil {
			return 0, nil, err
		}

		return covered, achieved, nil
	}

	result := &LearningOutcomeOverviewResult{}
	err := r.learningOutcomeOverviewCache.Get(ctx, condition, result)
	if err != nil {
		return 0, nil, err
	}

	return result.Covered, result.Achieved, nil
}

type ILearningOutcomeReport interface {
	GetLessonPlanFilter(ctx context.Context, op *entity.Operator, classID string) (items []*entity.ScheduleShortInfo, err error)
	GetLearnerOutcomeOverview(ctx context.Context, condition *LearningOutcomeOverviewQueryCondition) (int, []*entity.StudentOutcomeAchievedCount, error)
	GetStudentOutcomeCount(ctx context.Context, op *entity.Operator, req entity.ListStudentsAchievementReportRequest) (items []*entity.StudentOutcomeCountItem, err error)
}

func (r *ReportDA) GetStudentOutcomeCount(ctx context.Context, op *entity.Operator, req entity.ListStudentsAchievementReportRequest) (items []*entity.StudentOutcomeCountItem, err error) {
	items = []*entity.StudentOutcomeCountItem{}
	sbAssessmentID := NewSqlBuilder(ctx, `
select 
		av.id
	from schedules s 
	inner join assessments_v2 av on av.schedule_id =s.id 
	
	where s.start_at < ?
	and s.lesson_plan_id in (
		select ?
		union all
		select id from cms_contents cc where cc.latest_id in (
			select latest_id from cms_contents where id=?  and latest_id not in ('','-')
		)
	)
	and EXISTS (
		select 1 from schedules_relations sr 
		where sr.schedule_id =s.id 
		and sr.relation_type in (?)
		and sr.relation_id = ?
	)
	and s.class_type in (?)
	and av.status = ?	
`,
		time.Now().Add(constant.ScheduleAllowGoLiveTime).Unix(),
		req.LessonPlanID,
		req.LessonPlanID,
		[]interface{}{
			entity.ScheduleRelationTypeClassRosterClass,
			entity.ScheduleRelationTypeParticipantClass,
		},
		req.ClassID,
		[]interface{}{
			entity.ScheduleClassTypeOnlineClass,
			entity.ScheduleClassTypeOfflineClass,
			entity.ScheduleClassTypeHomework,
		},
		v2.AssessmentStatusComplete,
	)
	sbSelect := NewSqlBuilder(ctx, `
auv.user_id as student_id,
	auov.outcome_id,
	sum(IF(auov.status=?,1,0)) as count_of_unknown,
  	sum(IF(auov.status=?,1,0)) as count_of_achieved,
  	sum(IF(auov.status=?,1,0)) as count_of_not_covered,
  	sum(IF(auov.status=?,1,0)) as count_of_not_achieved,
  	count(1) as count_of_all 
`,
		v2.AssessmentUserOutcomeStatusUnknown,
		v2.AssessmentUserOutcomeStatusAchieved,
		v2.AssessmentUserOutcomeStatusNotCovered,
		v2.AssessmentUserOutcomeStatusNotAchieved,
	)
	sql := `
select 
	{{.sbSelect}}
from assessments_users_outcomes_v2 auov  
inner join assessments_users_v2 auv on auv.id=auov.assessment_user_id  
inner join assessments_v2 av on av.id = auv.assessment_id 
inner join schedules s on s.id =  av.schedule_id 
where auv.assessment_id in (
	 {{.sbAssessmentID}}
)
{{.sbWhere}}
group by auv.user_id,auov.outcome_id
`
	sbWhere := NewSqlBuilder(ctx, ` 
and auv.user_type=? 
`, v2.AssessmentUserTypeStudent)

	sb := NewSqlBuilder(ctx, sql).
		Replace(ctx, "sbSelect", sbSelect).
		Replace(ctx, "sbAssessmentID", sbAssessmentID).
		Replace(ctx, "sbWhere", sbWhere)
	sql, args, err := sb.Build(ctx)
	if err != nil {
		return
	}
	err = r.QueryRawSQL(ctx, &items, sql, args...)
	if err != nil {
		return
	}
	return
}
func (r *ReportDA) getLearningOutcomeOverview(ctx context.Context, condition interface{}) (interface{}, error) {
	qc, ok := condition.(*LearningOutcomeOverviewQueryCondition)
	if !ok {
		log.Error(ctx, "invalid request", log.Any("condition", condition))
		return nil, constant.ErrInvalidArgs
	}

	tx := dbo.MustGetDB(ctx)
	covered, err := r.getCompleteLearnOutcomeCount(ctx, tx, qc)
	if err != nil {
		return nil, err
	}

	achieved, err := r.getStudentAchievedOutcome(ctx, tx, qc)
	if err != nil {
		return nil, err
	}

	return &LearningOutcomeOverviewResult{Covered: covered, Achieved: achieved}, nil
}

func (r *ReportDA) getStudentAchievedOutcome(ctx context.Context, tx *dbo.DBContext, condition *LearningOutcomeOverviewQueryCondition) (studentOutcomeAchievedCounts []*entity.StudentOutcomeAchievedCount, err error) {
	studentOutcomeAchievedCounts = []*entity.StudentOutcomeAchievedCount{}
	if len(condition.TeacherIDs) == 0 {
		return
	}

	sql := `
select 
	ass.student_id,
	count(1) as total_achieved_outcome_count,
	SUM(IF(auov.status='Achieved',1,0)) as achieved_outcome_count 
from (
	select auv.assessment_id , auv.user_id as student_id ,a.schedule_id ,auv.id as assessment_user_id
	from assessments_users_v2 auv
	left  join assessments_v2 a on auv.assessment_id = a.id 
	where 
	auv.user_type = ? 
	and a.complete_at >= ?
	and a.complete_at < ? 
) ass
inner JOIN assessments_users_outcomes_v2 auov on auov.assessment_user_id =ass.assessment_user_id
where EXISTS (
select relation_id from schedules_relations sr where sr.relation_type =? and sr.schedule_id =ass.schedule_id and sr.relation_id in(?)
)
group by ass.student_id
`
	args := []interface{}{
		v2.AssessmentUserTypeStudent,
		condition.From,
		condition.To,
		entity.ScheduleRelationTypeClassRosterTeacher,
		condition.TeacherIDs,
	}
	err = r.QueryRawSQL(ctx, &studentOutcomeAchievedCounts, sql, args...)
	if err != nil {
		return
	}
	return
}
func (r *ReportDA) getCompleteLearnOutcomeCount(ctx context.Context, tx *dbo.DBContext, condition *LearningOutcomeOverviewQueryCondition) (cnt int, err error) {
	if len(condition.TeacherIDs) == 0 {
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
		condition.From,
		condition.To,
		condition.From,
		condition.To,
		condition.TeacherIDs,
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
