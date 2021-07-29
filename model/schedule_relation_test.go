package model

import (
	"context"
	"testing"
)

func TestGetOutcomeIDs(t *testing.T) {
	outcomeIDs, err := GetScheduleRelationModel().GetOutcomeIDs(context.TODO(), "60f92bd0f964d549bf922b46")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(outcomeIDs)
}
