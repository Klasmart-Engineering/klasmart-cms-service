package intergrate_academic_profile

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type subjectTestCase struct {
	ProgramID    string
	SubjectID    string
	AmsSubjectID string
}

func TestMapperImpl_Subject(t *testing.T) {
	tests := []subjectTestCase{
		{
			// Bada Genius
			ProgramID:    "program4",
			SubjectID:    "subject1",
			AmsSubjectID: "66a453b0-d38f-472e-b055-7a94a94d66c4",
		},
		{
			// Bada Math
			ProgramID:    "program2",
			SubjectID:    "subject2",
			AmsSubjectID: "36c4f793-9aa3-4fb8-84f0-68a2ab920d5a",
		},
		{
			// Bada Read
			ProgramID:    "program5",
			SubjectID:    "subject1",
			AmsSubjectID: "b997e0d1-2dd7-40d8-847a-b8670247e96b",
		},
		{
			// Bada Rhyme
			ProgramID:    "program7",
			SubjectID:    "subject1",
			AmsSubjectID: "49c8d5ee-472b-47a6-8c57-58daf863c2e1",
		},
		{
			// Bada Sound
			ProgramID:    "program6",
			SubjectID:    "subject1",
			AmsSubjectID: "b19f511e-a46b-488d-9212-22c0369c8afd",
		},
		{
			// Bada STEM
			ProgramID:    "program3",
			SubjectID:    "subject3",
			AmsSubjectID: "29d24801-0089-4b8e-85d3-77688e961efb",
		},
		{
			// Bada Talk
			ProgramID:    "program1",
			SubjectID:    "subject1",
			AmsSubjectID: "f037ee92-212c-4592-a171-ed32fb892162",
		},
		{
			// Default
			ProgramID:    "5fd9ddface9660cbc5f667d8",
			SubjectID:    "subject0",
			AmsSubjectID: "5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71",
		},
		{
			// ESL
			ProgramID:    "5fdac06ea878718a554ff00d",
			SubjectID:    "subject1",
			AmsSubjectID: "20d6ca2f-13df-4a7a-8dcb-955908db7baa",
		},
		{
			// Math
			ProgramID:    "5fdac0f61f066722a1351adb",
			SubjectID:    "subject2",
			AmsSubjectID: "7cf8d3a3-5493-46c9-93eb-12f220d101d0",
		},
		{
			// Science
			ProgramID:    "5fdac0fe1f066722a1351ade",
			SubjectID:    "subject3",
			AmsSubjectID: "fab745e8-9e31-4d0c-b780-c40120c98b27",
		},
	}

	ctx := context.Background()
	for _, test := range tests {
		got, err := testMapper.Subject(ctx, testOperator.OrgID, test.ProgramID, test.SubjectID)
		if err == constant.ErrRecordNotFound {
			t.Errorf("subject not found, pid: %s sid: %s", test.ProgramID, test.SubjectID)
			continue
		}

		if err != nil {
			t.Errorf("MapperImpl.Subject() error = %v", err)
			return
		}

		if got != test.AmsSubjectID {
			t.Errorf("MapperImpl.Subject() = %v, want %v", got, test.AmsSubjectID)
		}
	}
}
