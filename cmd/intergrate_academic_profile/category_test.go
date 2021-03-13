package intergrate_academic_profile

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type categoryTestCase struct {
	ProgramID     string
	CategoryID    string
	AmsCategoryID string
}

func TestMapperImpl_Category(t *testing.T) {
	tests := []categoryTestCase{
		{
			// Speech & Language Skills
			ProgramID:     "5fdac0f61f066722a1351adb",
			CategoryID:    "developmental1",
			AmsCategoryID: "1080d319-8ce7-4378-9c71-a5019d6b9386",
		},
		{
			// Fine Motor Skills
			ProgramID:     "5fdac0f61f066722a1351adb",
			CategoryID:    "developmental2",
			AmsCategoryID: "f9d82bdd-4ee2-49dd-a707-133407cdef19",
		},
		{
			// Gross Motor Skills
			ProgramID:     "5fdac0f61f066722a1351adb",
			CategoryID:    "developmental3",
			AmsCategoryID: "a1c26321-e3a7-4ff2-9f1c-bb1c5e420fb7",
		},
		{
			// Cognitive Skills
			ProgramID:     "5fdac0f61f066722a1351adb",
			CategoryID:    "developmental4",
			AmsCategoryID: "c12f363a-633b-4080-bd2b-9ced8d034379",
		},
		{
			// Personal Development
			ProgramID:     "5fdac0f61f066722a1351adb",
			CategoryID:    "developmental5",
			AmsCategoryID: "e06ad483-085c-4869-bd88-56d17c7810a0",
		},
	}

	ctx := context.Background()

	for _, test := range tests {
		got, err := testMapper.Category(ctx, testOperator.OrgID, test.ProgramID, test.CategoryID)
		if err == constant.ErrRecordNotFound {
			t.Errorf("category not found, pid: %s sid: %s", test.ProgramID, test.CategoryID)
			continue
		}

		if err != nil {
			t.Errorf("MapperImpl.Category() error = %v", err)
			return
		}

		if got != test.AmsCategoryID {
			t.Errorf("MapperImpl.Category() = %v, want %v", got, test.AmsCategoryID)
		} else {
			t.Logf("MapperImpl.Category(),got:%v,  categoryID:%v, amsID:%v", got, test.CategoryID, test.AmsCategoryID)
		}
	}
}
