package external

import (
	"context"
	"testing"
)

func TestGetRoomInfoByRoomID(t *testing.T) {
	roomInfo, err := GetAssessmentServiceProvider().Get(context.TODO(),
		testOperator,
		[]string{"62455082eb9f2a90ff31bbe7", "6243e56246d184f1e4b6cb77"},
		WithAssessmentGetScore(false),
		WithAssessmentGetTeacherComment(false))
	if err != nil {
		t.Errorf("GetH5PRoomCommentServiceProvider().BatchGet() error = %v", err)
		return
	}

	t.Log(roomInfo)
}
