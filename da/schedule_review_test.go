package da

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func TestScheduleReviewQuery(t *testing.T) {
	ctx := context.TODO()
	condition := &ScheduleReviewCondition{
		ScheduleIDs: entity.NullStrings{
			Valid:   true,
			Strings: []string{"614091d5e8155193e489a9ba"},
		},
		StudentIDs: entity.NullStrings{
			Valid:   true,
			Strings: []string{"1234", "12345"},
		},
		ReviewStatuses: entity.NullStrings{
			Valid:   true,
			Strings: []string{string(entity.ScheduleReviewStatusSuccess), string(entity.ScheduleReviewStatusPending), string(entity.ScheduleReviewStatusFailed)},
		},
		ReviewTypes: entity.NullStrings{
			Valid:   false,
			Strings: []string{string(entity.ScheduleReviewTypePersonalized)},
		},
	}

	var scheduleReviews []*entity.ScheduleReview
	err := GetScheduleReviewDA().Query(ctx, condition, &scheduleReviews)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(scheduleReviews)
}
