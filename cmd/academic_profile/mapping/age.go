package mapping

import "context"

func (s mapperImpl) Age(ctx context.Context, organizationID, programID, AgeID string) string {
	programID = s.Program(ctx, organizationID, programID)
	switch programID {
	case "75004121-0c0d-486c-ba65-4c57deacb44b":
		// ESL
		switch AgeID {
		case "7965d220-619d-400f-8cab-42bd98c7d23c",
			"age1":
			// 3 - 4 year(s)
			return "7965d220-619d-400f-8cab-42bd98c7d23c"
		case "bb7982cd-020f-4e1a-93fc-4a6874917f07",
			"age2":
			// 4 - 5 year(s)
			return "bb7982cd-020f-4e1a-93fc-4a6874917f07"
		case "fe0b81a4-5b02-4548-8fb0-d49cd4a4604a",
			"age3":
			// 5 - 6 year(s)
			return "fe0b81a4-5b02-4548-8fb0-d49cd4a4604a"
		case "145edddc-2019-43d9-97e1-c5830e7ed689",
			"age4":
			// 6 - 7 year(s)
			return "145edddc-2019-43d9-97e1-c5830e7ed689"
		default:
			// 3 - 4 year(s)
			return "7965d220-619d-400f-8cab-42bd98c7d23c"
		}
	case "14d350f1-a7ba-4f46-bef9-dc847f0cbac5", // Math
		"04c630cc-fabe-4176-80f2-30a029907a33", // Science
		"4591423a-2619-4ef8-a900-f5d924939d02", // Bada Math
		"d1bbdcc5-0d80-46b0-b98e-162e7439058f": // Bada STEM
		switch AgeID {
		case "7965d220-619d-400f-8cab-42bd98c7d23c",
			"age1":
			// 3 - 4 year(s)
			return "7965d220-619d-400f-8cab-42bd98c7d23c"
		case "bb7982cd-020f-4e1a-93fc-4a6874917f07",
			"age2":
			// 4 - 5 year(s)
			return "bb7982cd-020f-4e1a-93fc-4a6874917f07"
		case "fe0b81a4-5b02-4548-8fb0-d49cd4a4604a",
			"age3":
			// 5 - 6 year(s)
			return "fe0b81a4-5b02-4548-8fb0-d49cd4a4604a"
		case "145edddc-2019-43d9-97e1-c5830e7ed689",
			"age4":
			// 6 - 7 year(s)
			return "145edddc-2019-43d9-97e1-c5830e7ed689"
		case "21f1da64-b6c8-4e74-9fef-09d08cfd8e6c",
			"age5":
			// 7 - 8 year(s)
			return "21f1da64-b6c8-4e74-9fef-09d08cfd8e6c"
		default:
			// 3 - 4 year(s)
			return "7965d220-619d-400f-8cab-42bd98c7d23c"
		}
	case "f6617737-5022-478d-9672-0354667e0338", // Bada Talk
		"b39edb9a-ab91-4245-94a4-eb2b5007c033", // Bada Genius
		"93f293e8-2c6a-47ad-bc46-1554caac99e4": // Bada Rhyme
		switch AgeID {
		case "7965d220-619d-400f-8cab-42bd98c7d23c",
			"age1":
			// 3 - 4 year(s)
			return "7965d220-619d-400f-8cab-42bd98c7d23c"
		case "bb7982cd-020f-4e1a-93fc-4a6874917f07",
			"age2":
			// 4 - 5 year(s)
			return "bb7982cd-020f-4e1a-93fc-4a6874917f07"
		case "fe0b81a4-5b02-4548-8fb0-d49cd4a4604a",
			"age3":
			// 5 - 6 year(s)
			return "fe0b81a4-5b02-4548-8fb0-d49cd4a4604a"
		default:
			// 3 - 4 year(s)
			return "7965d220-619d-400f-8cab-42bd98c7d23c"
		}
	case "7a8c5021-142b-44b1-b60b-275c29d132fe", // Bada Read
		"56e24fa0-e139-4c80-b365-61c9bc42cd3f": // Bada Sound
		switch AgeID {
		case "7965d220-619d-400f-8cab-42bd98c7d23c",
			"age1":
			// 3 - 4 year(s)
			return "7965d220-619d-400f-8cab-42bd98c7d23c"
		case "bb7982cd-020f-4e1a-93fc-4a6874917f07",
			"age2":
			// 4 - 5 year(s)
			return "bb7982cd-020f-4e1a-93fc-4a6874917f07"
		case "fe0b81a4-5b02-4548-8fb0-d49cd4a4604a",
			"age3":
			// 5 - 6 year(s)
			return "fe0b81a4-5b02-4548-8fb0-d49cd4a4604a"
		case "21f1da64-b6c8-4e74-9fef-09d08cfd8e6c",
			"age5":
			// 7 - 8 year(s)
			return "21f1da64-b6c8-4e74-9fef-09d08cfd8e6c"
		default:
			// 3 - 4 year(s)
			return "7965d220-619d-400f-8cab-42bd98c7d23c"
		}
	case "7565ae11-8130-4b7d-ac24-1d9dd6f792f2",
		"age0":
		// None Specified
		return "023eeeb1-5f72-4fa3-a2a7-63603607ac2b"
	default:
		// None Specified
		return "023eeeb1-5f72-4fa3-a2a7-63603607ac2b"
	}
}
