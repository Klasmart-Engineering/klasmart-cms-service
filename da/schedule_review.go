package da

import (
	"context"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IScheduleReviewDA interface {
	dbo.DataAccesser
	GetScheduleReviewByScheduleIDAndStudentID(ctx context.Context, tx *dbo.DBContext, scheduleID, studentID string) (*entity.ScheduleReview, error)
	GetScheduleReviewsByScheduleID(ctx context.Context, tx *dbo.DBContext, scheduleID string) ([]*entity.ScheduleReview, error)
	UpdateScheduleReview(ctx context.Context, tx *dbo.DBContext, scheduleID, studentID string, status entity.ScheduleReviewStatus, reviewType entity.ScheduleReviewType, liveLessonPlan *entity.ScheduleLiveLessonPlan) error
	DeleteScheduleReviewByScheduleID(ctx context.Context, tx *dbo.DBContext, scheduleID string) error
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

func (s *scheduleReviewDA) GetScheduleReviewsByScheduleID(ctx context.Context, tx *dbo.DBContext, scheduleID string) ([]*entity.ScheduleReview, error) {
	tx.ResetCondition()

	var result []*entity.ScheduleReview
	err := tx.Where("schedule_id = ?", scheduleID).
		Find(&result).
		Error
	if err != nil {
		log.Error(ctx, "GetScheduleReviewByScheduleID error",
			log.Err(err),
			log.String("scheduleID", scheduleID),
		)
		return nil, err
	}

	return result, nil
}

func (s *scheduleReviewDA) UpdateScheduleReview(ctx context.Context, tx *dbo.DBContext,
	scheduleID, studentID string, status entity.ScheduleReviewStatus,
	reviewType entity.ScheduleReviewType, liveLessonPlan *entity.ScheduleLiveLessonPlan) error {
	tx.ResetCondition()

	err := tx.Table(constant.TableNameScheduleReview).
		Where("schedule_id = ? and student_id = ?", scheduleID, studentID).
		Updates(map[string]interface{}{
			"live_lesson_plan": liveLessonPlan,
			"review_status":    status,
			"type":             reviewType,
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

func (s *scheduleReviewDA) DeleteScheduleReviewByScheduleID(ctx context.Context, tx *dbo.DBContext, scheduleID string) error {
	tx.ResetCondition()
	err := tx.Where("schedule_id = ?", scheduleID).
		Delete(&entity.ScheduleReview{}).Error
	if err != nil {
		log.Error(ctx, "DeleteScheduleReviewByScheduleID error",
			log.Err(err),
			log.String("scheduleID", scheduleID),
		)
		return err
	}

	return nil
}

type ScheduleReviewCondition struct {
	ScheduleIDs    entity.NullStrings
	StudentIDs     entity.NullStrings
	ReviewStatuses entity.NullStrings
	ReviewTypes    entity.NullStrings
}

func (c ScheduleReviewCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.ScheduleIDs.Valid {
		wheres = append(wheres, "schedule_id in (?)")
		params = append(params, c.ScheduleIDs.Strings)
	}

	if c.StudentIDs.Valid {
		wheres = append(wheres, "student_id in (?)")
		params = append(params, c.StudentIDs.Strings)
	}

	if c.ReviewStatuses.Valid {
		wheres = append(wheres, "review_status in (?)")
		params = append(params, c.ReviewStatuses.Strings)
	}

	if c.ReviewTypes.Valid {
		wheres = append(wheres, "type in (?)")
		params = append(params, c.ReviewTypes.Strings)
	}

	return wheres, params
}

func (c ScheduleReviewCondition) GetOrderBy() string {
	return ""
}

func (c ScheduleReviewCondition) GetPager() *dbo.Pager {
	return nil
}
