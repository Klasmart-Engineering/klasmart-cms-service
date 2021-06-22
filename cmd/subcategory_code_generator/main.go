package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

var (
	testOperator = &entity.Operator{
		UserID: "229c3c0a-20f4-5ccf-b6f7-3f1f51296cb9",
		OrgID:  "740ec808-bd56-46c6-8bcb-babbe1666dc4", // Badanamu HQ
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImYyNjI2YTIxLTNlOTgtNTE3ZC1hYzRhLWVkNmYzMzIzMTg2OSIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYyMDc5MTU4NSwiaXNzIjoia2lkc2xvb3AifQ.JC0jAcFYjgOauUmxeCWIB1yiXQ3u4gg4bNNhHp58sOPHScTL5IadXrrh_hAo-6x-mQyRKV86TrAH8Z-KCP6Rm2fybmY5bIUwUi4AGsPjn8CD40kZXSDDGn4yquHLmvl1NFQAupLCnYfq91BJu4F3EwWKNFC8nMXE7VTTWMouy-J_cgBTElSzV1G-WHTe4dSx3mcr1p6OSBP5UyJMqg3DH55Vxe8keCacLP9yz5gtSoCnmBoX4Cn8Uwe1i1YIahQn0ssxgnTnUsUF6v2UPdk2gOSZDTKxdGJxLJV3-cQhQw-uU0LkkMkssYSTzGGIl2UYbdYAg0Cwo0k9XcAomaSyWw",
	}
	oldDict = map[string]string{
		"Academic Skill (Drawing, Tracing, Coloring, Writing)": "skills22",
		"Art":                     "5fb241c7993e7591084c83f7",
		"Body Coordination":       "skills24",
		"Character Education":     "5fb241fe993e7591084c8403",
		"Click":                   "5fb243f9993e7591084c8457",
		"Coding":                  "skills42",
		"Cognitive Development":   "skills50",
		"Coloring":                "5fb243e1993e7591084c844e",
		"Colors":                  "skills37",
		"Communication":           "skills46",
		"Complex Movements":       "skills10",
		"Comprehension":           "5fb24df3993e7591084c85a2",
		"Counting and Operations": "skills26",
		"Critical Thinking (Interpretation, Analysis, Evaluation, Inference, Explanation, and Self-Regulation)": "skills35",
		"Drag":                            "5fb243f4993e7591084c8454",
		"Drawing":                         "5fb243d2993e7591084c8448",
		"Emergent Reading":                "skills47",
		"Emergent Writing":                "skills48",
		"Emotional Skills":                "skills15",
		"Empathy":                         "skills30",
		"Engineering":                     "skills53",
		"Experimenting & Problem Solving": "skills43",
		"Fluency":                         "skills21",
		"Hand-Eye Coordination":           "skills8",
		"Health":                          "5fb241e7993e7591084c83fd",
		"Interpreting":                    "skills19",
		"Language Support":                "skills45",
		"Letters":                         "skills39",
		"Listening":                       "skills17",
		"Logic & Memory":                  "skills12",
		"Logical Problem-Solving":         "skills27",
		"Math":                            "skills41",
		"Miscellaneous":                   "skills54",
		"Music":                           "5fb241f2993e7591084c8400",
		"None Specified":                  "skills0",
		"Numbers":                         "skills20",
		"Patterns":                        "skills28",
		"Phonemic Awareness, Phonics, and Word Recognition": "5fb24c82993e7591084c856d",
		"Phonics":                                  "skills2",
		"Phonological Awareness":                   "skills44",
		"Physical Coordination":                    "skills51",
		"Physical Skills":                          "skills11",
		"Play Skill (Drag and Drop, Screen Click)": "skills23",
		"Reading":                                  "5fb242d6993e7591084c842d",
		"Reading Skills and Comprehension":         "skills5",
		"Reasoning":                                "skills16",
		"Reasoning Skills":                         "skills36",
		"Science":                                  "skills40",
		"Science Process (Observing, Classifying, Communicating, Measuring, Predicting)": "skills34",
		"Self-Control":              "skills32",
		"Self-Identity":             "skills31",
		"Sensory":                   "skills7",
		"Shapes":                    "skills38",
		"Sight Words":               "skills6",
		"Simple Movements":          "skills9",
		"Social Interactions":       "skills29",
		"Social Skills":             "skills14",
		"Social Studies":            "5fb241d2993e7591084c83fa",
		"Social-Emotional Learning": "skills49",
		"Spatial Representation":    "skills25",
		"Speaking":                  "skills18",
		"Speaking & Listening":      "skills1",
		"Technology":                "skills52",
		"Thematic Concepts":         "skills4",
		"Tracing":                   "5fb243db993e7591084c844b",
		"Visual":                    "skills13",
		"Vocabulary":                "skills3",
		"Writing":                   "skills33",
	}

	programIDs = []string{
		"75004121-0c0d-486c-ba65-4c57deacb44b",
		"14d350f1-a7ba-4f46-bef9-dc847f0cbac5",
		"04c630cc-fabe-4176-80f2-30a029907a33",
		"f6617737-5022-478d-9672-0354667e0338",
		"4591423a-2619-4ef8-a900-f5d924939d02",
		"d1bbdcc5-0d80-46b0-b98e-162e7439058f",
		"b39edb9a-ab91-4245-94a4-eb2b5007c033",
		"7a8c5021-142b-44b1-b60b-275c29d132fe",
		"56e24fa0-e139-4c80-b365-61c9bc42cd3f",
		"93f293e8-2c6a-47ad-bc46-1554caac99e4",
		"7565ae11-8130-4b7d-ac24-1d9dd6f792f2",
	}
)

