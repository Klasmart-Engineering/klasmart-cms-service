package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"strings"

	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ILearnerWeekly interface {
	GetLearnerWeeklyReportOverview(ctx context.Context, op *entity.Operator, tr entity.TimeRange, cond entity.GetUserCountCondition) (res entity.LearnerReportOverview, err error)
	QueryLiveClassesSummaryV2(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (res []*entity.LiveClassSummaryItemV2, err error)
	QueryAssignmentsSummaryV2(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, filter *entity.LearningSummaryFilter) (items []*entity.AssignmentsSummaryItemV2, err error)
	QueryOutcomesByAssessmentID(ctx context.Context, op *entity.Operator, assessmentID string, studentID string) (items []*entity.LearningSummaryOutcomeItem, err error)
}

func (r *ReportDA) QueryOutcomesByAssessmentID(ctx context.Context, op *entity.Operator, assessmentID string, studentID string) (items []*entity.LearningSummaryOutcomeItem, err error) {
	items = []*entity.LearningSummaryOutcomeItem{}
	sql := `
	select lo.id,lo.name,t.* from (
	      	select
	      	auov.outcome_id,
	      	sum(IF(auov.status='Unknown',1,0)) as count_of_unknown,
	      	sum(IF(auov.status='Achieved',1,0)) as count_of_achieved,
	      	sum(IF(auov.status='NotCovered',1,0)) as count_of_not_covered,
	      	sum(IF(auov.status='NotAchieved',1,0)) as count_of_not_achieved,
	      	count(1) as count_of_all
	      from assessments_users_outcomes_v2 auov
	      inner join assessments_users_v2 auv  on auov.assessment_user_id =auv.id
	      where auv.assessment_id = ? and auv.user_type = ?  and auv.user_id = ?
	      group by auov.outcome_id
	      ) t
	      left join learning_outcomes lo on lo.id =t.outcome_id
`
	sb := NewSqlBuilder(ctx, sql, assessmentID, v2.AssessmentUserTypeStudent, studentID)
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
func (r *ReportDA) QueryAssignmentsSummaryV2(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (items []*entity.AssignmentsSummaryItemV2, err error) {
	items = []*entity.AssignmentsSummaryItemV2{}
	sbSchedule, err := r.getSqlBuilderOfSchedulesByLearningSummaryFilter(ctx, operator, filter, entity.LearningSummaryTypeLiveClass)
	if err != nil {
		return
	}

	sbStatus := NewSqlBuilder(ctx, `
IF(av.status=?,?,?) as status,
`, v2.AssessmentStatusComplete, entity.AssessmentStatusComplete, entity.AssessmentStatusInProgress)
	sbAssessmentType := NewSqlBuilder(ctx, `
IF(s.is_home_fun=1, ? , ? ) as assessment_type,`, entity.AssessmentTypeHomeFunStudy, entity.AssessmentTypeStudy)
	sbWhere := NewSqlBuilder(ctx, `
where s.id in ({{.sbSchedule}})`).Replace(ctx, "sbSchedule", sbSchedule)
	sql := `
select 
	{{.sbAssessmentType}}
	{{.sbStatus}},
	av.title  as assessment_title,
	cc.content_name as lesson_plan_name,	 
	av.schedule_id,
	av.id as assessment_id,
	av.complete_at,
	av.create_at
from assessments_v2 av 
inner join schedules s on s.id = av.schedule_id 
inner join cms_contents cc on cc.id = s.lesson_plan_id 
{{.sbWhere}}
`
	sb := NewSqlBuilder(ctx, sql).
		Replace(ctx, "sbAssessmentType", sbAssessmentType).
		Replace(ctx, "sbStatus", sbStatus).
		Replace(ctx, "sbWhere", sbWhere)
	sql, args, err := sb.Build(ctx)
	if err != nil {
		return
	}
	err = r.QueryRawSQLTx(ctx, tx, &items, sql, args...)
	if err != nil {
		return
	}

	return
}

func (r *ReportDA) getSqlBuilderOfSchedulesByLearningSummaryFilter(ctx context.Context, operator *entity.Operator, filter *entity.LearningSummaryFilter, typ entity.LearningSummaryType) (sb *sqlBuilder, err error) {
	sqlSchedule := strings.Builder{}
	sqlSchedule.WriteString(`
select 
	s.id 
from schedules s 
where s.org_id =  ?
and s.class_type in (?) 
and exists (
	select
		1
	from
		schedules_relations sr
	where
		sr.relation_id in (?)
		and sr.relation_type in (?,?)
		and s.id = sr.schedule_id)
and (s.delete_at = 0) 
`)
	argsSchedule := []interface{}{
		operator.OrgID,
		entity.ScheduleClassTypeOnlineClass,
		filter.StudentID,
		entity.ScheduleRelationTypeClassRosterStudent,
		entity.ScheduleRelationTypeParticipantStudent,
	}
	switch typ {
	case entity.LearningSummaryTypeLiveClass:
		sqlSchedule.WriteString(`and s.class_type = ? `)
		argsSchedule = append(argsSchedule, entity.ScheduleClassTypeOnlineClass)
		if filter.WeekStart > 0 {
			sqlSchedule.WriteString(`and s.start_at >= ? `)
			argsSchedule = append(argsSchedule, filter.WeekStart)
		}
		if filter.WeekEnd > 0 {
			sqlSchedule.WriteString(`and s.start_at < ? `)
			argsSchedule = append(argsSchedule, filter.WeekEnd)
		}
	case entity.LearningSummaryTypeAssignment:
		sqlSchedule.WriteString(`and s.class_type = ? `)
		argsSchedule = append(argsSchedule, entity.ScheduleClassTypeHomework)
		if filter.WeekStart > 0 {
			sqlSchedule.WriteString(`and s.complete_time >= ? `)
			argsSchedule = append(argsSchedule, filter.WeekStart)
		}
		if filter.WeekEnd > 0 {
			sqlSchedule.WriteString(`and s.complete_time < ? `)
			argsSchedule = append(argsSchedule, filter.WeekEnd)
		}
	}

	if len(filter.SchoolIDs) > 0 {
		if filter.SchoolIDs[0] == constant.LearningSummaryFilterOptionNoneID {
			var classes []*external.NullableClass
			classes, err = external.GetClassServiceProvider().GetOnlyUnderOrgClasses(ctx, operator, operator.OrgID)
			if err != nil {
				log.Error(ctx, "find related schedules: get only under org classes failed",
					log.Err(err),
					log.Any("operator", operator),
				)
				return
			}
			classIDs := make([]string, 0, len(classes))
			for _, c := range classes {
				classIDs = append(classIDs, c.ID)
			}
			sqlSchedule.WriteString(`
and EXISTS(
	select 
		1 
	from schedules_relations sr 
	where sr.relation_id in (?) 
	and sr.relation_type in (?) 
	and s.id = sr.schedule_id
)
`)
			argsSchedule = append(argsSchedule, classIDs, []interface{}{entity.ScheduleRelationTypeClassRosterClass, entity.ScheduleRelationTypeParticipantClass})
		}

		sqlSchedule.WriteString(`
and EXISTS (
	select 
		1 
	from schedules_relations sr 
	where  sr.schedule_id = s.id 
	and  sr.relation_type = ?  
	and sr.relation_id in (?)
)`)
		argsSchedule = append(argsSchedule, entity.ScheduleRelationTypeSchool, filter.SchoolIDs)

	}

	if len(filter.ClassID) > 0 {
		if filter.ClassID == constant.LearningSummaryFilterOptionNoneID {
			var users []*external.User
			users, err = external.GetUserServiceProvider().GetOnlyUnderOrgUsers(ctx, operator, operator.OrgID)
			if err != nil {
				log.Error(ctx, "find related schedules: get only under org users failed",
					log.Err(err),
					log.Any("operator", operator),
				)
				return
			}
			userIDs := make([]string, 0, len(users))
			for _, u := range users {
				userIDs = append(userIDs, u.ID)
			}

			sqlSchedule.WriteString(`
and EXISTS (
	select 
		1 
	from schedules_relations sr 
	where  sr.schedule_id = s.id 
	and  sr.relation_type in (?)  
	and sr.relation_id in (?)
)`)
			argsSchedule = append(argsSchedule, []interface{}{
				entity.ScheduleRelationTypeClassRosterTeacher,
				entity.ScheduleRelationTypeParticipantTeacher,
				entity.ScheduleRelationTypeClassRosterStudent,
				entity.ScheduleRelationTypeParticipantStudent,
			}, userIDs)
		} else {
			sqlSchedule.WriteString(`
and EXISTS (
	select 
		1 
	from schedules_relations sr 
	where  sr.schedule_id = s.id 
	and  sr.relation_type in (?)  
	and sr.relation_id in (?)
)`)
			argsSchedule = append(argsSchedule, []interface{}{
				entity.ScheduleRelationTypeClassRosterClass,
				entity.ScheduleRelationTypeParticipantClass,
			}, filter.ClassID)
		}

	}
	if len(filter.TeacherID) > 0 {
		sqlSchedule.WriteString(`
and EXISTS (
	select 
		1 
	from schedules_relations sr 
	where  sr.schedule_id = s.id 
	and  sr.relation_type in (?)  
	and sr.relation_id in (?)
)`)
		argsSchedule = append(argsSchedule, []interface{}{
			entity.ScheduleRelationTypeClassRosterTeacher,
			entity.ScheduleRelationTypeParticipantTeacher,
		}, filter.TeacherID)
	}

	if len(filter.StudentID) > 0 {
		sqlSchedule.WriteString(`
and EXISTS (
	select 
		1 
	from schedules_relations sr 
	where  sr.schedule_id = s.id 
	and  sr.relation_type in (?)  
	and sr.relation_id in (?)
)`)
		argsSchedule = append(argsSchedule, []interface{}{
			entity.ScheduleRelationTypeClassRosterStudent,
			entity.ScheduleRelationTypeParticipantStudent,
		}, filter.StudentID)
	}
	if len(filter.SubjectID) > 0 {
		sqlSchedule.WriteString(`
and EXISTS (
	select 
		1 
	from schedules_relations sr 
	where  sr.schedule_id = s.id 
	and  sr.relation_type in (?)  
	and sr.relation_id in (?)
)`)
		argsSchedule = append(argsSchedule, []interface{}{
			entity.ScheduleRelationTypeSubject,
		}, filter.SubjectID)
	}
	sb = NewSqlBuilder(ctx, sqlSchedule.String(), argsSchedule...)
	return
}
func (r *ReportDA) QueryLiveClassesSummaryV2(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, filter *entity.LearningSummaryFilter) (res []*entity.LiveClassSummaryItemV2, err error) {
	res = []*entity.LiveClassSummaryItemV2{}
	sbSchedule, err := r.getSqlBuilderOfSchedulesByLearningSummaryFilter(ctx, operator, filter, entity.LearningSummaryTypeLiveClass)
	if err != nil {
		return
	}
	sbAssessment := NewSqlBuilder(ctx, `
select av.id from assessments_v2 av where av.schedule_id in ({{.sbSchedule}}) 
`).Replace(ctx, "sbSchedule", sbSchedule)

	sbStatus := NewSqlBuilder(ctx, `
IF(av.status=?,?,?) as status,
`, v2.AssessmentStatusComplete, entity.AssessmentStatusComplete, entity.AssessmentStatusInProgress)
	sbAbsentSelect := NewSqlBuilder(ctx, `
IF(sum(IF(auv.status_by_system =?,1,0))>0,0,1) as absent
`, v2.AssessmentUserStatusParticipate)
	sbAbsent := NewSqlBuilder(ctx, `
select 
	auv.assessment_id,
	{{.sbAbsentSelect}}
from   assessments_users_v2 auv 
where  auv.assessment_id in ({{.sbAssessment}})
group by auv.assessment_id 
`).Replace(ctx, "sbAbsentSelect", sbAbsentSelect).Replace(ctx, "sbAssessment", sbAssessment)
	sbWhere := NewSqlBuilder(ctx, `
and auv.user_type = ?
and auv.user_id in (?)
`, v2.AssessmentUserTypeStudent, filter.StudentID)
	sb := NewSqlBuilder(ctx, `
select 
	{{.sbStatus}}	
	auv.absent,
	s.start_at as class_start_time,
	s.title as schedule_title,
	cc.content_name as lesson_plan_name,
	s.id as schedule_id,
	av.id as assessment_id,
	av.complete_at,
	av.create_at
from (
	{{.sbAbsent}}
) auv 
inner join assessments_v2 av on auv.assessment_id =av.id 
inner join schedules s on av.schedule_id =s.id 
inner join cms_contents cc on cc.id = s.lesson_plan_id 
where auv.assessment_id in ({{.sbAssessment}})
{{.sbWhere}}
order by s.start_at
`).
		Replace(ctx, "sbStatus", sbStatus).
		Replace(ctx, "sbAbsent", sbAbsent).
		Replace(ctx, "sbWhere", sbWhere).
		Replace(ctx, "sbAssessment", sbAssessment)
	sql, args, err := sb.Build(ctx)
	if err != nil {
		return
	}
	err = r.QueryRawSQLTx(ctx, tx, &res, sql, args...)
	if err != nil {
		return
	}

	return
}
func (r *ReportDA) GetLearnerWeeklyReportOverview(ctx context.Context, op *entity.Operator, tr entity.TimeRange, cond entity.GetUserCountCondition) (res entity.LearnerReportOverview, err error) {
	sqlSchedule := strings.Builder{}
	sqlSchedule.WriteString(`
	select id from schedules s 
	where ((s.class_type=? and s.end_at >= ? and s.end_at <?) or (s.class_type=? and s.created_at >= ? and s.created_at <?))
	and s.org_id = ?
`)
	start, end, err := tr.Value(ctx)
	if err != nil {
		return
	}
	argsSchedule := []interface{}{
		entity.ScheduleClassTypeOnlineClass,
		start,
		end,
		entity.ScheduleClassTypeHomework,
		start,
		end,
		op.OrgID,
	}

	if cond.SchoolIDs.Valid {
		sqlSchedule.WriteString(`
and EXISTS (
	select * from schedules_relations sr 
	where sr.relation_type = ? 
	and sr.relation_id in (?)	
	and sr.schedule_id = s.id 
)`)
		argsSchedule = append(argsSchedule, entity.ScheduleRelationTypeSchool, cond.SchoolIDs.Strings)
	}
	if cond.ClassIDs.Valid {
		sqlSchedule.WriteString(`
and EXISTS (
	select * from schedules_relations sr 
	where sr.relation_type = ? 
	and sr.relation_id in (?)	
	and sr.schedule_id = s.id 
)`)
		argsSchedule = append(argsSchedule, entity.ScheduleRelationTypeClassRosterClass, cond.ClassIDs.Strings)
	}
	if cond.StudentID.Valid {
		sqlSchedule.WriteString(`
and EXISTS (
	select * from schedules_relations sr 
	where sr.relation_type = ? 
	and sr.relation_id in (?)	
	and sr.schedule_id = s.id
)`)
		argsSchedule = append(argsSchedule, entity.ScheduleRelationTypeClassRosterStudent, cond.StudentID.String)
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
{{.sbSelectRate}}
from assessments_users_v2 
where assessment_id in({{.sbAssessmentOnlineClass}})
{{.sbWhereUserType}}
{{.sbWhereUserID}}
group by user_id 

union all 

select
user_id as student_id,
'study' as typ,
{{.sbSelectRate}}
from assessments_users_v2 
where assessment_id in({{.sbAssessmentStudy}})
{{.sbWhereUserType}}
{{.sbWhereUserID}}
group by user_id 
`
	sbWhereUserType := NewSqlBuilder(ctx, "and user_type =?", v2.AssessmentUserTypeStudent)
	sbWhereUserID := NewSqlBuilder(ctx, "")
	if cond.StudentID.Valid {
		sbWhereUserID = NewSqlBuilder(ctx, "and user_id =?", cond.StudentID.String)
	}
	sbSelectRate := NewSqlBuilder(ctx, "sum(if(status_by_system=?,1,0))/count(1) as rate ", v2.AssessmentUserStatusParticipate)
	sb := NewSqlBuilder(ctx, sql).
		Replace(ctx, "sbAssessmentOnlineClass", sbAssessmentOnlineClass).
		Replace(ctx, "sbAssessmentStudy", sbAssessmentStudy).
		Replace(ctx, "sbSelectRate", sbSelectRate).
		Replace(ctx, "sbWhereUserType", sbWhereUserType).
		Replace(ctx, "sbWhereUserID", sbWhereUserID)
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
		res.Status = constant.LearnerReportOverviewStatusNoData
		return
	}
	m := map[string][]float64{}
	for _, item := range *ret {

		m[item.StudentID] = append(m[item.StudentID], item.Rate)
	}

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
			res.NumAbove++
		} else if rate0 < 0.49 && rate1 < 0.49 {
			res.NumBelow++
		} else {
			res.NumMeet++
		}
	}
	return
}
