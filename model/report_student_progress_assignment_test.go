package model

import (
	"context"
	"fmt"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

func TestGetAssignmentCompletion(t *testing.T) {
	ctx := context.Background()
	request := entity.AssignmentRequest{
		ClassID:                 "0c01504d-d6ae-4c40-9862-68566bff0767",
		StudentID:               "4f614ccc-0867-5e5c-91f2-b71b895d2c48",
		SelectedSubjectIDList:   []string{""},
		UnSelectedSubjectIDList: []string{},
		Durations:               []entity.TimeRange{"0-1630016787", "01630016787-1630516787"},
	}
	res, err := GetReportModel().GetAssignmentCompletion(ctx, nil, &request)
	if err != nil {
		t.Fatal(err)
	}
	for i := range res.Assignments {
		fmt.Printf("%#v\n", res.Assignments[i])
	}
}
