package model

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

func TestAssessmentOfflineStudyModel_Page(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImM1N2VmNjhkLWE2MzUtNDUxZC1iOTk3LWFlYmMzYzI5Yjk5YSIsImVtYWlsIjoib3JnYmFkYUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0MDY4MjI0MSwiaXNzIjoia2lkc2xvb3AifQ.YGGvnQBHLQxe_7LldZ5LdfxyswGdVvWY9KNIbt5dtfvPmm6n7PQD3TB4n3ZzRqrQAhLVwdHpHGeEQ7J2k2HyMGMPuOQY_2zmWIH4eguHe2SjNd1_g5iGft9AcK5G1Wna3SYIHFtyc7x8zWZQn2hv7DTxyUUmeTqHQBmtHW6j2BUwSCcS1N2oBcYgOGI7BLZyA2KRo7NTxkDaKhRW-tjvILbQk40MPaqhpnaLTgHg33w5T7qSLDYtPxfQjXrzWZ5Orrljl8NyjKzZMPWtv893ryRUCyAiL3rPMRVUjCqRB9h07puDJ5V1vg1FhcYs_yCNQXPeHOcA8_HUnaEBsqTuAg"
	t.Log(op)
	result, err := GetAssessmentOfflineStudyModel().Page(ctx, op, &v2.AssessmentQueryReq{
		QueryKey:  "",
		QueryType: "",
		//AssessmentType: v2.AssessmentTypeOnlineStudy,
		OrderBy:   "-create_at",
		Status:    "NotStared,Started,Draft,Complete", //"Complete",
		PageIndex: 3,
		PageSize:  3,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

//class m-830830-1
func TestAssessmentOfflineStudyModel_GetByID(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE3YTI4MzM4LTNiODgtNGJhYy1hYjE1LWNjZTM4ODdhZjM1NyIsImVtYWlsIjoib3JnbWlAeW9wbWFpbC5jb20iLCJleHAiOjE2Mzk5ODMwMTAsImlzcyI6ImtpZHNsb29wIn0.Pd8hG7ChoH3Gduxx8o16PkupaqxDtanc82-oCQv9mntuVhm6XRyNrC-6pt7TRLyKYPMLBBKryBtDLxeItggENWgvA_V7Z_0puRo1MTg14nRxx72uQ2TAdnB08IVrCh-Qc5OtsdfnGjIzKCIEaSbGeeGdyaF3rHDg2RgeCAAmbli7APrpTftq36y_nzzGAXciIKs-coY7ZzSpcY-PoMvXZeKoNp3rVZzZ0cD2kG-Mn728ooeoY1D8LUFHdC1LVLZQsrFzarVdC8vQSmqkP1QULrf6q41RLtXAKXPDWGTtxejFGyQDTupx8eCumjIN36sAD0VGgNP24W-wRl-3YbKDug"

	result, err := GetAssessmentOfflineStudyModel().GetByID(ctx, op, "60f7c2d618f2487344d9d46b")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestAssessmentOfflineStudyModel_IsAnyOneCompleteByScheduleIDs(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE3YTI4MzM4LTNiODgtNGJhYy1hYjE1LWNjZTM4ODdhZjM1NyIsImVtYWlsIjoib3JnbWlAeW9wbWFpbC5jb20iLCJleHAiOjE2Mzk5ODMwMTAsImlzcyI6ImtpZHNsb29wIn0.Pd8hG7ChoH3Gduxx8o16PkupaqxDtanc82-oCQv9mntuVhm6XRyNrC-6pt7TRLyKYPMLBBKryBtDLxeItggENWgvA_V7Z_0puRo1MTg14nRxx72uQ2TAdnB08IVrCh-Qc5OtsdfnGjIzKCIEaSbGeeGdyaF3rHDg2RgeCAAmbli7APrpTftq36y_nzzGAXciIKs-coY7ZzSpcY-PoMvXZeKoNp3rVZzZ0cD2kG-Mn728ooeoY1D8LUFHdC1LVLZQsrFzarVdC8vQSmqkP1QULrf6q41RLtXAKXPDWGTtxejFGyQDTupx8eCumjIN36sAD0VGgNP24W-wRl-3YbKDug"

	result, err := GetAssessmentOfflineStudyModel().IsAnyOneCompleteByScheduleIDs(ctx, op, []string{"6099c4fee05f6e94002738a3"})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestAssessmentOfflineStudyModel_GetUserResult(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE3YTI4MzM4LTNiODgtNGJhYy1hYjE1LWNjZTM4ODdhZjM1NyIsImVtYWlsIjoib3JnbWlAeW9wbWFpbC5jb20iLCJleHAiOjE2Mzk5ODMwMTAsImlzcyI6ImtpZHNsb29wIn0.Pd8hG7ChoH3Gduxx8o16PkupaqxDtanc82-oCQv9mntuVhm6XRyNrC-6pt7TRLyKYPMLBBKryBtDLxeItggENWgvA_V7Z_0puRo1MTg14nRxx72uQ2TAdnB08IVrCh-Qc5OtsdfnGjIzKCIEaSbGeeGdyaF3rHDg2RgeCAAmbli7APrpTftq36y_nzzGAXciIKs-coY7ZzSpcY-PoMvXZeKoNp3rVZzZ0cD2kG-Mn728ooeoY1D8LUFHdC1LVLZQsrFzarVdC8vQSmqkP1QULrf6q41RLtXAKXPDWGTtxejFGyQDTupx8eCumjIN36sAD0VGgNP24W-wRl-3YbKDug"

	result, err := GetAssessmentOfflineStudyModel().GetUserResult(ctx, op, []string{"612ca6a6c9855a9b80e2b23d"}, []string{"79d78876-79bb-4b79-9868-4a99e03ca757"})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}
