package mapping

import "context"

func (s mapperImpl) SubCategory(ctx context.Context, organizationID, programID, categoryID, subCategoryID string) string {
	categoryID = s.Category(ctx, organizationID, programID, categoryID)
	switch categoryID {
	case "84b8f87a-7b61-4580-a190-a9ce3fe90dd3":
		// Speech & Language Skills
		switch subCategoryID {
		case "2b6b5d54-0243-4c7e-917a-1627f107f198",
			"skills1":
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		case "8b955cbc-6808-49b2-adc0-5bec8b59f4fe",
			"skills2":
			// Phonics
			return "8b955cbc-6808-49b2-adc0-5bec8b59f4fe"
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		default:
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		}
	case "ce9014a4-01a9-49d5-bf10-6b08bc454fc1":
		// Fine Motor Skills
		switch subCategoryID {
		case "963729a4-7853-49d2-b75d-2c61d291afee",
			"skills7":
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		default:
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		}
	case "61996d3d-a37d-4873-bcdc-03b22fc6977e":
		// Gross Motor Skills
		switch subCategoryID {
		case "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5",
			"skills9":
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		default:
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		}
	case "e08f3578-a7d4-4cac-b028-ef7a8c93f53f":
		// Cognitive Skills
		switch subCategoryID {
		case "b32321db-3b4a-4b1e-8db9-c485d045bf01",
			"skills12":
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		default:
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		}
	case "76cc6f90-86ef-48b7-9138-7b2f0bc378e7":
		// Personal Development
		switch subCategoryID {
		case "ba77f705-9087-4424-bff9-50fcd0b1731e",
			"skills14":
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		default:
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		}
	case "1080d319-8ce7-4378-9c71-a5019d6b9386":
		// Speech & Language Skills
		switch subCategoryID {
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "43c9d2c5-7a23-42c9-8ad9-1132fb9c3853",
			"skills37":
			// Colors
			return "43c9d2c5-7a23-42c9-8ad9-1132fb9c3853"
		case "8d49bbbb-b230-4d5a-900b-cde6283519a3",
			"skills20":
			// Numbers
			return "8d49bbbb-b230-4d5a-900b-cde6283519a3"
		case "ed88dcc7-30e4-4ec7-bccd-34aaacb47139",
			"skills38":
			// Shapes
			return "ed88dcc7-30e4-4ec7-bccd-34aaacb47139"
		case "1cb17f8a-d516-498c-97ea-8ad4d7a0c018",
			"skills39":
			// Letters
			return "1cb17f8a-d516-498c-97ea-8ad4d7a0c018"
		case "cd06e622-a323-40f3-8409-5384395e00d2",
			"skills40":
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		case "81b09f61-4509-4ce0-b099-c208e62870f9",
			"skills41":
			// Math
			return "81b09f61-4509-4ce0-b099-c208e62870f9"
		case "39ac1475-4ade-4d0b-b79a-f31256521297",
			"skills42":
			// Coding
			return "39ac1475-4ade-4d0b-b79a-f31256521297"
		default:
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		}
	case "f9d82bdd-4ee2-49dd-a707-133407cdef19":
		// Fine Motor Skills
		switch subCategoryID {
		case "963729a4-7853-49d2-b75d-2c61d291afee",
			"skills7":
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		default:
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		}
	case "a1c26321-e3a7-4ff2-9f1c-bb1c5e420fb7":
		// Gross Motor Skills
		switch subCategoryID {
		case "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5",
			"skills9":
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		case "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7",
			"skills10":
			// Complex Movements
			return "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7"
		case "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67",
			"skills11":
			// Physical Skills
			return "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67"
		default:
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		}
	case "c12f363a-633b-4080-bd2b-9ced8d034379":
		// Cognitive Skills
		switch subCategoryID {
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "43c9d2c5-7a23-42c9-8ad9-1132fb9c3853",
			"skills37":
			// Colors
			return "43c9d2c5-7a23-42c9-8ad9-1132fb9c3853"
		case "8d49bbbb-b230-4d5a-900b-cde6283519a3",
			"skills20":
			// Numbers
			return "8d49bbbb-b230-4d5a-900b-cde6283519a3"
		case "ed88dcc7-30e4-4ec7-bccd-34aaacb47139",
			"skills38":
			// Shapes
			return "ed88dcc7-30e4-4ec7-bccd-34aaacb47139"
		case "1cb17f8a-d516-498c-97ea-8ad4d7a0c018",
			"skills39":
			// Letters
			return "1cb17f8a-d516-498c-97ea-8ad4d7a0c018"
		case "cd06e622-a323-40f3-8409-5384395e00d2",
			"skills40":
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		case "81b09f61-4509-4ce0-b099-c208e62870f9",
			"skills41":
			// Math
			return "81b09f61-4509-4ce0-b099-c208e62870f9"
		case "39ac1475-4ade-4d0b-b79a-f31256521297",
			"skills42":
			// Coding
			return "39ac1475-4ade-4d0b-b79a-f31256521297"
		default:
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		}
	case "e06ad483-085c-4869-bd88-56d17c7810a0":
		// Personal Development
		switch subCategoryID {
		case "ba77f705-9087-4424-bff9-50fcd0b1731e",
			"skills14":
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		case "824bb6cb-0169-4335-b7a5-6ece2b929da3",
			"skills15":
			// Emotional Skills
			return "824bb6cb-0169-4335-b7a5-6ece2b929da3"
		default:
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		}
	case "1cc44ecc-153a-47e9-b6e8-3b1ef94a9dee":
		// Speech & Language Skills
		switch subCategoryID {
		case "cd06e622-a323-40f3-8409-5384395e00d2",
			"skills40":
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		case "81b09f61-4509-4ce0-b099-c208e62870f9",
			"skills41":
			// Math
			return "81b09f61-4509-4ce0-b099-c208e62870f9"
		case "39ac1475-4ade-4d0b-b79a-f31256521297",
			"skills42":
			// Coding
			return "39ac1475-4ade-4d0b-b79a-f31256521297"
		default:
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		}
	case "0523610d-cf11-47b6-b7ab-bdbf8c3e09b6":
		// Fine Motor Skills
		switch subCategoryID {
		case "963729a4-7853-49d2-b75d-2c61d291afee",
			"skills7":
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		case "bf89c192-93dd-4192-97ab-f37198548ead",
			"skills8":
			// Hand-Eye Coordination
			return "bf89c192-93dd-4192-97ab-f37198548ead"
		default:
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		}
	case "d1783a8c-6bcd-492a-ad17-37386df80c56":
		// Gross Motor Skills
		switch subCategoryID {
		case "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5",
			"skills9":
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		case "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7",
			"skills10":
			// Complex Movements
			return "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7"
		case "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67",
			"skills11":
			// Physical Skills
			return "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67"
		default:
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		}
	case "1ef6ca6c-fbc4-4422-a5cb-2bcd999e4b2b":
		// Cognitive Skills
		switch subCategoryID {
		case "cd06e622-a323-40f3-8409-5384395e00d2",
			"skills40":
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		case "81b09f61-4509-4ce0-b099-c208e62870f9",
			"skills41":
			// Math
			return "81b09f61-4509-4ce0-b099-c208e62870f9"
		case "39ac1475-4ade-4d0b-b79a-f31256521297",
			"skills42":
			// Coding
			return "39ac1475-4ade-4d0b-b79a-f31256521297"
		case "19803be1-0503-4232-afc1-e6ef06186523",
			"skills43":
			// Experimenting & Problem Solving
			return "19803be1-0503-4232-afc1-e6ef06186523"
		default:
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		}
	case "8488eeac-28bd-4f86-8093-9853b19f51db":
		// Personal Development
		switch subCategoryID {
		case "ba77f705-9087-4424-bff9-50fcd0b1731e",
			"skills14":
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		case "824bb6cb-0169-4335-b7a5-6ece2b929da3",
			"skills15":
			// Emotional Skills
			return "824bb6cb-0169-4335-b7a5-6ece2b929da3"
		default:
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		}
	case "1bb26398-3e38-441e-9a8a-460057f2d8c0":
		// Speech & Language Skills
		switch subCategoryID {
		case "2b6b5d54-0243-4c7e-917a-1627f107f198",
			"skills1":
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		case "8b955cbc-6808-49b2-adc0-5bec8b59f4fe",
			"skills2":
			// Phonics
			return "8b955cbc-6808-49b2-adc0-5bec8b59f4fe"
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "3fca3a2b-97b6-4ec9-a5b1-1d0ef5f1b445",
			"skills5":
			// Reading Skills and Comprehension
			return "3fca3a2b-97b6-4ec9-a5b1-1d0ef5f1b445"
		case "9a9882f1-d890-461c-a710-ca37fb78ddf5",
			"skills6":
			// Sight Words
			return "9a9882f1-d890-461c-a710-ca37fb78ddf5"
		case "0fd7d721-df1b-41eb-baa4-08ba4ac2b2e7",
			"skills4":
			// Thematic Concepts
			return "0fd7d721-df1b-41eb-baa4-08ba4ac2b2e7"
		default:
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		}
	case "e65ea6b4-7093-490a-927e-d2235643f6ca":
		// Fine Motor Skills
		switch subCategoryID {
		case "963729a4-7853-49d2-b75d-2c61d291afee",
			"skills7":
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		case "bf89c192-93dd-4192-97ab-f37198548ead",
			"skills8":
			// Hand-Eye Coordination
			return "bf89c192-93dd-4192-97ab-f37198548ead"
		default:
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		}
	case "88fff890-d614-4b88-be57-b7441fa40b66":
		// Gross Motor Skills
		switch subCategoryID {
		case "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5",
			"skills9":
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		case "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7",
			"skills10":
			// Complex Movements
			return "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7"
		case "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67",
			"skills11":
			// Physical Skills
			return "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67"
		default:
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		}
	case "b18d60c6-a545-46ff-8988-cd5d46ab9660":
		// Cognitive Skills
		switch subCategoryID {
		case "b32321db-3b4a-4b1e-8db9-c485d045bf01",
			"skills12":
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		case "f385c1ec-6cfa-4f49-a219-fd28374cf2a6",
			"skills13":
			// Visual
			return "f385c1ec-6cfa-4f49-a219-fd28374cf2a6"
		default:
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		}
	case "c83fd174-6504-4cc3-9175-2728d023c39d":
		// Personal Development
		switch subCategoryID {
		case "ba77f705-9087-4424-bff9-50fcd0b1731e",
			"skills14":
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		case "824bb6cb-0169-4335-b7a5-6ece2b929da3",
			"skills15":
			// Emotional Skills
			return "824bb6cb-0169-4335-b7a5-6ece2b929da3"
		default:
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		}
	case "d17f1bee-cdef-4759-8c23-3e9b64d08ec1":
		// Oral Language
		switch subCategoryID {
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "b2cc7a69-4e64-4e97-9587-0078dccd845a",
			"skills45":
			// Language Support
			return "b2cc7a69-4e64-4e97-9587-0078dccd845a"
		case "843e4fea-7f4d-4746-87ff-693f5a44b467",
			"skills46":
			// Communication
			return "843e4fea-7f4d-4746-87ff-693f5a44b467"
		case "ec1d6481-ab50-42b6-a4b5-1a5fb98796d0":
			// Phonemic Awareness
			return "ec1d6481-ab50-42b6-a4b5-1a5fb98796d0"
		default:
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		}
	case "dd59f36d-717f-4982-9ae6-df32537faba0":
		// Literacy
		switch subCategoryID {
		case "9b955fb9-8eda-4469-bd31-4e8f91192663",
			"skills48":
			// Emergent Writing
			return "9b955fb9-8eda-4469-bd31-4e8f91192663"
		case "644ba535-904c-4919-8b8c-688df2b6f7ee",
			"skills47":
			// Emergent Reading
			return "644ba535-904c-4919-8b8c-688df2b6f7ee"
		default:
			// Emergent Writing
			return "9b955fb9-8eda-4469-bd31-4e8f91192663"
		}
	case "8d464354-16d9-41af-b887-103f18f4b376":
		// Whole-Child
		switch subCategoryID {
		case "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7",
			"skills50":
			// Cognitive Development
			return "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7"
		case "96f81756-70e3-41e5-9143-740376574e35",
			"skills49":
			// Social-Emotional Learning
			return "96f81756-70e3-41e5-9143-740376574e35"
		case "144a3478-1946-4460-a965-0d7d74e63d65",
			"skills51":
			// Physical Coordination
			return "144a3478-1946-4460-a965-0d7d74e63d65"
		default:
			// Cognitive Development
			return "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7"
		}
	case "dfed32b5-f0bd-42ea-999b-e10b376038d5":
		// Knowledge
		switch subCategoryID {
		case "cd06e622-a323-40f3-8409-5384395e00d2",
			"skills40":
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		case "81b09f61-4509-4ce0-b099-c208e62870f9",
			"skills41":
			// Math
			return "81b09f61-4509-4ce0-b099-c208e62870f9"
		case "5b405510-384a-4721-a526-e12b3cbf2092",
			"skills53":
			// Engineering
			return "5b405510-384a-4721-a526-e12b3cbf2092"
		case "9a52fb0a-6ce8-45df-92a0-f25b5d3d2344",
			"5fb241f2993e7591084c8400":
			// Music
			return "9a52fb0a-6ce8-45df-92a0-f25b5d3d2344"
		case "4114f381-a7c5-4e88-be84-2bef4eb04ad0",
			"5fb241e7993e7591084c83fd":
			// Health
			return "4114f381-a7c5-4e88-be84-2bef4eb04ad0"
		case "f4b07251-1d67-4a84-bcda-86c71cbf9cfd",
			"5fb241d2993e7591084c83fa":
			// Social Studies
			return "f4b07251-1d67-4a84-bcda-86c71cbf9cfd"
		case "49e73e4f-8ffc-47e3-9b87-0f9686d361d7",
			"skills52":
			// Technology
			return "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"
		case "852c3495-1ced-4580-a584-9d475217f3d5",
			"5fb241fe993e7591084c8403":
			// Character Education
			return "852c3495-1ced-4580-a584-9d475217f3d5"
		case "3b148168-31d0-4bef-9152-63c3ff516180",
			"skills54":
			// Miscellaneous
			return "3b148168-31d0-4bef-9152-63c3ff516180"
		case "6fb79402-2fb6-4415-874c-338c949332ed",
			"5fb241c7993e7591084c83f7":
			// Art
			return "6fb79402-2fb6-4415-874c-338c949332ed"
		default:
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		}
	case "2a637bea-c529-4868-8269-d0936696da7e":
		// Language and Numeracy Skills
		switch subCategoryID {
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "8d49bbbb-b230-4d5a-900b-cde6283519a3",
			"skills20":
			// Numbers
			return "8d49bbbb-b230-4d5a-900b-cde6283519a3"
		case "c06b848d-8769-44e9-8dc7-929588cec0bc",
			"skills18":
			// Speaking
			return "c06b848d-8769-44e9-8dc7-929588cec0bc"
		case "01191172-b276-449f-ab11-8e66e990941e",
			"5fb242d6993e7591084c842d":
			// Reading
			return "01191172-b276-449f-ab11-8e66e990941e"
		case "55cbd434-36ce-4c57-b47e-d7119b578d7e",
			"skills21":
			// Fluency
			return "55cbd434-36ce-4c57-b47e-d7119b578d7e"
		case "a048cf91-2c96-4306-a7c2-cac2fe1d688a",
			"skills16":
			// Reasoning
			return "a048cf91-2c96-4306-a7c2-cac2fe1d688a"
		case "ddf87dff-1eb0-4971-9b27-2aaa534f34b1",
			"skills17":
			// Listening
			return "ddf87dff-1eb0-4971-9b27-2aaa534f34b1"
		case "47169b0a-ac39-4e25-bd6e-77eecaf4e051",
			"skills19":
			// Interpreting
			return "47169b0a-ac39-4e25-bd6e-77eecaf4e051"
		default:
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		}
	case "6933de3e-a568-4d56-8c01-e110bda22926":
		// Fine Motor Skills
		switch subCategoryID {
		case "11351e3f-afc3-476e-b3af-a0c7718269ac",
			"5fb243e1993e7591084c844e":
			// Coloring
			return "11351e3f-afc3-476e-b3af-a0c7718269ac"
		case "d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb",
			"5fb243f4993e7591084c8454":
			// Drag
			return "d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb"
		case "e2190c0c-918d-4a05-a045-6696ae31d5c4",
			"5fb243f9993e7591084c8457":
			// Click
			return "e2190c0c-918d-4a05-a045-6696ae31d5c4"
		case "a7850bd6-f5fd-4016-b708-7b823784ef0a",
			"skills33":
			// Writing
			return "a7850bd6-f5fd-4016-b708-7b823784ef0a"
		case "bea9244e-ff17-47fc-8e7c-bceadf0f4f6e",
			"5fb243d2993e7591084c8448":
			// Drawing
			return "bea9244e-ff17-47fc-8e7c-bceadf0f4f6e"
		case "7848bb23-2bb9-4108-938b-51f2f7d1d30f",
			"5fb243db993e7591084c844b":
			// Tracing
			return "7848bb23-2bb9-4108-938b-51f2f7d1d30f"
		default:
			// Coloring
			return "11351e3f-afc3-476e-b3af-a0c7718269ac"
		}
	case "3af9f093-4153-4348-a097-986c15d1f912":
		// Gross Motor Skills
		switch subCategoryID {
		case "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5",
			"skills9":
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		case "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7",
			"skills10":
			// Complex Movements
			return "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7"
		case "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67",
			"skills11":
			// Physical Skills
			return "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67"
		case "9c30644b-0e9c-43aa-a19a-442e9f6aa6ae",
			"skills24":
			// Body Coordination
			return "9c30644b-0e9c-43aa-a19a-442e9f6aa6ae"
		default:
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		}
	case "a11a6f56-3ae3-4b70-8daa-30cdb63ef5b6":
		// Cognitive
		switch subCategoryID {
		case "367c5e70-1487-4b33-96c0-529a37dbc5f2",
			"skills26":
			// Counting and Operations
			return "367c5e70-1487-4b33-96c0-529a37dbc5f2"
		case "e45ff0ff-40a4-4be4-ab26-426aedba7597",
			"skills25":
			// Spatial Representation
			return "e45ff0ff-40a4-4be4-ab26-426aedba7597"
		case "ff838eb9-11b9-4de5-b854-a24d4d526f5e",
			"skills27":
			// Logical Problem-Solving
			return "ff838eb9-11b9-4de5-b854-a24d4d526f5e"
		case "4ab80faf-60b9-4cc2-8f51-3d3b7f9fee13",
			"skills28":
			// Patterns
			return "4ab80faf-60b9-4cc2-8f51-3d3b7f9fee13"
		default:
			// Counting and Operations
			return "367c5e70-1487-4b33-96c0-529a37dbc5f2"
		}
	case "665616dd-32c2-44c4-91c9-63f7493c9fd3":
		// Social and Emotional
		switch subCategoryID {
		case "b79735db-91c7-4bcb-860b-fe23902f81ea",
			"skills29":
			// Social Interactions
			return "b79735db-91c7-4bcb-860b-fe23902f81ea"
		case "6ccc8306-1a9e-42bd-83ff-55bac3449853",
			"skills32":
			// Self-Control
			return "6ccc8306-1a9e-42bd-83ff-55bac3449853"
		case "c79be603-ccf4-4284-9c8e-61b55ec53067",
			"skills31":
			// Self-Identity
			return "c79be603-ccf4-4284-9c8e-61b55ec53067"
		case "188c621a-cbc7-42e2-9d01-56f4847682cb",
			"skills30":
			// Empathy
			return "188c621a-cbc7-42e2-9d01-56f4847682cb"
		default:
			// Social Interactions
			return "b79735db-91c7-4bcb-860b-fe23902f81ea"
		}
	case "6090e473-ec19-4bf0-ae5c-2d6a4c793f55":
		// Speech & Language Skills
		switch subCategoryID {
		case "8b955cbc-6808-49b2-adc0-5bec8b59f4fe",
			"skills2":
			// Phonics
			return "8b955cbc-6808-49b2-adc0-5bec8b59f4fe"
		case "c06b848d-8769-44e9-8dc7-929588cec0bc",
			"skills18":
			// Speaking
			return "c06b848d-8769-44e9-8dc7-929588cec0bc"
		case "01191172-b276-449f-ab11-8e66e990941e",
			"5fb242d6993e7591084c842d":
			// Reading
			return "01191172-b276-449f-ab11-8e66e990941e"
		case "a7850bd6-f5fd-4016-b708-7b823784ef0a",
			"skills33":
			// Writing
			return "a7850bd6-f5fd-4016-b708-7b823784ef0a"
		case "55cbd434-36ce-4c57-b47e-d7119b578d7e",
			"skills21":
			// Fluency
			return "55cbd434-36ce-4c57-b47e-d7119b578d7e"
		case "eb29827a-0053-4eee-83cd-8f4afb1b7cb4",
			"5fb24df3993e7591084c85a2":
			// Comprehension
			return "eb29827a-0053-4eee-83cd-8f4afb1b7cb4"
		case "ddf87dff-1eb0-4971-9b27-2aaa534f34b1",
			"skills17":
			// Listening
			return "ddf87dff-1eb0-4971-9b27-2aaa534f34b1"
		default:
			// Phonics
			return "8b955cbc-6808-49b2-adc0-5bec8b59f4fe"
		}
	case "da9fa132-dcf7-4148-9037-b381850ba088":
		// Fine Motor Skills
		switch subCategoryID {
		case "11351e3f-afc3-476e-b3af-a0c7718269ac",
			"5fb243e1993e7591084c844e":
			// Coloring
			return "11351e3f-afc3-476e-b3af-a0c7718269ac"
		case "d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb",
			"5fb243f4993e7591084c8454":
			// Drag
			return "d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb"
		case "e2190c0c-918d-4a05-a045-6696ae31d5c4",
			"5fb243f9993e7591084c8457":
			// Click
			return "e2190c0c-918d-4a05-a045-6696ae31d5c4"
		case "a7850bd6-f5fd-4016-b708-7b823784ef0a",
			"skills33":
			// Writing
			return "a7850bd6-f5fd-4016-b708-7b823784ef0a"
		case "bea9244e-ff17-47fc-8e7c-bceadf0f4f6e",
			"5fb243d2993e7591084c8448":
			// Drawing
			return "bea9244e-ff17-47fc-8e7c-bceadf0f4f6e"
		case "7848bb23-2bb9-4108-938b-51f2f7d1d30f",
			"5fb243db993e7591084c844b":
			// Tracing
			return "7848bb23-2bb9-4108-938b-51f2f7d1d30f"
		default:
			// Coloring
			return "11351e3f-afc3-476e-b3af-a0c7718269ac"
		}
	case "585f38e6-f7be-45f2-855a-f2a4bddca125":
		// Gross Motor Skills
		switch subCategoryID {
		case "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5",
			"skills9":
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		case "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7",
			"skills10":
			// Complex Movements
			return "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7"
		case "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67",
			"skills11":
			// Physical Skills
			return "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67"
		default:
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		}
	case "c3ea1b4a-d220-4248-9b3f-07559b415c56":
		// Cognitive Skills
		switch subCategoryID {
		case "b9d5a570-5be3-491b-9fdc-d26ea1c13847",
			"skills36":
			// Reasoning Skills
			return "b9d5a570-5be3-491b-9fdc-d26ea1c13847"
		case "9a1e0589-0361-40e1-851c-b95b641e271e",
			"skills35":
			// Critical Thinking (Interpretation, Analysis, Evaluation, Inference, Explanation, and Self-Regulation)
			return "9a1e0589-0361-40e1-851c-b95b641e271e"
		case "8d3f987a-7f7c-4035-a709-9526060b2177",
			"skills34":
			// Science Process (Observing, Classifying, Communicating, Measuring, Predicting)
			return "8d3f987a-7f7c-4035-a709-9526060b2177"
		default:
			// Reasoning Skills
			return "b9d5a570-5be3-491b-9fdc-d26ea1c13847"
		}
	case "7826ff58-25d0-41f1-b38e-3e3a77ed32f6":
		// Social and Emotional
		switch subCategoryID {
		case "b79735db-91c7-4bcb-860b-fe23902f81ea",
			"skills29":
			// Social Interactions
			return "b79735db-91c7-4bcb-860b-fe23902f81ea"
		case "6ccc8306-1a9e-42bd-83ff-55bac3449853",
			"skills32":
			// Self-Control
			return "6ccc8306-1a9e-42bd-83ff-55bac3449853"
		case "c79be603-ccf4-4284-9c8e-61b55ec53067",
			"skills31":
			// Self-Identity
			return "c79be603-ccf4-4284-9c8e-61b55ec53067"
		case "188c621a-cbc7-42e2-9d01-56f4847682cb",
			"skills30":
			// Empathy
			return "188c621a-cbc7-42e2-9d01-56f4847682cb"
		default:
			// Social Interactions
			return "b79735db-91c7-4bcb-860b-fe23902f81ea"
		}
	case "b8c76823-150d-4d83-861e-dce7d7bc4f6d":
		// Speech & Language Skills
		switch subCategoryID {
		case "2b6b5d54-0243-4c7e-917a-1627f107f198",
			"skills1":
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		case "8b955cbc-6808-49b2-adc0-5bec8b59f4fe",
			"skills2":
			// Phonics
			return "8b955cbc-6808-49b2-adc0-5bec8b59f4fe"
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "3fca3a2b-97b6-4ec9-a5b1-1d0ef5f1b445",
			"skills5":
			// Reading Skills and Comprehension
			return "3fca3a2b-97b6-4ec9-a5b1-1d0ef5f1b445"
		case "9a9882f1-d890-461c-a710-ca37fb78ddf5",
			"skills6":
			// Sight Words
			return "9a9882f1-d890-461c-a710-ca37fb78ddf5"
		case "0fd7d721-df1b-41eb-baa4-08ba4ac2b2e7",
			"skills4":
			// Thematic Concepts
			return "0fd7d721-df1b-41eb-baa4-08ba4ac2b2e7"
		default:
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		}
	case "b4cd42b8-a09b-4f66-a03a-b9f6b6f69895":
		// Fine Motor Skills
		switch subCategoryID {
		case "963729a4-7853-49d2-b75d-2c61d291afee",
			"skills7":
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		case "bf89c192-93dd-4192-97ab-f37198548ead",
			"skills8":
			// Hand-Eye Coordination
			return "bf89c192-93dd-4192-97ab-f37198548ead"
		default:
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		}
	case "bcfd9d76-cf05-4ccd-9a41-6b886da661be":
		// Gross Motor Skills
		switch subCategoryID {
		case "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5",
			"skills9":
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		case "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7",
			"skills10":
			// Complex Movements
			return "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7"
		case "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67",
			"skills11":
			// Physical Skills
			return "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67"
		default:
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		}
	case "c909acad-0c52-4fd3-8427-3b1e90a730da":
		// Cognitive Skills
		switch subCategoryID {
		case "b32321db-3b4a-4b1e-8db9-c485d045bf01",
			"skills12":
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		case "f385c1ec-6cfa-4f49-a219-fd28374cf2a6",
			"skills13":
			// Visual
			return "f385c1ec-6cfa-4f49-a219-fd28374cf2a6"
		default:
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		}
	case "fa8ff09d-9062-4955-9b20-5fa20757f1d9":
		// Personal Development
		switch subCategoryID {
		case "ba77f705-9087-4424-bff9-50fcd0b1731e",
			"skills14":
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		case "824bb6cb-0169-4335-b7a5-6ece2b929da3",
			"skills15":
			// Emotional Skills
			return "824bb6cb-0169-4335-b7a5-6ece2b929da3"
		default:
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		}
	case "29a0ab9e-6364-47b6-b63a-1388a7861c6c":
		// Oral Language
		switch subCategoryID {
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "b2cc7a69-4e64-4e97-9587-0078dccd845a",
			"skills45":
			// Language Support
			return "b2cc7a69-4e64-4e97-9587-0078dccd845a"
		case "843e4fea-7f4d-4746-87ff-693f5a44b467",
			"skills46":
			// Communication
			return "843e4fea-7f4d-4746-87ff-693f5a44b467"
		case "5bb19c81-9261-428e-95ed-c87cc9f0560b",
			"skills44":
			// Phonological Awareness
			return "5bb19c81-9261-428e-95ed-c87cc9f0560b"
		default:
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		}
	case "49cbbf19-2ad7-4acb-b8c8-66531578116a":
		// Literacy
		switch subCategoryID {
		case "9b955fb9-8eda-4469-bd31-4e8f91192663",
			"skills48":
			// Emergent Writing
			return "9b955fb9-8eda-4469-bd31-4e8f91192663"
		case "644ba535-904c-4919-8b8c-688df2b6f7ee",
			"skills47":
			// Emergent Reading
			return "644ba535-904c-4919-8b8c-688df2b6f7ee"
		default:
			// Emergent Writing
			return "9b955fb9-8eda-4469-bd31-4e8f91192663"
		}
	case "bd55fd6b-73ef-41ed-8a86-d7bbc501e773":
		// Whole-Child
		switch subCategoryID {
		case "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7",
			"skills50":
			// Cognitive Development
			return "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7"
		case "96f81756-70e3-41e5-9143-740376574e35",
			"skills49":
			// Social-Emotional Learning
			return "96f81756-70e3-41e5-9143-740376574e35"
		case "144a3478-1946-4460-a965-0d7d74e63d65",
			"skills51":
			// Physical Coordination
			return "144a3478-1946-4460-a965-0d7d74e63d65"
		default:
			// Cognitive Development
			return "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7"
		}
	case "dd3dbf0c-2872-433b-8b61-9ea78f3c9e97":
		// Knowledge
		switch subCategoryID {
		case "cd06e622-a323-40f3-8409-5384395e00d2",
			"skills40":
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		case "81b09f61-4509-4ce0-b099-c208e62870f9",
			"skills41":
			// Math
			return "81b09f61-4509-4ce0-b099-c208e62870f9"
		case "5b405510-384a-4721-a526-e12b3cbf2092",
			"skills53":
			// Engineering
			return "5b405510-384a-4721-a526-e12b3cbf2092"
		case "9a52fb0a-6ce8-45df-92a0-f25b5d3d2344",
			"5fb241f2993e7591084c8400":
			// Music
			return "9a52fb0a-6ce8-45df-92a0-f25b5d3d2344"
		case "4114f381-a7c5-4e88-be84-2bef4eb04ad0",
			"5fb241e7993e7591084c83fd":
			// Health
			return "4114f381-a7c5-4e88-be84-2bef4eb04ad0"
		case "f4b07251-1d67-4a84-bcda-86c71cbf9cfd",
			"5fb241d2993e7591084c83fa":
			// Social Studies
			return "f4b07251-1d67-4a84-bcda-86c71cbf9cfd"
		case "49e73e4f-8ffc-47e3-9b87-0f9686d361d7",
			"skills52":
			// Technology
			return "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"
		case "852c3495-1ced-4580-a584-9d475217f3d5",
			"5fb241fe993e7591084c8403":
			// Character Education
			return "852c3495-1ced-4580-a584-9d475217f3d5"
		case "3b148168-31d0-4bef-9152-63c3ff516180",
			"skills54":
			// Miscellaneous
			return "3b148168-31d0-4bef-9152-63c3ff516180"
		case "6fb79402-2fb6-4415-874c-338c949332ed",
			"5fb241c7993e7591084c83f7":
			// Art
			return "6fb79402-2fb6-4415-874c-338c949332ed"
		default:
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		}
	case "64e000aa-4a2c-4e2e-9d8d-f779e97bdd73":
		// Speech & Language Skills
		switch subCategoryID {
		case "2b6b5d54-0243-4c7e-917a-1627f107f198",
			"skills1":
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		case "8b955cbc-6808-49b2-adc0-5bec8b59f4fe",
			"skills2":
			// Phonics
			return "8b955cbc-6808-49b2-adc0-5bec8b59f4fe"
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "3fca3a2b-97b6-4ec9-a5b1-1d0ef5f1b445",
			"skills5":
			// Reading Skills and Comprehension
			return "3fca3a2b-97b6-4ec9-a5b1-1d0ef5f1b445"
		case "9a9882f1-d890-461c-a710-ca37fb78ddf5",
			"skills6":
			// Sight Words
			return "9a9882f1-d890-461c-a710-ca37fb78ddf5"
		case "0fd7d721-df1b-41eb-baa4-08ba4ac2b2e7",
			"skills4":
			// Thematic Concepts
			return "0fd7d721-df1b-41eb-baa4-08ba4ac2b2e7"
		default:
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		}
	case "59c47920-4d0d-477c-a33b-06e7f13873d7":
		// Fine Motor Skills
		switch subCategoryID {
		case "963729a4-7853-49d2-b75d-2c61d291afee",
			"skills7":
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		case "bf89c192-93dd-4192-97ab-f37198548ead",
			"skills8":
			// Hand-Eye Coordination
			return "bf89c192-93dd-4192-97ab-f37198548ead"
		default:
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		}
	case "7e887129-1e7d-40dc-8caa-5e1e0197fb4d":
		// Gross Motor Skills
		switch subCategoryID {
		case "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5",
			"skills9":
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		case "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7",
			"skills10":
			// Complex Movements
			return "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7"
		case "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67",
			"skills11":
			// Physical Skills
			return "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67"
		default:
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		}
	case "9e35379a-c333-4471-937e-ac9eeb89cc77":
		// Cognitive Skills
		switch subCategoryID {
		case "b32321db-3b4a-4b1e-8db9-c485d045bf01",
			"skills12":
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		case "f385c1ec-6cfa-4f49-a219-fd28374cf2a6",
			"skills13":
			// Visual
			return "f385c1ec-6cfa-4f49-a219-fd28374cf2a6"
		default:
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		}
	case "5c75ab94-c4c8-43b6-a43b-b439f449a7fb":
		// Personal Development
		switch subCategoryID {
		case "ba77f705-9087-4424-bff9-50fcd0b1731e",
			"skills14":
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		case "824bb6cb-0169-4335-b7a5-6ece2b929da3",
			"skills15":
			// Emotional Skills
			return "824bb6cb-0169-4335-b7a5-6ece2b929da3"
		default:
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		}
	case "ae82bafe-6513-4288-8951-18d93c07e3f1":
		// Oral Language
		switch subCategoryID {
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "b2cc7a69-4e64-4e97-9587-0078dccd845a",
			"skills45":
			// Language Support
			return "b2cc7a69-4e64-4e97-9587-0078dccd845a"
		case "843e4fea-7f4d-4746-87ff-693f5a44b467",
			"skills46":
			// Communication
			return "843e4fea-7f4d-4746-87ff-693f5a44b467"
		default:
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		}
	case "c68865b4-2ba3-4608-955c-dcc098291159":
		// Literacy
		switch subCategoryID {
		case "01191172-b276-449f-ab11-8e66e990941e",
			"5fb242d6993e7591084c842d":
			// Reading
			return "01191172-b276-449f-ab11-8e66e990941e"
		case "a7850bd6-f5fd-4016-b708-7b823784ef0a",
			"skills33":
			// Writing
			return "a7850bd6-f5fd-4016-b708-7b823784ef0a"
		default:
			// Reading
			return "01191172-b276-449f-ab11-8e66e990941e"
		}
	case "61f517d8-2c2e-47fd-a2de-6e86465abc59":
		// Whole-Child
		switch subCategoryID {
		case "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7",
			"skills50":
			// Cognitive Development
			return "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7"
		case "96f81756-70e3-41e5-9143-740376574e35",
			"skills49":
			// Social-Emotional Learning
			return "96f81756-70e3-41e5-9143-740376574e35"
		case "144a3478-1946-4460-a965-0d7d74e63d65",
			"skills51":
			// Physical Coordination
			return "144a3478-1946-4460-a965-0d7d74e63d65"
		default:
			// Cognitive Development
			return "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7"
		}
	case "26e4aedc-2222-44e1-a375-388b138c695d":
		// Knowledge
		switch subCategoryID {
		case "cd06e622-a323-40f3-8409-5384395e00d2",
			"skills40":
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		case "81b09f61-4509-4ce0-b099-c208e62870f9",
			"skills41":
			// Math
			return "81b09f61-4509-4ce0-b099-c208e62870f9"
		case "5b405510-384a-4721-a526-e12b3cbf2092",
			"skills53":
			// Engineering
			return "5b405510-384a-4721-a526-e12b3cbf2092"
		case "9a52fb0a-6ce8-45df-92a0-f25b5d3d2344",
			"5fb241f2993e7591084c8400":
			// Music
			return "9a52fb0a-6ce8-45df-92a0-f25b5d3d2344"
		case "4114f381-a7c5-4e88-be84-2bef4eb04ad0",
			"5fb241e7993e7591084c83fd":
			// Health
			return "4114f381-a7c5-4e88-be84-2bef4eb04ad0"
		case "f4b07251-1d67-4a84-bcda-86c71cbf9cfd",
			"5fb241d2993e7591084c83fa":
			// Social Studies
			return "f4b07251-1d67-4a84-bcda-86c71cbf9cfd"
		case "49e73e4f-8ffc-47e3-9b87-0f9686d361d7",
			"skills52":
			// Technology
			return "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"
		case "852c3495-1ced-4580-a584-9d475217f3d5",
			"5fb241fe993e7591084c8403":
			// Character Education
			return "852c3495-1ced-4580-a584-9d475217f3d5"
		case "3b148168-31d0-4bef-9152-63c3ff516180",
			"skills54":
			// Miscellaneous
			return "3b148168-31d0-4bef-9152-63c3ff516180"
		case "6fb79402-2fb6-4415-874c-338c949332ed",
			"5fb241c7993e7591084c83f7":
			// Art
			return "6fb79402-2fb6-4415-874c-338c949332ed"
		default:
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		}
	case "fc06f364-98fe-487f-97fd-d2d6358dccc6":
		// Speech & Language Skills
		switch subCategoryID {
		case "2b6b5d54-0243-4c7e-917a-1627f107f198",
			"skills1":
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		case "8b955cbc-6808-49b2-adc0-5bec8b59f4fe",
			"skills2":
			// Phonics
			return "8b955cbc-6808-49b2-adc0-5bec8b59f4fe"
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "3fca3a2b-97b6-4ec9-a5b1-1d0ef5f1b445",
			"skills5":
			// Reading Skills and Comprehension
			return "3fca3a2b-97b6-4ec9-a5b1-1d0ef5f1b445"
		case "9a9882f1-d890-461c-a710-ca37fb78ddf5",
			"skills6":
			// Sight Words
			return "9a9882f1-d890-461c-a710-ca37fb78ddf5"
		case "0fd7d721-df1b-41eb-baa4-08ba4ac2b2e7",
			"skills4":
			// Thematic Concepts
			return "0fd7d721-df1b-41eb-baa4-08ba4ac2b2e7"
		default:
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		}
	case "0e66242a-4733-4970-a055-d0d6486f8674":
		// Fine Motor Skills
		switch subCategoryID {
		case "963729a4-7853-49d2-b75d-2c61d291afee",
			"skills7":
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		case "bf89c192-93dd-4192-97ab-f37198548ead",
			"skills8":
			// Hand-Eye Coordination
			return "bf89c192-93dd-4192-97ab-f37198548ead"
		default:
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		}
	case "e63956d9-3a36-40b3-a89d-bd45dc8c3181":
		// Gross Motor Skills
		switch subCategoryID {
		case "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5",
			"skills9":
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		case "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7",
			"skills10":
			// Complex Movements
			return "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7"
		case "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67",
			"skills11":
			// Physical Skills
			return "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67"
		default:
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		}
	case "b0b983e4-bf3c-4315-912e-67c8de4f9e11":
		// Cognitive Skills
		switch subCategoryID {
		case "b32321db-3b4a-4b1e-8db9-c485d045bf01",
			"skills12":
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		case "f385c1ec-6cfa-4f49-a219-fd28374cf2a6",
			"skills13":
			// Visual
			return "f385c1ec-6cfa-4f49-a219-fd28374cf2a6"
		default:
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		}
	case "84619bee-0b1f-447f-8208-4a39f32062c9":
		// Personal Development
		switch subCategoryID {
		case "ba77f705-9087-4424-bff9-50fcd0b1731e",
			"skills14":
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		case "824bb6cb-0169-4335-b7a5-6ece2b929da3",
			"skills15":
			// Emotional Skills
			return "824bb6cb-0169-4335-b7a5-6ece2b929da3"
		default:
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		}
	case "4b247e7e-dcf9-46a6-a477-a69635142d14":
		// Oral Language
		switch subCategoryID {
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "b2cc7a69-4e64-4e97-9587-0078dccd845a",
			"skills45":
			// Language Support
			return "b2cc7a69-4e64-4e97-9587-0078dccd845a"
		case "843e4fea-7f4d-4746-87ff-693f5a44b467",
			"skills46":
			// Communication
			return "843e4fea-7f4d-4746-87ff-693f5a44b467"
		default:
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		}
	case "59565e03-8d8f-4475-a231-cfc551f004b5":
		// Literacy
		switch subCategoryID {
		case "01191172-b276-449f-ab11-8e66e990941e",
			"5fb242d6993e7591084c842d":
			// Reading
			return "01191172-b276-449f-ab11-8e66e990941e"
		case "a7850bd6-f5fd-4016-b708-7b823784ef0a",
			"skills33":
			// Writing
			return "a7850bd6-f5fd-4016-b708-7b823784ef0a"
		case "39e96a23-5ac3-47c9-94fc-e71965f75880",
			"5fb24c82993e7591084c856d":
			// Phonemic Awareness, Phonics, and Word Recognition
			return "39e96a23-5ac3-47c9-94fc-e71965f75880"
		default:
			// Reading
			return "01191172-b276-449f-ab11-8e66e990941e"
		}
	case "880bc0fd-0209-4f72-999d-3103f9577edf":
		// Whole-Child
		switch subCategoryID {
		case "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7",
			"skills50":
			// Cognitive Development
			return "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7"
		case "96f81756-70e3-41e5-9143-740376574e35",
			"skills49":
			// Social-Emotional Learning
			return "96f81756-70e3-41e5-9143-740376574e35"
		case "144a3478-1946-4460-a965-0d7d74e63d65",
			"skills51":
			// Physical Coordination
			return "144a3478-1946-4460-a965-0d7d74e63d65"
		default:
			// Cognitive Development
			return "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7"
		}
	case "bac3d444-6dcc-4d6c-a4d7-fb6c96fcfc72":
		// Knowledge
		switch subCategoryID {
		case "cd06e622-a323-40f3-8409-5384395e00d2",
			"skills40":
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		case "81b09f61-4509-4ce0-b099-c208e62870f9",
			"skills41":
			// Math
			return "81b09f61-4509-4ce0-b099-c208e62870f9"
		case "5b405510-384a-4721-a526-e12b3cbf2092",
			"skills53":
			// Engineering
			return "5b405510-384a-4721-a526-e12b3cbf2092"
		case "9a52fb0a-6ce8-45df-92a0-f25b5d3d2344",
			"5fb241f2993e7591084c8400":
			// Music
			return "9a52fb0a-6ce8-45df-92a0-f25b5d3d2344"
		case "4114f381-a7c5-4e88-be84-2bef4eb04ad0",
			"5fb241e7993e7591084c83fd":
			// Health
			return "4114f381-a7c5-4e88-be84-2bef4eb04ad0"
		case "f4b07251-1d67-4a84-bcda-86c71cbf9cfd",
			"5fb241d2993e7591084c83fa":
			// Social Studies
			return "f4b07251-1d67-4a84-bcda-86c71cbf9cfd"
		case "49e73e4f-8ffc-47e3-9b87-0f9686d361d7",
			"skills52":
			// Technology
			return "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"
		case "852c3495-1ced-4580-a584-9d475217f3d5",
			"5fb241fe993e7591084c8403":
			// Character Education
			return "852c3495-1ced-4580-a584-9d475217f3d5"
		case "3b148168-31d0-4bef-9152-63c3ff516180",
			"skills54":
			// Miscellaneous
			return "3b148168-31d0-4bef-9152-63c3ff516180"
		case "6fb79402-2fb6-4415-874c-338c949332ed",
			"5fb241c7993e7591084c83f7":
			// Art
			return "6fb79402-2fb6-4415-874c-338c949332ed"
		default:
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		}
	case "bf1cd84d-da71-4111-82c6-e85224ab85ca":
		// Speech & Language Skills
		switch subCategoryID {
		case "2b6b5d54-0243-4c7e-917a-1627f107f198",
			"skills1":
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		case "8b955cbc-6808-49b2-adc0-5bec8b59f4fe",
			"skills2":
			// Phonics
			return "8b955cbc-6808-49b2-adc0-5bec8b59f4fe"
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "3fca3a2b-97b6-4ec9-a5b1-1d0ef5f1b445",
			"skills5":
			// Reading Skills and Comprehension
			return "3fca3a2b-97b6-4ec9-a5b1-1d0ef5f1b445"
		case "9a9882f1-d890-461c-a710-ca37fb78ddf5",
			"skills6":
			// Sight Words
			return "9a9882f1-d890-461c-a710-ca37fb78ddf5"
		case "0fd7d721-df1b-41eb-baa4-08ba4ac2b2e7",
			"skills4":
			// Thematic Concepts
			return "0fd7d721-df1b-41eb-baa4-08ba4ac2b2e7"
		default:
			// Speaking & Listening
			return "2b6b5d54-0243-4c7e-917a-1627f107f198"
		}
	case "ba2db2b5-7f20-4cb7-88ef-cee0fcde7937":
		// Fine Motor Skills
		switch subCategoryID {
		case "963729a4-7853-49d2-b75d-2c61d291afee",
			"skills7":
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		case "bf89c192-93dd-4192-97ab-f37198548ead",
			"skills8":
			// Hand-Eye Coordination
			return "bf89c192-93dd-4192-97ab-f37198548ead"
		default:
			// Sensory
			return "963729a4-7853-49d2-b75d-2c61d291afee"
		}
	case "07786ea3-ac7b-43e0-bb91-6cd813318185":
		// Gross Motor Skills
		switch subCategoryID {
		case "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5",
			"skills9":
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		case "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7",
			"skills10":
			// Complex Movements
			return "f78c01f9-4b8a-480c-8c4b-80d1ec1747a7"
		case "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67",
			"skills11":
			// Physical Skills
			return "f5a1e3a6-c0b1-4b2f-991f-9df7897dac67"
		default:
			// Simple Movements
			return "bd7adbd0-9ce7-4c50-aa8e-85b842683fb5"
		}
	case "c3f73955-26f0-49bf-91f7-8c42c81fb9d3":
		// Cognitive Skills
		switch subCategoryID {
		case "b32321db-3b4a-4b1e-8db9-c485d045bf01",
			"skills12":
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		case "f385c1ec-6cfa-4f49-a219-fd28374cf2a6",
			"skills13":
			// Visual
			return "f385c1ec-6cfa-4f49-a219-fd28374cf2a6"
		default:
			// Logic & Memory
			return "b32321db-3b4a-4b1e-8db9-c485d045bf01"
		}
	case "aebc88cd-0673-487b-a194-06e3958670a4":
		// Personal Development
		switch subCategoryID {
		case "ba77f705-9087-4424-bff9-50fcd0b1731e",
			"skills14":
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		case "824bb6cb-0169-4335-b7a5-6ece2b929da3",
			"skills15":
			// Emotional Skills
			return "824bb6cb-0169-4335-b7a5-6ece2b929da3"
		default:
			// Social Skills
			return "ba77f705-9087-4424-bff9-50fcd0b1731e"
		}
	case "22520430-b13e-43ba-930f-fd051bbbc42a":
		// Oral Language
		switch subCategoryID {
		case "2d1152a3-fb03-4c4e-aeba-98856c3241bd",
			"skills3":
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		case "b2cc7a69-4e64-4e97-9587-0078dccd845a",
			"skills45":
			// Language Support
			return "b2cc7a69-4e64-4e97-9587-0078dccd845a"
		case "843e4fea-7f4d-4746-87ff-693f5a44b467",
			"skills46":
			// Communication
			return "843e4fea-7f4d-4746-87ff-693f5a44b467"
		case "5bb19c81-9261-428e-95ed-c87cc9f0560b",
			"skills44":
			// Phonological Awareness
			return "5bb19c81-9261-428e-95ed-c87cc9f0560b"
		default:
			// Vocabulary
			return "2d1152a3-fb03-4c4e-aeba-98856c3241bd"
		}
	case "c3175001-2d1e-4b00-aacf-d188f4ae5cdf":
		// Literacy
		switch subCategoryID {
		case "9b955fb9-8eda-4469-bd31-4e8f91192663",
			"skills48":
			// Emergent Writing
			return "9b955fb9-8eda-4469-bd31-4e8f91192663"
		case "644ba535-904c-4919-8b8c-688df2b6f7ee",
			"skills47":
			// Emergent Reading
			return "644ba535-904c-4919-8b8c-688df2b6f7ee"
		default:
			// Emergent Writing
			return "9b955fb9-8eda-4469-bd31-4e8f91192663"
		}
	case "19ac71c4-04e4-4d1c-8526-1acb292b7137":
		// Whole-Child
		switch subCategoryID {
		case "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7",
			"skills50":
			// Cognitive Development
			return "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7"
		case "96f81756-70e3-41e5-9143-740376574e35",
			"skills49":
			// Social-Emotional Learning
			return "96f81756-70e3-41e5-9143-740376574e35"
		case "144a3478-1946-4460-a965-0d7d74e63d65",
			"skills51":
			// Physical Coordination
			return "144a3478-1946-4460-a965-0d7d74e63d65"
		default:
			// Cognitive Development
			return "0e6b1c2b-5e2f-47e1-8422-2a183f3e15c7"
		}
	case "d896bf1a-fb5b-4a57-b833-87b0959ba926":
		// Knowledge
		switch subCategoryID {
		case "cd06e622-a323-40f3-8409-5384395e00d2",
			"skills40":
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		case "81b09f61-4509-4ce0-b099-c208e62870f9",
			"skills41":
			// Math
			return "81b09f61-4509-4ce0-b099-c208e62870f9"
		case "5b405510-384a-4721-a526-e12b3cbf2092",
			"skills53":
			// Engineering
			return "5b405510-384a-4721-a526-e12b3cbf2092"
		case "9a52fb0a-6ce8-45df-92a0-f25b5d3d2344",
			"5fb241f2993e7591084c8400":
			// Music
			return "9a52fb0a-6ce8-45df-92a0-f25b5d3d2344"
		case "4114f381-a7c5-4e88-be84-2bef4eb04ad0",
			"5fb241e7993e7591084c83fd":
			// Health
			return "4114f381-a7c5-4e88-be84-2bef4eb04ad0"
		case "f4b07251-1d67-4a84-bcda-86c71cbf9cfd",
			"5fb241d2993e7591084c83fa":
			// Social Studies
			return "f4b07251-1d67-4a84-bcda-86c71cbf9cfd"
		case "49e73e4f-8ffc-47e3-9b87-0f9686d361d7",
			"skills52":
			// Technology
			return "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"
		case "852c3495-1ced-4580-a584-9d475217f3d5",
			"5fb241fe993e7591084c8403":
			// Character Education
			return "852c3495-1ced-4580-a584-9d475217f3d5"
		case "3b148168-31d0-4bef-9152-63c3ff516180",
			"skills54":
			// Miscellaneous
			return "3b148168-31d0-4bef-9152-63c3ff516180"
		case "6fb79402-2fb6-4415-874c-338c949332ed",
			"5fb241c7993e7591084c83f7":
			// Art
			return "6fb79402-2fb6-4415-874c-338c949332ed"
		default:
			// Science
			return "cd06e622-a323-40f3-8409-5384395e00d2"
		}
	case "2d5ea951-836c-471e-996e-76823a992689":
		// None Specified
		switch subCategoryID {
		case "40a232cd-d6e8-4ec1-97ec-4e4df7d00a78",
			"skills0":
			// None Specified
			return "40a232cd-d6e8-4ec1-97ec-4e4df7d00a78"
		default:
			// None Specified
			return "40a232cd-d6e8-4ec1-97ec-4e4df7d00a78"
		}
	default:
		// None Specified
		return "40a232cd-d6e8-4ec1-97ec-4e4df7d00a78"
	}
}