func main() {
	config.Set(&config.Config{
		AMS: config.AMSConfig{
			EndPoint: "https://api.alpha.kidsloop.net/user/",
		},
	})

	ctx := context.Background()
	sub, err := external.GetSubCategoryServiceProvider().GetByCategory(ctx, testOperator, "84b8f87a-7b61-4580-a190-a9ce3fe90dd3")
	if err != nil {
		log.Fatalf("query failed due to %v", err)
	}

	log.Printf("query success: %#v", sub)

	generate(ctx, programIDs, oldDict)
}

func generate(ctx context.Context, programIDs []string, dict map[string]string) error {
	file, _ := os.Create("code.txt")
	defer file.Close()

	for _, programID := range programIDs {
		categories, err := external.GetCategoryServiceProvider().GetByProgram(ctx, testOperator, programID, external.WithStatus(external.Active))
		if err != nil {
			return err
		}

		for _, category := range categories {
			subCategories, err := external.GetSubCategoryServiceProvider().GetByCategory(ctx, testOperator, category.ID, external.WithStatus(external.Active))
			if err != nil {
				return err
			}

			if len(subCategories) == 0 {
				continue
			}

			fmt.Fprintf(file, "case \"%s\":\n", category.ID)
			fmt.Fprintf(file, "\t// %s\n", category.Name)

			fmt.Fprintf(file, "\tswitch subCategoryID {\n")
			for index, subCategory := range subCategories {
				fmt.Fprintf(file, "\tcase \"%s\"", subCategory.ID)

				oldID, found := dict[subCategory.Name]
				if found {
					fmt.Fprintf(file, ",\n\t\t\"%s\":\n", oldID)
				} else {
					fmt.Fprint(file, ":\n")
				}

				fmt.Fprintf(file, "\t\t// %s\n", subCategory.Name)
				fmt.Fprintf(file, "\t\treturn \"%s\"\n", subCategory.ID)

				if index == len(subCategories)-1 {
					fmt.Fprint(file, "\tdefault:\n")
					fmt.Fprintf(file, "\t\t// %s\n", subCategories[0].Name)
					fmt.Fprintf(file, "\t\treturn \"%s\"\n", subCategories[0].ID)
				}
			}
			fmt.Fprint(file, "\t}\n")
		}
	}

	return nil
}
