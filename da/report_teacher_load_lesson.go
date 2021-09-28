package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
)

type ITeacherLoadLesson interface {
	ListTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error)
	SummaryTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummary, error)
}

func (r *ReportDA) ListTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error) {
	sql := `
select
       teacher_id,
       count(if(is_attended=1 and class_type='${OnlineClass}', 1, null)) as live_completed_count,
       count(if(is_attended=1 and class_type='${OfflineClass}', 1, null)) as in_class_completed_count,
       count(if(is_attended=0 and class_type='${OnlineClass}', 1, null)) as live_missed_count,
       count(if(is_attended=0 and class_type='${OfflineClass}', 1, null)) as in_class_missed_count,
       count(1) as total_schedule
from (
     select distinct tsa.teacher_id, tsa.schedule_id, tsa.duration, tsa.class_type, if(assessments_attendances.attendance_id is null, 0, 1) as is_attended from
         (select teacher_schedule.teacher_id as teacher_id, teacher_schedule.schedule_id, teacher_schedule.duration, teacher_schedule.class_type, assessments.id as assessment_id from
             (select teacher.id as teacher_id, s.id as schedule_id, s.duration, s.class_type from
                 (select relation_id as id, schedule_id from schedules_relations
                  where relation_type in ('${class_roster_teacher}', '${participant_teacher}')
                    and relation_id in (?)   -- teacher_id params
                 ) as teacher
                     join
                 (select id, end_at-start_at as duration, class_type from schedules
                  where class_type in ('${OnlineClass}', '${OfflineClass}') and delete_at = 0
                    and end_at >= ? and end_at < ?  -- params
                    and exists (
                          select schedule_id
                          from schedules_relations
                          where schedules.id = schedule_id
                            and relation_type = '${class_roster_class}'
                            and relation_id in (?) -- class_id params
                      )
                 ) as s
                 on teacher.schedule_id = s.id
             ) as teacher_schedule
                 left join assessments on teacher_schedule.schedule_id=assessments.schedule_id
         ) as tsa
             left join assessments_attendances on tsa.teacher_id=assessments_attendances.attendance_id and tsa.assessment_id=assessments_attendances.assessment_id
) tl
group by teacher_id;
`
	sql = strings.Replace(sql, "${OnlineClass}", entity.ScheduleClassTypeOnlineClass.String(), -1)
	sql = strings.Replace(sql, "${OfflineClass}", entity.ScheduleClassTypeOfflineClass.String(), -1)
	sql = strings.Replace(sql, "${class_roster_teacher}", entity.ScheduleRelationTypeClassRosterTeacher.String(), -1)
	sql = strings.Replace(sql, "${participant_teacher}", entity.ScheduleRelationTypeParticipantTeacher.String(), -1)
	sql = strings.Replace(sql, "${class_roster_class}", entity.ScheduleRelationTypeClassRosterClass.String(), -1)

	start, end, err := args.Duration.Value(ctx)
	if err != nil {
		log.Error(ctx, "ListTeacherLoadLessons: time range value failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("args", args))
		return nil, err
	}
	var result []*entity.TeacherLoadLesson
	err = tx.Raw(sql, args.TeacherIDs, start, end, args.ClassIDs).Find(&result).Error
	if err != nil {
		log.Error(ctx, "ListTeacherLoadLessons: sql exe failed",
			log.Err(err),
			log.String("sql", sql),
			log.Any("args", args))
		return nil, err
	}
	return result, nil
}

func (r *ReportDA) SummaryTeacherLoadLessons(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummary, error) {
	panic("implement me")
}
