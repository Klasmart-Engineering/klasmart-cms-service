package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IScheduleReviewDA interface {
	dbo.DataAccesser
	GetScheduleReviewByScheduleIDAndStudentID(ctx context.Context, tx *dbo.DBContext, scheduleID, studentID string) (*entity.ScheduleReview, error)
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
