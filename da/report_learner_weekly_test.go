package da

import (
	"context"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

func TestReportDA_GetLearnerWeeklyReportOverview(t *testing.T) {
	ctx := context.Background()
	reportDA := GetReportDA()

	condition := &entity.GetUserCountCondition{
		OrgID: entity.NullString{String: "6300b3c5-8936-497e-ba1f-d67164b59c65", Valid: true},
	}

	op := &entity.Operator{
		OrgID: condition.OrgID.String,
	}

	tr := entity.TimeRange("1648742400-1649325198")

	_, err := reportDA.GetLearnerWeeklyReportOverview(ctx, op, tr, condition)
	if err != nil {
		t.Errorf("get learner report overview failed due to %v", err)
	}
}

func BenchmarkReportDA_GetLearnerWeeklyReportOverview(b *testing.B) {
	ctx := context.Background()
	reportDA := GetReportDA()

	condition := &entity.GetUserCountCondition{
		OrgID: entity.NullString{String: "6300b3c5-8936-497e-ba1f-d67164b59c65", Valid: true},
	}

	op := &entity.Operator{
		OrgID: condition.OrgID.String,
	}

	tr := entity.TimeRange("1648742400-1649325198")

	for n := 0; n < b.N; n++ {
		_, err := reportDA.GetLearnerWeeklyReportOverview(ctx, op, tr, condition)
		if err != nil {
			b.Errorf("get learner report overview failed due to %v", err)
		}
	}
}
