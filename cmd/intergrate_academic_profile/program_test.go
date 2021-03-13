package intergrate_academic_profile

import (
	"context"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

func TestMapperImpl_Program(t *testing.T) {
	ids := []string{
		"5fd9ddface9660cbc5f667d8",
		"5fdac06ea878718a554ff00d",
		"5fdac0f61f066722a1351adb",
		"5fdac0fe1f066722a1351ade",
		"program0",
		"program1",
		"program2",
		"program3",
		"program4",
		"program5",
		"program6",
		"program7",
	}
	want := []string{
		"7565ae11-8130-4b7d-ac24-1d9dd6f792f2", // None Specified
		"75004121-0c0d-486c-ba65-4c57deacb44b", // ESL
		"14d350f1-a7ba-4f46-bef9-dc847f0cbac5", // Math
		"04c630cc-fabe-4176-80f2-30a029907a33", // Science
		"7565ae11-8130-4b7d-ac24-1d9dd6f792f2", // None Specified *
		"f6617737-5022-478d-9672-0354667e0338", // Bada Talk
		"4591423a-2619-4ef8-a900-f5d924939d02", // Bada Math
		"d1bbdcc5-0d80-46b0-b98e-162e7439058f", // Bada STEM
		"b39edb9a-ab91-4245-94a4-eb2b5007c033", // Bada Genius
		"7a8c5021-142b-44b1-b60b-275c29d132fe", // Bada Read
		"56e24fa0-e139-4c80-b365-61c9bc42cd3f", // Bada Sound
		"93f293e8-2c6a-47ad-bc46-1554caac99e4", // Bada Rhyme
	}

	ctx := context.Background()
	for index, id := range ids {
		got, err := testMapper.Program(ctx, testOperator.OrgID, id)
		if err == constant.ErrRecordNotFound {
			t.Errorf("program not found, id: %s", id)
			continue
		}

		if err != nil {
			t.Errorf("MapperImpl.Program() error = %v", err)
			return
		}

		if got != want[index] {
			t.Errorf("MapperImpl.Program() = %v, want %v", got, want[index])
		}
	}
}
