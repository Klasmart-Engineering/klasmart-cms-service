package mapping

import "context"

func (s mapperImpl) Grade(ctx context.Context, organizationID, programID, gradeID string) string {
	programID = s.Program(ctx, organizationID, programID)
	switch programID {
	case "75004121-0c0d-486c-ba65-4c57deacb44b":
		// ESL
		switch gradeID {
		case "0ecb8fa9-d77e-4dd3-b220-7e79704f1b03",
			"grade2":
			// PreK-1
			return "0ecb8fa9-d77e-4dd3-b220-7e79704f1b03"
		case "66fcda51-33c8-4162-a8d1-0337e1d6ade3",
			"grade3":
			// PreK-2
			return "66fcda51-33c8-4162-a8d1-0337e1d6ade3"
		case "a9f0217d-f7ec-4add-950d-4e8986ab2c82",
			"grade12":
			// Kindergarten
			return "a9f0217d-f7ec-4add-950d-4e8986ab2c82"
		case "e4d16af5-5b8f-4051-b065-13acf6c694be",
			"grade5":
			// Grade 1
			return "e4d16af5-5b8f-4051-b065-13acf6c694be"
		default:
			// PreK-1
			return "0ecb8fa9-d77e-4dd3-b220-7e79704f1b03"
		}
	case "14d350f1-a7ba-4f46-bef9-dc847f0cbac5",
		"04c630cc-fabe-4176-80f2-30a029907a33":
		// Math
		// Science
		switch gradeID {
		case "b20eaf10-3e40-4ef7-9d74-93a13782d38f",
			"grade7":
			// PreK-3
			return "b20eaf10-3e40-4ef7-9d74-93a13782d38f"
		case "89d71050-186e-4fb2-8cbd-9598ca312be9",
			"grade8":
			// PreK-4
			return "89d71050-186e-4fb2-8cbd-9598ca312be9"
		case "abc900b9-5b8c-4e54-a4a8-54f102b2c1c6",
			"grade9":
			// PreK-5
			return "abc900b9-5b8c-4e54-a4a8-54f102b2c1c6"
		case "3ee3fd4c-6208-494f-9551-d48fabc4f42a",
			"grade10":
			// PreK-6
			return "3ee3fd4c-6208-494f-9551-d48fabc4f42a"
		case "781e8a08-29e8-4171-8392-7e8ac9f183a0",
			"grade11":
			// PreK-7
			return "781e8a08-29e8-4171-8392-7e8ac9f183a0"
		default:
			// PreK-3
			return "b20eaf10-3e40-4ef7-9d74-93a13782d38f"
		}
	case "4591423a-2619-4ef8-a900-f5d924939d02",
		"d1bbdcc5-0d80-46b0-b98e-162e7439058f":
		// Bada Math
		// Bada STEM
		switch gradeID {
		case "d7e2e258-d4b3-4e95-b929-49ae702de4be",
			"grade2":
			// PreK-1
			return "d7e2e258-d4b3-4e95-b929-49ae702de4be"
		case "3e7979f6-7375-450a-9818-ddb09b250bb2",
			"grade3":
			// PreK-2
			return "3e7979f6-7375-450a-9818-ddb09b250bb2"
		case "81dcbcc6-3d70-4bdf-99bc-14833c57c628",
			"grade4":
			// K
			return "81dcbcc6-3d70-4bdf-99bc-14833c57c628"
		case "100f774a-3d7e-4be5-9c2c-ae70f40f0b50",
			"grade5":
			// Grade 1
			return "100f774a-3d7e-4be5-9c2c-ae70f40f0b50"
		case "9d3e591d-06a6-4fc4-9714-cf155a15b415",
			"grade6":
			// Grade 2
			return "9d3e591d-06a6-4fc4-9714-cf155a15b415"
		default:
			// PreK-1
			return "d7e2e258-d4b3-4e95-b929-49ae702de4be"
		}
	case "f6617737-5022-478d-9672-0354667e0338", // Bada Talk
		"b39edb9a-ab91-4245-94a4-eb2b5007c033", // Bada Genius
		"7a8c5021-142b-44b1-b60b-275c29d132fe", // Bada Read
		"56e24fa0-e139-4c80-b365-61c9bc42cd3f", // Bada Sound
		"93f293e8-2c6a-47ad-bc46-1554caac99e4", // Bada Rhyme
		"7565ae11-8130-4b7d-ac24-1d9dd6f792f2", // None Specified
		"grade0",                               // None Specified
		"grade1":                               // Not Specific
		// None Specified
		return "98461ca1-06a1-432a-97d0-4e1dff33e1a5"
	default:
		// None Specified
		return "98461ca1-06a1-432a-97d0-4e1dff33e1a5"
	}
}
