package da

import (
	"context"
	"strings"

	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ILearnerWeekly interface {
	GetLearnerWeeklyReportOverview(ctx context.Context, op *entity.Operator, tr entity.TimeRange, cond entity.GetUserCountCondition) (res entity.LearnerWeeklyReportOverview, err error)
}

func (r *ReportDA) GetLearnerWeeklyReportOverview(ctx context.Context, op *entity.Operator, tr entity.TimeRange, cond entity.GetUserCountCondition) (res entity.LearnerWeeklyReportOverview, err error) {
	sqlSchedule := strings.Builder{}
	sqlSchedule.WriteString(`
select id from schedules s 
where s.class_type in (?) 
and s.end_at >= ? and s.end_at <?
and s.org_id = ?
`)
	var argsSchedule []interface{}
	argsSchedule = append(argsSchedule, []interface{}{
		entity.ScheduleClassTypeOnlineClass,
		entity.ScheduleClassTypeHomework,
	})
	start, end, err := tr.Value(ctx)
	if err != nil {
		return
	}
	argsSchedule = append(argsSchedule, start, end, op.OrgID)
	if cond.SchoolIDs.Valid {
		sqlSchedule.WriteString(`
and EXISTS (
	select * from schedules_relations sr 
	where sr.relation_type = ? 
	and sr.relation_id in (?)	
)`)
		argsSchedule = append(argsSchedule, entity.ScheduleRelationTypeSchool, cond.SchoolIDs.Strings)
	}
	if cond.ClassIDs.Valid {
		sqlSchedule.WriteString(`
and EXISTS (
	select * from schedules_relations sr 
	where sr.relation_type = ? 
	and sr.relation_id in (?)	
)`)
		argsSchedule = append(argsSchedule, entity.ScheduleRelationTypeClassRosterClass, cond.ClassIDs.Strings)
	}

	sbSchedule := NewSqlBuilder(ctx, sqlSchedule.String(), argsSchedule...)
	sbAssessmentTypeOnlineClass := NewSqlBuilder(ctx, `av.assessment_type=?`, v2.AssessmentTypeOnlineClass)
	sbAssessmentOnlineClass := NewSqlBuilder(ctx, `
	select id from assessments_v2 av  where {{.sbAssessmentTypeOnlineClass}} and av.schedule_id in ({{.sbSchedule}})`).
		Replace(ctx, "sbSchedule", sbSchedule).
		Replace(ctx, "sbAssessmentTypeOnlineClass", sbAssessmentTypeOnlineClass)

	sbAssessmentTypeStudy := NewSqlBuilder(ctx, `av.assessment_type in ( ? , ? )`, v2.AssessmentTypeOnlineStudy, v2.AssessmentTypeOfflineStudy)
	sbAssessmentStudy := NewSqlBuilder(ctx, `
	select id from assessments_v2 av  where {{.sbAssessmentTypeStudy}} and av.schedule_id in ({{.sbSchedule}}) `).
		Replace(ctx, "sbSchedule", sbSchedule).
		Replace(ctx, "sbAssessmentTypeStudy", sbAssessmentTypeStudy)
	sql := `
select
user_id as student_id,
'online_class' as typ,
sum(if(status_by_system='Participate',1,0))/count(1) as rate 
from assessments_users_v2 
where assessment_id in({{.sbAssessmentOnlineClass}})
and user_type ='Student'
group by user_id 

union all 

select
user_id as student_id,
'study' as typ,
sum(if(status_by_system='Participate',1,0))/count(1) as rate 
from assessments_users_v2 
where assessment_id in({{.sbAssessmentStudy}})
and user_type ='Student'
group by user_id 
`
	sbSelectRate := NewSqlBuilder(ctx, "sum(if(status_by_system=?,1,0))/count(1) as rate ", v2.AssessmentUserStatusParticipate)
	sb := NewSqlBuilder(ctx, sql).
		Replace(ctx, "sbAssessmentOnlineClass", sbAssessmentOnlineClass).
		Replace(ctx, "sbAssessmentStudy", sbAssessmentStudy).
		Replace(ctx, "sbSelectRate", sbSelectRate)
	sql, args, err := sb.Build(ctx)
	if err != nil {
		return
	}
	ret := &[]struct {
		StudentID string
		Typ       string
		Rate      float64
	}{}
	err = r.QueryRawSQL(ctx, ret, sql, args...)
	if err != nil {
		return
	}
	if len(*ret) == 0 {
		res.Status = constant.LearnerWeeklyReportOverviewStatusNoData
		return
	}
	m := map[string][]float64{}
	for _, item := range *ret {
		m[item.StudentID] = append(m[item.StudentID], item.Rate)
	}

	var numAbove float64
	var numBelow float64
	var numMeet float64
	for _, rates := range m {
		var rate0 float64
		var rate1 float64
		switch len(rates) {
		case 0:
			continue
		case 1:
			rate0 = rates[0]
			rate1 = 1
		default:
			rate0, rate1 = rates[0], rates[1]
		}

		if rate0 >= 0.8 && rate1 >= 0.8 {
			numAbove++
		} else if rate0 < 0.49 && rate1 < 0.49 {
			numBelow++
		} else {
			numMeet++
		}
	}
	res.AboveExpectation = numAbove / (numAbove + numBelow + numMeet)
	res.BelowExpectation = numBelow / (numAbove + numBelow + numMeet)
	res.MeetExpectation = numMeet / (numAbove + numBelow + numMeet)
	return
}
