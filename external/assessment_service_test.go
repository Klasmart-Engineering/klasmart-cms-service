package external

import (
	"context"
	"testing"
)

func TestGetRoomInfoByRoomID(t *testing.T) {
	roomInfo, err := GetAssessmentServiceProvider().Get(context.TODO(),
		testOperator,
		[]string{"62455082eb9f2a90ff31bbe7", "6243e56246d184f1e4b6cb77"})
	if err != nil {
		t.Errorf("GetH5PRoomCommentServiceProvider().BatchGet() error = %v", err)
		return
	}

	t.Log(roomInfo)
}

func TestAssessmentService_SetScoreAndCommentByRoomIDs(t *testing.T) {
	err := GetAssessmentServiceProvider().SetScoreAndComment(context.TODO(),
		testOperator,
		[]*SetScoreAndComment{
			&SetScoreAndComment{
				RoomID:    "6243e56246d184f1e4b6cb77",
				StudentID: "4831e78a-9a2f-45eb-8ca0-743642e674dc",
				Comment:   "333333",
				Scores: []*SetScore{
					&SetScore{
						ContentID:    "614d7854bbbecb046db012e1",
						SubContentID: "",
						Score:        2,
					},
				},
			},
		})

	t.Log(err)
}
