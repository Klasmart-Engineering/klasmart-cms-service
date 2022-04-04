package external

import (
	"context"
	"testing"
)

func TestGetRoomInfoByRoomID(t *testing.T) {
	roomInfo, err := GetAssessmentServiceProvider().GetRoomInfoByRoomID(context.TODO(),
		testOperator,
		"62455082eb9f2a90ff31bbe7")
	if err != nil {
		t.Errorf("GetH5PRoomCommentServiceProvider().BatchGet() error = %v", err)
		return
	}

	t.Log(roomInfo)
}
