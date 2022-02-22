package assessmentV2

import (
	"context"
	"fmt"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type OfflineStudyOutcomeDBView struct {
	ScheduleID   string `gorm:"schedule_id"`
	StudentID    string `gorm:"student_id"`
	OutcomeID    string `gorm:"outcome_id"`
	Checked      bool   `gorm:"checked"`
	NoneAchieved bool   `gorm:"none_achieved"`
	Skip         bool   `gorm:"skip"`
}

type NotOfflineStudyOutcomeDBView struct {
	AssessmentID string `gorm:"assessment_id"`
	StudentID    string `gorm:"student_id"`
	OutcomeID    string `gorm:"outcome_id"`
	ContentID    string `gorm:"content_id"`
	Checked      bool   `gorm:"checked"`
	NoneAchieved bool   `gorm:"none_achieved"`
	Skip         bool   `gorm:"skip"`
}

type IAssessmentMigrateDA interface {
	dbo.DataAccesser
	GetNoAssessmentSchedule(ctx context.Context) ([]*entity.Schedule, error)
	GetOfflineStudyOutcome(ctx context.Context) ([]*OfflineStudyOutcomeDBView, error)
	GetNotOfflineStudyOutcome(ctx context.Context) ([]*NotOfflineStudyOutcomeDBView, error)
}

var (
	assessmentMigrateDAInstance     IAssessmentMigrateDA
	assessmentMigrateDAInstanceOnce = sync.Once{}
)

func GetAssessmentMigrateDA() IAssessmentMigrateDA {
	assessmentMigrateDAInstanceOnce.Do(func() {
		assessmentMigrateDAInstance = &assessmentMigrateDA{}
	})
	return assessmentMigrateDAInstance
}

type assessmentMigrateDA struct {
	dbo.BaseDA
}

func (a *assessmentMigrateDA) GetNotOfflineStudyOutcome(ctx context.Context) ([]*NotOfflineStudyOutcomeDBView, error) {
	tx := dbo.MustGetDB(ctx)
	tx.ResetCondition()

	sql := fmt.Sprintf(`
select t1.assessment_id,t1.outcome_id,t2.attendance_id student_id,t2.content_id, t1.checked,t1.none_achieved,t1.skip 
from assessments_outcomes t1 left join contents_outcomes_attendances t2 on t1.assessment_id=t2.assessment_id and t1.outcome_id = t2.outcome_id
where t2.content_id is not null 
`)

	var result []*NotOfflineStudyOutcomeDBView
	err := tx.Raw(sql).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *assessmentMigrateDA) GetOfflineStudyOutcome(ctx context.Context) ([]*OfflineStudyOutcomeDBView, error) {
	tx := dbo.MustGetDB(ctx)
	tx.ResetCondition()

	sql := fmt.Sprintf(`
select t1.schedule_id,t1.student_id,t2.outcome_id,t2.checked,t2.none_achieved,t2.skip 
from home_fun_studies t1 inner join assessments_outcomes t2 on t1.id = t2.assessment_id
`)

	var result []*OfflineStudyOutcomeDBView
	err := tx.Raw(sql).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *assessmentMigrateDA) GetNoAssessmentSchedule(ctx context.Context) ([]*entity.Schedule, error) {
	tx := dbo.MustGetDB(ctx)
	tx.ResetCondition()

	sql := fmt.Sprintf(`
select t1.* from %s t1 
left join %s t2 
on t1.id = t2.schedule_id 
where t2.id is null and t1.class_type in (?);
`,
		constant.TableNameSchedule,
		constant.TableNameAssessmentV2,
	)

	var schedules []*entity.Schedule
	err := tx.Raw(sql, []string{
		entity.ScheduleClassTypeHomework.String(),
		entity.ScheduleClassTypeOfflineClass.String(),
		entity.ScheduleClassTypeOnlineClass.String(),
	}).Scan(&schedules).Error
	if err != nil {
		return nil, err
	}

	return schedules, nil
}
