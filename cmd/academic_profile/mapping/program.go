package mapping

import "context"

func (s *mapperImpl) Program(ctx context.Context, organizationID, programID string) string {
	switch programID {
	case "program4",
		"b39edb9a-ab91-4245-94a4-eb2b5007c033":
		// Bada Genius
		return "b39edb9a-ab91-4245-94a4-eb2b5007c033"
	case "program2",
		"4591423a-2619-4ef8-a900-f5d924939d02":
		// Bada Math
		return "4591423a-2619-4ef8-a900-f5d924939d02"
	case "program5",
		"7a8c5021-142b-44b1-b60b-275c29d132fe":
		// Bada Read
		return "7a8c5021-142b-44b1-b60b-275c29d132fe"
	case "program7",
		"93f293e8-2c6a-47ad-bc46-1554caac99e4":
		// Bada Rhyme
		return "93f293e8-2c6a-47ad-bc46-1554caac99e4"
	case "program6",
		"56e24fa0-e139-4c80-b365-61c9bc42cd3f":
		// Bada Sound
		return "56e24fa0-e139-4c80-b365-61c9bc42cd3f"
	case "program3",
		"d1bbdcc5-0d80-46b0-b98e-162e7439058f":
		// Bada STEM
		return "d1bbdcc5-0d80-46b0-b98e-162e7439058f"
	case "program1",
		"f6617737-5022-478d-9672-0354667e0338":
		// Bada Talk
		return "f6617737-5022-478d-9672-0354667e0338"
	case "5fdac06ea878718a554ff00d",
		"75004121-0c0d-486c-ba65-4c57deacb44b":
		// ESL
		return "75004121-0c0d-486c-ba65-4c57deacb44b"
	case "5fdac0f61f066722a1351adb",
		"14d350f1-a7ba-4f46-bef9-dc847f0cbac5":
		// Math
		return "14d350f1-a7ba-4f46-bef9-dc847f0cbac5"
	case "5fdac0fe1f066722a1351ade",
		"04c630cc-fabe-4176-80f2-30a029907a33":
		// Science
		return "04c630cc-fabe-4176-80f2-30a029907a33"
	case "program0",
		"5fd9ddface9660cbc5f667d8",
		"7565ae11-8130-4b7d-ac24-1d9dd6f792f2":
		// None Specified
		return "7565ae11-8130-4b7d-ac24-1d9dd6f792f2"
	default:
		// None Specified
		return "7565ae11-8130-4b7d-ac24-1d9dd6f792f2"
	}
}
