package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IScheduleReviewDA interface {
	dbo.DataAccesser
	GetScheduleReviewByScheduleIDAndStudentID(ctx context.Context, tx *dbo.DBContext, scheduleID, studentID string) (*entity.ScheduleReview, error)
	UpdateScheduleReview(ctx context.Context, tx *dbo.DBContext, scheduleID, studentID string, status entity.ScheduleReviewStatus, liveLessonPlan *entity.ScheduleLiveLessonPlan) error
}

type scheduleReviewDA struct {
	dbo.BaseDA
}

var (
	_scheduleReviewOnce sync.Once
	_scheduleReviewDA   IScheduleReviewDA
)

func GetScheduleReviewDA() IScheduleReviewDA {
	_scheduleReviewOnce.Do(func() {
		_scheduleReviewDA = &scheduleReviewDA{}
	})
	return _scheduleReviewDA
}

func (s *scheduleReviewDA) GetScheduleReviewByScheduleIDAndStudentID(ctx context.Context, tx *dbo.DBContext, scheduleID, studentID string) (*entity.ScheduleReview, error) {
	tx.ResetCondition()

	var result *entity.ScheduleReview
	err := tx.Where("schedule_id = ?", scheduleID).
		Where("student_id = ?", studentID).
		Find(&result).
		Error
	if err != nil {
		log.Error(ctx, "GetScheduleReviewByScheduleIDAndStudentID error",
			log.Err(err),
			log.String("scheduleID", scheduleID),
			log.String("studentID", studentID),
		)
		return nil, err
	}

	return result, nil
}

func (s *scheduleReviewDA) UpdateScheduleReview(ctx context.Context, tx *dbo.DBContext,
	scheduleID, studentID string, status entity.ScheduleReviewStatus,
	liveLessonPlan *entity.ScheduleLiveLessonPlan) error {
	tx.ResetCondition()

	err := tx.Table(constant.TableNameScheduleReview).
		Where("schedule_id = ? and student_id = ?", scheduleID, studentID).
		Updates(map[string]interface{}{
			"live_lesson_plan": liveLessonPlan,
			"review_status":    status,
		}).Error
	if err != nil {
		log.Error(ctx, "UpdateScheduleReview error",
			log.Err(err),
			log.String("scheduleID", scheduleID),
			log.String("studentID", studentID),
			log.Any("reviewStatus", status),
			log.Any("liveLessonPlan", liveLessonPlan))
		return err
	}

	return nil
}
