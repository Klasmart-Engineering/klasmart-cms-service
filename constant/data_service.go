package constant

import (
	"errors"
	"time"
)

const (
	DataServiceAuthorizedHeaderKey = "x-api-key"

	DataServiceHttpTimeout = time.Minute

	DataServiceCreateScheduleReviewUrlPath = "/schedule_review"
	DataServiceCheckScheduleReviewUrlPath  = "/check_review_data"
	DataServiceDeleteScheduleReviewUrlPath = "/delete_schedules"
)

var (
	ErrDataServiceFailed = errors.New("data service failed")
)
