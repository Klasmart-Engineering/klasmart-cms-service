package intergrate_academic_profile

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type subCategoryTestCase struct {
	ProgramID        string
	CategoryID       string
	SubCategoryID    string
	AmsSubCategoryID string
}

func TestMapperImpl_SubCategory(t *testing.T) {
	tests := []subCategoryTestCase{
		{
			// Numbers
			ProgramID:        "5fdac0f61f066722a1351adb",
			CategoryID:       "developmental1",
			SubCategoryID:    "skills20",
			AmsSubCategoryID: "8d49bbbb-b230-4d5a-900b-cde6283519a3",
		},
		{
			// Vocabulary
			ProgramID:        "5fdac0f61f066722a1351adb",
			CategoryID:       "developmental1",
			SubCategoryID:    "skills3",
			AmsSubCategoryID: "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
		},
		{
			// Sensory
			ProgramID:        "5fdac0f61f066722a1351adb",
			CategoryID:       "developmental2",
			SubCategoryID:    "skills7",
			AmsSubCategoryID: "963729a4-7853-49d2-b75d-2c61d291afee",
		},
		{
			// Complex Movements
			ProgramID:        "5fdac0f61f066722a1351adb",
			CategoryID:       "developmental3",
			SubCategoryID:    "skills10",
			AmsSubCategoryID: "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7",
		},
		{
			// Numbers
			ProgramID:        "5fdac0f61f066722a1351adb",
			CategoryID:       "developmental4",
			SubCategoryID:    "skills20",
			AmsSubCategoryID: "8d49bbbb-b230-4d5a-900b-cde6283519a3",
		},
	}

	ctx := context.Background()

	for _, test := range tests {
		got, err := testMapper.SubCategory(ctx, testOperator.OrgID, test.ProgramID, test.CategoryID, test.SubCategoryID)
		if err == constant.ErrRecordNotFound {
			t.Errorf("subCategory not found, pid: %s sid: %s", test.ProgramID, test.SubCategoryID)
			continue
		}

		if err != nil {
			t.Errorf("MapperImpl.SubCategory() error = %v", err)
			return
		}

		if got != test.AmsSubCategoryID {
			t.Errorf("MapperImpl.SubCategory() = %v, want %v", got, test.AmsSubCategoryID)
		} else {
			t.Logf("MapperImpl.SubCategory(),got:%v,  subCategoryID:%v, amsID:%v", got, test.SubCategoryID, test.AmsSubCategoryID)
		}
	}
}
