package external

import (
	"context"
	"testing"
)

func TestGetRoomInfoByRoomID(t *testing.T) {
	roomInfo, err := GetAssessmentServiceProvider().Get(context.TODO(),
		testOperator,
		[]string{"62ba9aa10d2a0680661661a1", "6243e56246d184f1e4b6cb77"},
		
		WithAssessmentGetScore(),
	)
	if err != nil {
		t.Errorf("GetH5PRoomCommentServiceProvider().BatchGet() error = %v", err)
		return
	}

	t.Log(roomInfo)
}
