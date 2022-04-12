package v2

import (
	"encoding/json"
	"testing"
)

func TestAssessmentStudentResultReq(t *testing.T) {
	req := &AssessmentStudentResultReply{
		Answer:           "",
		Score:            0,
		Attempted:        false,
		ContentID:        "",
		Outcomes:         make([]*AssessmentStudentResultOutcomeReply, 1),
		StudentFeedBacks: nil,
		AssessScore:      0,
	}

	jb, _ := json.Marshal(req)
	t.Log(string(jb))
}
