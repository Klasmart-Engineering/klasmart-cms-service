package model

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func TestAssessmentOfflineStudyModel_IsAnyOneCompleteByScheduleIDs(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE3YTI4MzM4LTNiODgtNGJhYy1hYjE1LWNjZTM4ODdhZjM1NyIsImVtYWlsIjoib3JnbWlAeW9wbWFpbC5jb20iLCJleHAiOjE2Mzk5ODMwMTAsImlzcyI6ImtpZHNsb29wIn0.Pd8hG7ChoH3Gduxx8o16PkupaqxDtanc82-oCQv9mntuVhm6XRyNrC-6pt7TRLyKYPMLBBKryBtDLxeItggENWgvA_V7Z_0puRo1MTg14nRxx72uQ2TAdnB08IVrCh-Qc5OtsdfnGjIzKCIEaSbGeeGdyaF3rHDg2RgeCAAmbli7APrpTftq36y_nzzGAXciIKs-coY7ZzSpcY-PoMvXZeKoNp3rVZzZ0cD2kG-Mn728ooeoY1D8LUFHdC1LVLZQsrFzarVdC8vQSmqkP1QULrf6q41RLtXAKXPDWGTtxejFGyQDTupx8eCumjIN36sAD0VGgNP24W-wRl-3YbKDug"

	result, err := GetAssessmentFeedbackModel().IsCompleteByScheduleIDs(ctx, op, []string{"6099c4fee05f6e94002738a3"})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestAssessmentOfflineStudyModel_GetUserResult(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "92db7ddd-1f23-4f64-bd47-94f6d34a50c0",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE3YTI4MzM4LTNiODgtNGJhYy1hYjE1LWNjZTM4ODdhZjM1NyIsImVtYWlsIjoib3JnbWlAeW9wbWFpbC5jb20iLCJleHAiOjE2Mzk5ODMwMTAsImlzcyI6ImtpZHNsb29wIn0.Pd8hG7ChoH3Gduxx8o16PkupaqxDtanc82-oCQv9mntuVhm6XRyNrC-6pt7TRLyKYPMLBBKryBtDLxeItggENWgvA_V7Z_0puRo1MTg14nRxx72uQ2TAdnB08IVrCh-Qc5OtsdfnGjIzKCIEaSbGeeGdyaF3rHDg2RgeCAAmbli7APrpTftq36y_nzzGAXciIKs-coY7ZzSpcY-PoMvXZeKoNp3rVZzZ0cD2kG-Mn728ooeoY1D8LUFHdC1LVLZQsrFzarVdC8vQSmqkP1QULrf6q41RLtXAKXPDWGTtxejFGyQDTupx8eCumjIN36sAD0VGgNP24W-wRl-3YbKDug"

	result, err := GetAssessmentFeedbackModel().GetUserResult(ctx, op, []string{"60a5b97c03b03c3acdb5b589"}, []string{"79d78876-79bb-4b79-9868-4a99e03ca757"})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}
