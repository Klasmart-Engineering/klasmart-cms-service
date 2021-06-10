package mapping

import "context"

func (s mapperImpl) Category(ctx context.Context, organizationID, programID, categoryID string) string {
	programID = s.Program(ctx, organizationID, programID)
	switch programID {
	case "75004121-0c0d-486c-ba65-4c57deacb44b":
		// ESL
		switch categoryID {
		case "84b8f87a-7b61-4580-a190-a9ce3fe90dd3",
			"developmental1":
			// Speech & Language Skills
			return "84b8f87a-7b61-4580-a190-a9ce3fe90dd3"
		case "ce9014a4-01a9-49d5-bf10-6b08bc454fc1",
			"developmental2":
			// Fine Motor Skills
			return "ce9014a4-01a9-49d5-bf10-6b08bc454fc1"
		case "61996d3d-a37d-4873-bcdc-03b22fc6977e",
			"developmental3":
			// Gross Motor Skills
			return "61996d3d-a37d-4873-bcdc-03b22fc6977e"
		case "e08f3578-a7d4-4cac-b028-ef7a8c93f53f",
			"developmental4":
			// Cognitive Skills
			return "e08f3578-a7d4-4cac-b028-ef7a8c93f53f"
		case "76cc6f90-86ef-48b7-9138-7b2f0bc378e7",
			"developmental5":
			// Personal Development
			return "76cc6f90-86ef-48b7-9138-7b2f0bc378e7"
		default:
			// Speech & Language Skills
			return "84b8f87a-7b61-4580-a190-a9ce3fe90dd3"
		}
	case "14d350f1-a7ba-4f46-bef9-dc847f0cbac5":
		// Math
		switch categoryID {
		case "1080d319-8ce7-4378-9c71-a5019d6b9386",
			"developmental1":
			// Speech & Language Skills
			return "1080d319-8ce7-4378-9c71-a5019d6b9386"
		case "f9d82bdd-4ee2-49dd-a707-133407cdef19",
			"developmental2":
			// Fine Motor Skills
			return "f9d82bdd-4ee2-49dd-a707-133407cdef19"
		case "a1c26321-e3a7-4ff2-9f1c-bb1c5e420fb7",
			"developmental3":
			// Gross Motor Skills
			return "a1c26321-e3a7-4ff2-9f1c-bb1c5e420fb7"
		case "c12f363a-633b-4080-bd2b-9ced8d034379",
			"developmental4":
			// Cognitive Skills
			return "c12f363a-633b-4080-bd2b-9ced8d034379"
		case "e06ad483-085c-4869-bd88-56d17c7810a0",
			"developmental5":
			// Personal Development
			return "e06ad483-085c-4869-bd88-56d17c7810a0"
		default:
			// Speech & Language Skills
			return "1080d319-8ce7-4378-9c71-a5019d6b9386"
		}
	case "04c630cc-fabe-4176-80f2-30a029907a33":
		// Science
		switch categoryID {
		case "1cc44ecc-153a-47e9-b6e8-3b1ef94a9dee",
			"developmental1":
			// Speech & Language Skills
			return "1cc44ecc-153a-47e9-b6e8-3b1ef94a9dee"
		case "0523610d-cf11-47b6-b7ab-bdbf8c3e09b6",
			"developmental2":
			// Fine Motor Skills
			return "0523610d-cf11-47b6-b7ab-bdbf8c3e09b6"
		case "d1783a8c-6bcd-492a-ad17-37386df80c56",
			"developmental3":
			// Gross Motor Skills
			return "d1783a8c-6bcd-492a-ad17-37386df80c56"
		case "1ef6ca6c-fbc4-4422-a5cb-2bcd999e4b2b",
			"developmental4":
			// Cognitive Skills
			return "1ef6ca6c-fbc4-4422-a5cb-2bcd999e4b2b"
		case "8488eeac-28bd-4f86-8093-9853b19f51db",
			"developmental5":
			// Personal Development
			return "8488eeac-28bd-4f86-8093-9853b19f51db"
		default:
			// Speech & Language Skills
			return "1cc44ecc-153a-47e9-b6e8-3b1ef94a9dee"
		}
	case "f6617737-5022-478d-9672-0354667e0338":
		// Bada Talk
		switch categoryID {
		case "1bb26398-3e38-441e-9a8a-460057f2d8c0",
			"developmental1":
			// Speech & Language Skills
			return "1bb26398-3e38-441e-9a8a-460057f2d8c0"
		case "e65ea6b4-7093-490a-927e-d2235643f6ca",
			"developmental2":
			// Fine Motor Skills
			return "e65ea6b4-7093-490a-927e-d2235643f6ca"
		case "88fff890-d614-4b88-be57-b7441fa40b66",
			"developmental3":
			// Gross Motor Skills
			return "88fff890-d614-4b88-be57-b7441fa40b66"
		case "b18d60c6-a545-46ff-8988-cd5d46ab9660",
			"developmental4":
			// Cognitive Skills
			return "b18d60c6-a545-46ff-8988-cd5d46ab9660"
		case "c83fd174-6504-4cc3-9175-2728d023c39d",
			"developmental5":
			// Personal Development
			return "c83fd174-6504-4cc3-9175-2728d023c39d"
		case "d17f1bee-cdef-4759-8c23-3e9b64d08ec1",
			"developmental9":
			// Oral Language
			return "d17f1bee-cdef-4759-8c23-3e9b64d08ec1"
		case "dd59f36d-717f-4982-9ae6-df32537faba0",
			"developmental10":
			// Literacy
			return "dd59f36d-717f-4982-9ae6-df32537faba0"
		case "8d464354-16d9-41af-b887-103f18f4b376",
			"developmental11":
			// Whole-Child
			return "8d464354-16d9-41af-b887-103f18f4b376"
		case "dfed32b5-f0bd-42ea-999b-e10b376038d5",
			"developmental12":
			// Knowledge
			return "dfed32b5-f0bd-42ea-999b-e10b376038d5"
		default:
			// Speech & Language Skills
			return "1bb26398-3e38-441e-9a8a-460057f2d8c0"
		}
	case "4591423a-2619-4ef8-a900-f5d924939d02":
		// Bada Math
		switch categoryID {
		case "2a637bea-c529-4868-8269-d0936696da7e",
			"developmental6":
			// Language and Numeracy Skills
			return "2a637bea-c529-4868-8269-d0936696da7e"
		case "6933de3e-a568-4d56-8c01-e110bda22926",
			"developmental2":
			// Fine Motor Skills
			return "6933de3e-a568-4d56-8c01-e110bda22926"
		case "3af9f093-4153-4348-a097-986c15d1f912",
			"developmental3":
			// Gross Motor Skills
			return "3af9f093-4153-4348-a097-986c15d1f912"
		case "a11a6f56-3ae3-4b70-8daa-30cdb63ef5b6",
			"developmental4":
			// Cognitive Skills
			return "a11a6f56-3ae3-4b70-8daa-30cdb63ef5b6"
		case "665616dd-32c2-44c4-91c9-63f7493c9fd3",
			"developmental8":
			// Social and Emotional
			return "665616dd-32c2-44c4-91c9-63f7493c9fd3"
		default:
			// Language and Numeracy Skills
			return "2a637bea-c529-4868-8269-d0936696da7e"
		}
	case "d1bbdcc5-0d80-46b0-b98e-162e7439058f":
		// Bada STEM
		switch categoryID {
		case "6090e473-ec19-4bf0-ae5c-2d6a4c793f55",
			"developmental1":
			// Speech & Language Skills
			return "6090e473-ec19-4bf0-ae5c-2d6a4c793f55"
		case "da9fa132-dcf7-4148-9037-b381850ba088",
			"developmental2":
			// Fine Motor Skills
			return "da9fa132-dcf7-4148-9037-b381850ba088"
		case "585f38e6-f7be-45f2-855a-f2a4bddca125",
			"developmental3":
			// Gross Motor Skills
			return "585f38e6-f7be-45f2-855a-f2a4bddca125"
		case "c3ea1b4a-d220-4248-9b3f-07559b415c56",
			"developmental4":
			// Cognitive Skills
			return "c3ea1b4a-d220-4248-9b3f-07559b415c56"
		case "7826ff58-25d0-41f1-b38e-3e3a77ed32f6",
			"developmental8":
			// Social and Emotional
			return "7826ff58-25d0-41f1-b38e-3e3a77ed32f6"
		default:
			// Speech & Language Skills
			return "6090e473-ec19-4bf0-ae5c-2d6a4c793f55"
		}
	case "b39edb9a-ab91-4245-94a4-eb2b5007c033":
		// Bada Genius
		switch categoryID {
		case "b8c76823-150d-4d83-861e-dce7d7bc4f6d",
			"developmental1":
			// Speech & Language Skills
			return "b8c76823-150d-4d83-861e-dce7d7bc4f6d"
		case "b4cd42b8-a09b-4f66-a03a-b9f6b6f69895",
			"developmental2":
			// Fine Motor Skills
			return "b4cd42b8-a09b-4f66-a03a-b9f6b6f69895"
		case "bcfd9d76-cf05-4ccd-9a41-6b886da661be",
			"developmental3":
			// Gross Motor Skills
			return "bcfd9d76-cf05-4ccd-9a41-6b886da661be"
		case "c909acad-0c52-4fd3-8427-3b1e90a730da",
			"developmental4":
			// Cognitive Skills
			return "c909acad-0c52-4fd3-8427-3b1e90a730da"
		case "fa8ff09d-9062-4955-9b20-5fa20757f1d9",
			"developmental5":
			// Personal Development
			return "fa8ff09d-9062-4955-9b20-5fa20757f1d9"
		case "29a0ab9e-6364-47b6-b63a-1388a7861c6c",
			"developmental9":
			// Oral Language
			return "29a0ab9e-6364-47b6-b63a-1388a7861c6c"
		case "49cbbf19-2ad7-4acb-b8c8-66531578116a",
			"developmental10":
			// Literacy
			return "49cbbf19-2ad7-4acb-b8c8-66531578116a"
		case "bd55fd6b-73ef-41ed-8a86-d7bbc501e773",
			"developmental11":
			// Whole-Child
			return "bd55fd6b-73ef-41ed-8a86-d7bbc501e773"
		case "dd3dbf0c-2872-433b-8b61-9ea78f3c9e97",
			"developmental12":
			// Knowledge
			return "dd3dbf0c-2872-433b-8b61-9ea78f3c9e97"
		default:
			// Speech & Language Skills
			return "b8c76823-150d-4d83-861e-dce7d7bc4f6d"
		}
	case "7a8c5021-142b-44b1-b60b-275c29d132fe":
		// Bada Read
		switch categoryID {
		case "64e000aa-4a2c-4e2e-9d8d-f779e97bdd73",
			"developmental1":
			// Speech & Language Skills
			return "64e000aa-4a2c-4e2e-9d8d-f779e97bdd73"
		case "59c47920-4d0d-477c-a33b-06e7f13873d7",
			"developmental2":
			// Fine Motor Skills
			return "59c47920-4d0d-477c-a33b-06e7f13873d7"
		case "7e887129-1e7d-40dc-8caa-5e1e0197fb4d",
			"developmental3":
			// Gross Motor Skills
			return "7e887129-1e7d-40dc-8caa-5e1e0197fb4d"
		case "9e35379a-c333-4471-937e-ac9eeb89cc77",
			"developmental4":
			// Cognitive Skills
			return ""
		case "5c75ab94-c4c8-43b6-a43b-b439f449a7fb",
			"developmental5":
			// Personal Development
			return "5c75ab94-c4c8-43b6-a43b-b439f449a7fb"
		case "ae82bafe-6513-4288-8951-18d93c07e3f1",
			"developmental9":
			// Oral Language
			return "ae82bafe-6513-4288-8951-18d93c07e3f1"
		case "c68865b4-2ba3-4608-955c-dcc098291159",
			"developmental10":
			// Literacy
			return "c68865b4-2ba3-4608-955c-dcc098291159"
		case "61f517d8-2c2e-47fd-a2de-6e86465abc59",
			"developmental11":
			// Whole-Child
			return "61f517d8-2c2e-47fd-a2de-6e86465abc59"
		case "26e4aedc-2222-44e1-a375-388b138c695d",
			"developmental12":
			// Knowledge
			return "26e4aedc-2222-44e1-a375-388b138c695d"
		default:
			// Speech & Language Skills
			return "64e000aa-4a2c-4e2e-9d8d-f779e97bdd73"
		}
	case "56e24fa0-e139-4c80-b365-61c9bc42cd3f":
		// Bada Sound
		switch categoryID {
		case "fc06f364-98fe-487f-97fd-d2d6358dccc6",
			"developmental1":
			// Speech & Language Skills
			return "fc06f364-98fe-487f-97fd-d2d6358dccc6"
		case "0e66242a-4733-4970-a055-d0d6486f8674",
			"developmental2":
			// Fine Motor Skills
			return "0e66242a-4733-4970-a055-d0d6486f8674"
		case "e63956d9-3a36-40b3-a89d-bd45dc8c3181",
			"developmental3":
			// Gross Motor Skills
			return "e63956d9-3a36-40b3-a89d-bd45dc8c3181"
		case "b0b983e4-bf3c-4315-912e-67c8de4f9e11",
			"developmental4":
			// Cognitive Skills
			return "b0b983e4-bf3c-4315-912e-67c8de4f9e11"
		case "84619bee-0b1f-447f-8208-4a39f32062c9",
			"developmental5":
			// Personal Development
			return "84619bee-0b1f-447f-8208-4a39f32062c9"
		case "4b247e7e-dcf9-46a6-a477-a69635142d14",
			"developmental9":
			// Oral Language
			return "4b247e7e-dcf9-46a6-a477-a69635142d14"
		case "59565e03-8d8f-4475-a231-cfc551f004b5",
			"developmental10":
			// Literacy
			return "59565e03-8d8f-4475-a231-cfc551f004b5"
		case "880bc0fd-0209-4f72-999d-3103f9577edf",
			"developmental11":
			// Whole-Child
			return "880bc0fd-0209-4f72-999d-3103f9577edf"
		case "bac3d444-6dcc-4d6c-a4d7-fb6c96fcfc72",
			"developmental12":
			// Knowledge
			return "bac3d444-6dcc-4d6c-a4d7-fb6c96fcfc72"
		default:
			// Speech & Language Skills
			return "fc06f364-98fe-487f-97fd-d2d6358dccc6"
		}
	case "93f293e8-2c6a-47ad-bc46-1554caac99e4":
		// Bada Rhyme
		switch categoryID {
		case "bf1cd84d-da71-4111-82c6-e85224ab85ca",
			"developmental1":
			// Speech & Language Skills
			return "bf1cd84d-da71-4111-82c6-e85224ab85ca"
		case "ba2db2b5-7f20-4cb7-88ef-cee0fcde7937",
			"developmental2":
			// Fine Motor Skills
			return "ba2db2b5-7f20-4cb7-88ef-cee0fcde7937"
		case "07786ea3-ac7b-43e0-bb91-6cd813318185",
			"developmental3":
			// Gross Motor Skills
			return "07786ea3-ac7b-43e0-bb91-6cd813318185"
		case "c3f73955-26f0-49bf-91f7-8c42c81fb9d3",
			"developmental4":
			// Cognitive Skills
			return "c3f73955-26f0-49bf-91f7-8c42c81fb9d3"
		case "aebc88cd-0673-487b-a194-06e3958670a4",
			"developmental5":
			// Personal Development
			return "aebc88cd-0673-487b-a194-06e3958670a4"
		case "22520430-b13e-43ba-930f-fd051bbbc42a",
			"developmental9":
			// Oral Language
			return "22520430-b13e-43ba-930f-fd051bbbc42a"
		case "c3175001-2d1e-4b00-aacf-d188f4ae5cdf",
			"developmental10":
			// Literacy
			return "c3175001-2d1e-4b00-aacf-d188f4ae5cdf"
		case "19ac71c4-04e4-4d1c-8526-1acb292b7137",
			"developmental11":
			// Whole-Child
			return "19ac71c4-04e4-4d1c-8526-1acb292b7137"
		case "d896bf1a-fb5b-4a57-b833-87b0959ba926",
			"developmental12":
			// Knowledge
			return "d896bf1a-fb5b-4a57-b833-87b0959ba926"
		default:
			// Speech & Language Skills
			return "bf1cd84d-da71-4111-82c6-e85224ab85ca"
		}
	case "7565ae11-8130-4b7d-ac24-1d9dd6f792f2":
		// None Specified
		return "2d5ea951-836c-471e-996e-76823a992689"
	default:
		// None Specified
		return "2d5ea951-836c-471e-996e-76823a992689"
	}
}
