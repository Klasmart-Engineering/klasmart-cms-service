package external

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ScheduleReviewServiceProvider interface {
	CreateScheduleReview(ctx context.Context, operator *entity.Operator, createScheduleReviewRequest CreateScheduleReviewRequest) error
}

func GetScheduleReviewServiceProvider() ScheduleReviewServiceProvider {
	return &ScheduleReviewService{}
}

type ScheduleReviewService struct{}

func (s ScheduleReviewService) CreateScheduleReview(ctx context.Context, operator *entity.Operator, createScheduleReviewRequest CreateScheduleReviewRequest) error {
	// TODO: implement
	return nil
}

type CreateScheduleReviewRequest struct {
	ScheduleID     string   `json:"schedule_id"`
	DueAt          int64    `json:"due_at"`
	TimeZoneOffset int64    `json:"time_zone_offset"`
	ProgramID      string   `json:"program_id"`
	SubjectIDs     []string `json:"subject_ids"`
	ClassID        string   `json:"class_id"`
	StudentIDs     []string `json:"student_ids"`
	ContentStartAt int64    `json:"content_start_at"`
	ContentEndAt   int64    `json:"content_end_at"`
}
