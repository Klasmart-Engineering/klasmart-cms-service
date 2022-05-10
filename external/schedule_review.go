package external

import (
	"context"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type ScheduleReviewServiceProvider interface {
	CheckScheduleReview(ctx context.Context, operator *entity.Operator, checkScheduleReviewRequest CheckScheduleReviewRequest) (*CheckScheduleReviewResponse, error)
	CreateScheduleReview(ctx context.Context, operator *entity.Operator, createScheduleReviewRequest CreateScheduleReviewRequest) error
	DeleteScheduleReview(ctx context.Context, operator *entity.Operator, deleteScheduleReviewRequest DeleteScheduleReviewRequest) (*DeleteScheduleReviewResponse, error)
}

func GetScheduleReviewServiceProvider() ScheduleReviewServiceProvider {
	return &ScheduleReviewService{}
}

type ScheduleReviewService struct{}

func (s ScheduleReviewService) CreateScheduleReview(ctx context.Context, operator *entity.Operator, createScheduleReviewRequest CreateScheduleReviewRequest) error {

	response := &CreateScheduleReviewResponse{}

	statusCode, err := GetDataServiceClient().Run(ctx, constant.DataServiceCreateScheduleReviewUrlPath, &createScheduleReviewRequest, response)
	if err != nil {
		log.Error(ctx, "CreateScheduleReview error",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("createScheduleReviewRequest", createScheduleReviewRequest))
		return err
	}

	log.Debug(ctx, "CreateScheduleReview success",
		log.Int("statusCode", statusCode),
		log.Any("createScheduleReviewRequest", createScheduleReviewRequest),
		log.Any("response", response))

	return nil
}

func (s ScheduleReviewService) CheckScheduleReview(ctx context.Context, operator *entity.Operator, checkScheduleReviewRequest CheckScheduleReviewRequest) (*CheckScheduleReviewResponse, error) {

	response := &CheckScheduleReviewResponse{}

	statusCode, err := GetDataServiceClient().Run(ctx, constant.DataServiceCheckScheduleReviewUrlPath, &checkScheduleReviewRequest, response)
	if err != nil {
		log.Error(ctx, "CheckScheduleReview error",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("checkScheduleReviewRequest", checkScheduleReviewRequest))
		return nil, err
	}

	log.Debug(ctx, "CheckScheduleReview success",
		log.Int("statusCode", statusCode),
		log.Any("checkScheduleReviewRequest", checkScheduleReviewRequest),
		log.Any("response", response))

	return response, nil
}

func (s ScheduleReviewService) DeleteScheduleReview(ctx context.Context, operator *entity.Operator, deleteScheduleReviewRequest DeleteScheduleReviewRequest) (*DeleteScheduleReviewResponse, error) {

	response := &DeleteScheduleReviewResponse{}

	statusCode, err := GetDataServiceClient().Run(ctx, constant.DataServiceDeleteScheduleReviewUrlPath, &deleteScheduleReviewRequest, response)
	if err != nil {
		log.Error(ctx, "DeleteScheduleReview error",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("deleteScheduleReviewRequest", deleteScheduleReviewRequest))
		return nil, err
	}

	log.Debug(ctx, "DeleteScheduleReview success",
		log.Int("statusCode", statusCode),
		log.Any("deleteScheduleReviewRequest", deleteScheduleReviewRequest),
		log.Any("response", response))

	return response, nil
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

type CreateScheduleReviewResponse struct {
	ScheduleID string `json:"schedule_id"`
}

type CheckScheduleReviewRequest struct {
	TimeZoneOffset int64    `json:"time_zone_offset"`
	ProgramID      string   `json:"program_id"`
	SubjectIDs     []string `json:"subjects"`
	StudentIDs     []string `json:"student_ids"`
	ContentStartAt int64    `json:"content_start_at"`
	ContentEndAt   int64    `json:"content_end_at"`
}

type CheckScheduleReviewResponse struct {
	Results map[string]bool `json:"results"`
}

type DeleteScheduleReviewRequest struct {
	ScheduleIDs []string `json:"schedule_ids"`
}

type DeleteScheduleReviewResponse struct {
	Succeeded []string `json:"succeeded"`
	Failed    []string `json:"failed"`
}
