package external

import (
	"context"
	"testing"
)

func TestH5PRoomScoreService_BatchGet(t *testing.T) {
	scores, err := GetH5PRoomScoreServiceProvider().BatchGet(context.TODO(),
		testOperator,
		[]string{"60b49a40533e6a317aa2abb9"})
	if err != nil {
		t.Errorf("GetH5PRoomScoreServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(scores) == 0 {
		t.Error("GetH5PRoomScoreServiceProvider().BatchGet() get empty slice")
		return
	}

	for _, score := range scores {
		if len(score) == 0 {
			t.Error("GetH5PRoomScoreServiceProvider().BatchGet() get null")
			return
		}
	}
}

func TestH5PRoomScoreService_Set(t *testing.T) {
	score, err := GetH5PRoomScoreServiceProvider().Set(context.TODO(),
		testOperator,
		&H5PSetScoreRequest{
			RoomID:    "60a1d40a03b03c3acdb4f946",
			ContentID: "0a30bb0e-665f-443f-a80e-a1e29e5d1166",
			StudentID: "6ef232ce-5c37-4550-a8ca-8d27da5133f8",
			Score:     3.5,
		})
	if err != nil {
		t.Errorf("GetH5PRoomScoreServiceProvider().Set() error = %v", err)
		return
	}

	if score == nil {
		t.Error("GetH5PRoomScoreServiceProvider().Set() get empty result")
		return
	}
}

func TestH5PRoomScoreService_BatchSet(t *testing.T) {
	requests := []*H5PSetScoreRequest{
		{
			RoomID:    "60a1d40a03b03c3acdb4f946",
			ContentID: "0a30bb0e-665f-443f-a80e-a1e29e5d1166",
			StudentID: "6ef232ce-5c37-4550-a8ca-8d27da5133f8",
			Score:     3.5,
		},
		{
			RoomID:    "60a1d40a03b03c3acdb4f946",
			ContentID: "0a30bb0e-665f-443f-a80e-a1e29e5d1166",
			StudentID: "6ef232ce-5c37-4550-a8ca-8d27da5133f8",
			Score:     4,
		},
	}

	scoreResults, err := GetH5PRoomScoreServiceProvider().BatchSet(context.TODO(), testOperator, requests)
	if err != nil {
		t.Errorf("GetH5PRoomScoreServiceProvider().BatchSet() error = %v", err)
		return
	}

	if len(scoreResults) != len(requests) {
		t.Errorf("GetH5PRoomScoreServiceProvider().BatchSet() get invalid result, want %d, got %d", len(requests), len(scoreResults))
		return
	}
}
