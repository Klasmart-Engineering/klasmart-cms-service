package intergrate_academic_profile

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type ageTestCase struct {
	ProgramID string
	AgeID     string
	AmsAgeID  string
}

func TestMapperImpl_Age(t *testing.T) {
	tests := []ageTestCase{
		{
			// none program
			ProgramID: "program0",
			AgeID:     "age1",
			AmsAgeID:  "023eeeb1-5f72-4fa3-a2a7-63603607ac2b",
		},
		//{
		//	// none age
		//	ProgramID: "program7",
		//	AgeID:     "age5",
		//	AmsAgeID:  "023eeeb1-5f72-4fa3-a2a7-63603607ac2b",
		//},
		//{
		//	// 3-4
		//	ProgramID: "5fdac0f61f066722a1351adb",
		//	AgeID:     "age1",
		//	AmsAgeID:  "7965d220-619d-400f-8cab-42bd98c7d23c",
		//},
		//{
		//	// 4-5
		//	ProgramID: "5fdac0f61f066722a1351adb",
		//	AgeID:     "age2",
		//	AmsAgeID:  "bb7982cd-020f-4e1a-93fc-4a6874917f07",
		//},
		//{
		//	// 5-6
		//	ProgramID: "5fdac0f61f066722a1351adb",
		//	AgeID:     "age3",
		//	AmsAgeID:  "fe0b81a4-5b02-4548-8fb0-d49cd4a4604a",
		//},
		//{
		//	// 6-7
		//	ProgramID: "5fdac0f61f066722a1351adb",
		//	AgeID:     "age4",
		//	AmsAgeID:  "145edddc-2019-43d9-97e1-c5830e7ed689",
		//},
		//{
		//	// 7-8
		//	ProgramID: "5fdac0f61f066722a1351adb",
		//	AgeID:     "age5",
		//	AmsAgeID:  "21f1da64-b6c8-4e74-9fef-09d08cfd8e6c",
		//},
	}

	ctx := context.Background()

	for _, test := range tests {
		got, err := testMapper.Age(ctx, testOperator.OrgID, test.ProgramID, test.AgeID)
		if err == constant.ErrRecordNotFound {
			t.Errorf("age not found, pid: %s sid: %s", test.ProgramID, test.AgeID)
			continue
		}

		if err != nil {
			t.Errorf("MapperImpl.Age() error = %v", err)
			return
		}

		if got != test.AmsAgeID {
			t.Errorf("MapperImpl.Age() = %v, want %v", got, test.AmsAgeID)
		} else {
			t.Logf("MapperImpl.Age(),got:%v,  ageID:%v, amsID:%v", got, test.AgeID, test.AmsAgeID)
		}
	}
}
