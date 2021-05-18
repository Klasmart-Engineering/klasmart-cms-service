package external

import (
	"context"
	"testing"
)

func TestH5PRoomScoreService_BatchGet(t *testing.T) {
	scores, err := GetH5PRoomScoreServiceProvider().BatchGet(context.TODO(),
		testOperator,
		[]string{"60a1d40a03b03c3acdb4f946"})
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
