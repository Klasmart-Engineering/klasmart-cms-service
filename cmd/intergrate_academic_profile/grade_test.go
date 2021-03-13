package intergrate_academic_profile

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type gradeTestCase struct {
	ProgramID  string
	GradeID    string
	AmsGradeID string
}

func TestMapperImpl_Grade(t *testing.T) {
	tests := []gradeTestCase{
		{
			// PreK-3
			ProgramID:  "5fdac0f61f066722a1351adb",
			GradeID:    "grade7",
			AmsGradeID: "b20eaf10-3e40-4ef7-9d74-93a13782d38f",
		},
		{
			// PreK-4
			ProgramID:  "5fdac0f61f066722a1351adb",
			GradeID:    "grade8",
			AmsGradeID: "89d71050-186e-4fb2-8cbd-9598ca312be9",
		},
		{
			// PreK-5
			ProgramID:  "5fdac0f61f066722a1351adb",
			GradeID:    "grade9",
			AmsGradeID: "abc900b9-5b8c-4e54-a4a8-54f102b2c1c6",
		},
		{
			// PreK-6
			ProgramID:  "5fdac0f61f066722a1351adb",
			GradeID:    "grade10",
			AmsGradeID: "3ee3fd4c-6208-494f-9551-d48fabc4f42a",
		},
		{
			// PreK-7
			ProgramID:  "5fdac0f61f066722a1351adb",
			GradeID:    "grade11",
			AmsGradeID: "781e8a08-29e8-4171-8392-7e8ac9f183a0",
		},
	}

	ctx := context.Background()

	for _, test := range tests {
		got, err := testMapper.Grade(ctx, testOperator.OrgID, test.ProgramID, test.GradeID)
		if err == constant.ErrRecordNotFound {
			t.Errorf("grade not found, pid: %s sid: %s", test.ProgramID, test.GradeID)
			continue
		}

		if err != nil {
			t.Errorf("MapperImpl.Grade() error = %v", err)
			return
		}

		if got != test.AmsGradeID {
			t.Errorf("MapperImpl.Grade() = %v, want %v", got, test.AmsGradeID)
		} else {
			t.Logf("MapperImpl.Grade(),got:%v,  gradeID:%v, amsID:%v", got, test.GradeID, test.AmsGradeID)
		}
	}
}
