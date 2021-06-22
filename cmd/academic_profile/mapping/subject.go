package mapping

import "context"

func (s mapperImpl) Subject(ctx context.Context, organizationID, programID, subjectID string) string {
	programID = s.Program(ctx, organizationID, programID)
	switch programID {
	case "75004121-0c0d-486c-ba65-4c57deacb44b":
		// ESL
		// Language/Literacy
		return "20d6ca2f-13df-4a7a-8dcb-955908db7baa"
	case "14d350f1-a7ba-4f46-bef9-dc847f0cbac5":
		// Math
		// Math
		return "7cf8d3a3-5493-46c9-93eb-12f220d101d0"
	case "04c630cc-fabe-4176-80f2-30a029907a33":
		// Science
		// Science
		return "fab745e8-9e31-4d0c-b780-c40120c98b27"
	case "f6617737-5022-478d-9672-0354667e0338":
		// Bada Talk
		// Language/Literacy
		return "f037ee92-212c-4592-a171-ed32fb892162"
	case "4591423a-2619-4ef8-a900-f5d924939d02":
		// Bada Math
		// Math
		return "36c4f793-9aa3-4fb8-84f0-68a2ab920d5a"
	case "d1bbdcc5-0d80-46b0-b98e-162e7439058f":
		// Bada STEM
		// Science
		return "29d24801-0089-4b8e-85d3-77688e961efb"
	case "b39edb9a-ab91-4245-94a4-eb2b5007c033":
		// Bada Genius
		// Language/Literacy
		return "66a453b0-d38f-472e-b055-7a94a94d66c4"
	case "7a8c5021-142b-44b1-b60b-275c29d132fe":
		// Bada Read
		// Language/Literacy
		return "b997e0d1-2dd7-40d8-847a-b8670247e96b"
	case "56e24fa0-e139-4c80-b365-61c9bc42cd3f":
		// Bada Sound
		// Language/Literacy
		return "b19f511e-a46b-488d-9212-22c0369c8afd"
	case "93f293e8-2c6a-47ad-bc46-1554caac99e4":
		// Bada Rhyme
		// Language/Literacy
		return "49c8d5ee-472b-47a6-8c57-58daf863c2e1"
	case "7565ae11-8130-4b7d-ac24-1d9dd6f792f2":
		// None Specified
		return "5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71"
	default:
		// None Specified
		return "5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71"
	}
}
