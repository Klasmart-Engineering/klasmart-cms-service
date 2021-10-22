package da

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IStudentProgressAssignment interface {
	ListAssignments(ctx context.Context, op *entity.Operator, args *entity.AssignmentRequest) ([]*entity.StudentAssignmentStatus, error)
}

func (r *ReportDA) durationFieldAndCondition(ctx context.Context, duration []entity.TimeRange) (field string, condition string, err error) {
	if len(duration) == 0 {
		field = "when 1 then 1"
		condition = "1"
		return
	}
	whenThen := make([]string, len(duration))
	conditions := make([]string, len(duration))
	for i, d := range duration {
		min, max, err := d.Value(ctx)
		if err != nil {
			log.Debug(ctx, "durationFieldAndCondition: value failed", log.Any("duration", d))
			return field, condition, err
		}
		wt := fmt.Sprintf(" when schedules.created_at >= %d and schedules.created_at < %d then '%s' ", min, max, string(d))
		whenThen[i] = wt

		c := fmt.Sprintf(" (schedules.created_at >= %d and schedules.created_at < %d) ", min, max)
		conditions[i] = c
	}
	return strings.Join(whenThen, "\n"), strings.Join(conditions, " or "), nil
}

func (r *ReportDA) ListAssignments(ctx context.Context, op *entity.Operator, args *entity.AssignmentRequest) ([]*entity.StudentAssignmentStatus, error) {
	sql := `
select
       sss.class_id,
       sss.student_id,
       sss.subject_id,
       count(sss.schedule_id) as total,
       sum(if(classes_assignments_records.finish_counts>0, 1, 0)) as finish,
       case
		   ${durationField}
       end
       as duration
from
    (
        select distinct sr1.schedule_id,
                        sr2.relation_id as class_id,
                        sr4.relation_id as student_id,
                        if(sr3.relation_id is null, '${NoCategory}', sr3.relation_id) as subject_id
        from schedules_relations sr1
                 join
             schedules_relations sr2
             on sr1.schedule_id = sr2.schedule_id and sr2.relation_type = '${class_roster_class}' and
                sr2.relation_id = ?		-- param: class_id
                 left join
             schedules_relations sr3
             on sr1.schedule_id = sr3.schedule_id and sr3.relation_type = '${Subject}'
                 join
             schedules_relations sr4
             on sr1.schedule_id = sr4.schedule_id and sr4.relation_type = '${class_roster_student}'
        where sr3.relation_id in (?)  -- param: subject_id
           or sr3.relation_id is null
    ) as sss
        left join
    classes_assignments_records
    on sss.schedule_id=classes_assignments_records.schedule_id and sss.student_id=classes_assignments_records.attendance_id
        join
    schedules
    on sss.schedule_id=schedules.id
where
      (${durationCondition}) 
group by class_id, student_id, subject_id, duration
`

	//schedules.class_type='${Homework}'
	//and (delete_at=0 or delete_at is null)

	sql = strings.Replace(sql, "${NoCategory}", "", -1)
	sql = strings.Replace(sql, "${class_roster_class}", entity.ScheduleRelationTypeClassRosterClass.String(), -1)
	sql = strings.Replace(sql, "${Subject}", entity.ScheduleRelationTypeSubject.String(), -1)
	sql = strings.Replace(sql, "${class_roster_student}", entity.ScheduleRelationTypeClassRosterStudent.String(), -1)
	sql = strings.Replace(sql, "${Homework}", entity.ScheduleClassTypeOnlineClass.String(), -1)

	durationFields, durationCondition, err := r.durationFieldAndCondition(ctx, args.Durations)
	if err != nil {
		log.Debug(ctx, "ListAssignments: duration failed", log.Any("args", args))
		return nil, err
	}
	sql = strings.Replace(sql, "${durationField}", durationFields, -1)
	sql = strings.Replace(sql, "${durationCondition}", durationCondition, -1)

	var result []*entity.StudentAssignmentStatus

	subjectID := make([]string, 0, len(args.SelectedSubjectIDList)+len(args.UnSelectedSubjectIDList))
	subjectID = append(subjectID, args.SelectedSubjectIDList...)
	subjectID = append(subjectID, args.UnSelectedSubjectIDList...)
	err = r.QueryRawSQL(ctx, &result, sql, args.ClassID, subjectID)
	if err != nil {
		log.Error(ctx, "ListTeacherLoadLessons: sql exe failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("args", args))
		return nil, err
	}
	return result, nil
}
