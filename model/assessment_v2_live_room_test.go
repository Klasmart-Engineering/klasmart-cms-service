package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"testing"
)

func TestGetContent(t *testing.T) {
	//ctx := context.Background()

	jsonData := `
{
  "data": {
    "q13": {
      "scoresByUser": [
        {
          "user": {
            "user_id": "4831e78a-9a2f-45eb-8ca0-743642e674dc",
            "given_name": "syu101",
            "family_name": "BBs"
          },
          "scores": [
            {
              "seen": false,
              "content": {
                "parent_id": null,
                "content_id": "614d7854bbbecb046db012e1",
                "name": "eee",
                "type": "MemoryGame",
                "fileType": "H5P",
                "h5p_id": "614d7851eb70fa0015da40f9",
                "subcontent_id": null
              },
              "score": {
                "min": null,
                "max": null,
                "sum": 0,
                "scoreFrequency": 0,
                "mean": null,
                "scores": [],
                "answers": [],
                "median": null,
                "medians": []
              },
              "teacherScores": []
            },
            {
              "seen": false,
              "content": {
                "parent_id": null,
                "content_id": "613088f76115dbc269a1cff3",
                "name": "一样",
                "type": "ImageSequencing",
                "fileType": "H5P",
                "h5p_id": "60ffb65824e6bb0013166b40",
                "subcontent_id": null
              },
              "score": {
                "min": null,
                "max": null,
                "sum": 0,
                "scoreFrequency": 0,
                "mean": null,
                "scores": [],
                "answers": [],
                "median": null,
                "medians": []
              },
              "teacherScores": []
            }
          ]
        },
        {
          "user": {
            "user_id": "4a80a49e-814f-41ff-9a4c-f105f4f2eb85",
            "given_name": "fteach092401",
            "family_name": "FT"
          },
          "scores": [
            {
              "seen": false,
              "content": {
                "parent_id": null,
                "content_id": "614d7854bbbecb046db012e1",
                "name": "eee",
                "type": "MemoryGame",
                "fileType": "H5P",
                "h5p_id": "614d7851eb70fa0015da40f9",
                "subcontent_id": null
              },
              "score": {
                "min": null,
                "max": null,
                "sum": 0,
                "scoreFrequency": 0,
                "mean": null,
                "scores": [],
                "answers": [],
                "median": null,
                "medians": []
              },
              "teacherScores": []
            },
            {
              "seen": false,
              "content": {
                "parent_id": null,
                "content_id": "613088f76115dbc269a1cff3",
                "name": "一样",
                "type": "ImageSequencing",
                "fileType": "H5P",
                "h5p_id": "60ffb65824e6bb0013166b40",
                "subcontent_id": null
              },
              "score": {
                "min": null,
                "max": null,
                "sum": 0,
                "scoreFrequency": 0,
                "mean": null,
                "scores": [],
                "answers": [],
                "median": null,
                "medians": []
              },
              "teacherScores": []
            }
          ]
        },
        {
          "user": {
            "user_id": "aea0e494-e56f-417e-99a7-81774c879bf8",
            "given_name": "st01",
            "family_name": "Zy"
          },
          "scores": [
            {
              "seen": true,
              "content": {
                "parent_id": null,
                "content_id": "614d7854bbbecb046db012e1",
                "name": "eee",
                "type": "MemoryGame",
                "fileType": "H5P",
                "h5p_id": "614d7851eb70fa0015da40f9",
                "subcontent_id": null
              },
              "score": {
                "min": 1,
                "max": 1,
                "sum": 1,
                "scoreFrequency": 1,
                "mean": 1,
                "scores": [
                  1
                ],
                "answers": [
                  {
                    "answer": null,
                    "score": 1,
                    "date": 1648618621708,
                    "minimumPossibleScore": 0,
                    "maximumPossibleScore": 1
                  }
                ],
                "median": 1,
                "medians": [
                  1
                ]
              },
              "teacherScores": []
            },
            {
              "seen": true,
              "content": {
                "parent_id": null,
                "content_id": "613088f76115dbc269a1cff3",
                "name": "一样",
                "type": "ImageSequencing",
                "fileType": "H5P",
                "h5p_id": "60ffb65824e6bb0013166b40",
                "subcontent_id": null
              },
              "score": {
                "min": 3,
                "max": 3,
                "sum": 3,
                "scoreFrequency": 1,
                "mean": 3,
                "scores": [
                  3
                ],
                "answers": [
                  {
                    "answer": "item_0[,]item_5[,]item_2[,]item_3[,]item_1[,]item_4",
                    "score": 3,
                    "date": 1648618800055,
                    "minimumPossibleScore": 0,
                    "maximumPossibleScore": 6
                  }
                ],
                "median": 3,
                "medians": [
                  3
                ]
              },
              "teacherScores": []
            },
            {
              "seen": true,
              "content": {
                "parent_id": null,
                "content_id": "61275b9059a4a7c5ec3fea96",
                "name": "1111",
                "type": "InteractiveBook",
                "fileType": "H5P",
                "h5p_id": "611c81feb945e50013448d47",
                "subcontent_id": null
              },
              "score": {
                "min": null,
                "max": null,
                "sum": 0,
                "scoreFrequency": 0,
                "mean": null,
                "scores": [],
                "answers": [],
                "median": null,
                "medians": []
              },
              "teacherScores": []
            },
            {
              "seen": true,
              "content": {
                "parent_id": null,
                "content_id": "60ffb32fef0a8888bc28da78",
                "name": "1",
                "type": "Column",
                "fileType": "H5P",
                "h5p_id": "60ffb30f05fb6500132c02a5",
                "subcontent_id": null
              },
              "score": {
                "min": 1,
                "max": 1,
                "sum": 1,
                "scoreFrequency": 1,
                "mean": 1,
                "scores": [
                  1
                ],
                "answers": [
                  {
                    "answer": null,
                    "score": 1,
                    "date": 1648618825088,
                    "minimumPossibleScore": 0,
                    "maximumPossibleScore": 4
                  }
                ],
                "median": 1,
                "medians": [
                  1
                ]
              },
              "teacherScores": []
            },
            {
              "seen": true,
              "content": {
                "parent_id": "60ffb30f05fb6500132c02a5",
                "content_id": "60ffb32fef0a8888bc28da78",
                "name": "colum",
                "type": "MemoryGame",
                "fileType": "H5P",
                "h5p_id": "60ffb30f05fb6500132c02a5",
                "subcontent_id": "892960a0-0482-4888-a2df-21f91336a5f7"
              },
              "score": {
                "min": 1,
                "max": 1,
                "sum": 1,
                "scoreFrequency": 1,
                "mean": 1,
                "scores": [
                  1
                ],
                "answers": [
                  {
                    "answer": null,
                    "score": 1,
                    "date": 1648618825065,
                    "minimumPossibleScore": 0,
                    "maximumPossibleScore": 1
                  }
                ],
                "median": 1,
                "medians": [
                  1
                ]
              },
              "teacherScores": []
            },
            {
              "seen": true,
              "content": {
                "parent_id": "60ffb30f05fb6500132c02a5",
                "content_id": "60ffb32fef0a8888bc28da78",
                "name": "1",
                "type": "TrueFalse",
                "fileType": "H5P",
                "h5p_id": "60ffb30f05fb6500132c02a5",
                "subcontent_id": "78fffe52-fdb6-4dfa-8224-5fd0e4764bf9"
              },
              "score": {
                "min": 0,
                "max": 0,
                "sum": 0,
                "scoreFrequency": 1,
                "mean": 0,
                "scores": [
                  0
                ],
                "answers": [
                  {
                    "answer": "false",
                    "score": 0,
                    "date": 1648618818098,
                    "minimumPossibleScore": 0,
                    "maximumPossibleScore": 1
                  }
                ],
                "median": 0,
                "medians": [
                  0
                ]
              },
              "teacherScores": []
            },
            {
              "seen": true,
              "content": {
                "parent_id": "60ffb30f05fb6500132c02a5",
                "content_id": "60ffb32fef0a8888bc28da78",
                "name": "Essay",
                "type": "Essay",
                "fileType": "H5P",
                "h5p_id": "60ffb30f05fb6500132c02a5",
                "subcontent_id": "2e08cc63-9d5b-4a56-afd9-f7dea457b040"
              },
              "score": {
                "min": 0,
                "max": 0,
                "sum": 0,
                "scoreFrequency": 1,
                "mean": 0,
                "scores": [
                  0
                ],
                "answers": [
                  {
                    "answer": "kkkkkk",
                    "score": 0,
                    "date": 1648618821713,
                    "minimumPossibleScore": 0,
                    "maximumPossibleScore": 2
                  }
                ],
                "median": 0,
                "medians": [
                  0
                ]
              },
              "teacherScores": []
            },
            {
              "seen": true,
              "content": {
                "parent_id": null,
                "content_id": "6215f29911d9e9fd5634e369",
                "name": "Essay",
                "type": "Essay",
                "fileType": "H5P",
                "h5p_id": "60ffb44d24e6bb0013166b3f",
                "subcontent_id": null
              },
              "score": {
                "min": 0,
                "max": 0,
                "sum": 0,
                "scoreFrequency": 1,
                "mean": 0,
                "scores": [
                  0
                ],
                "answers": [
                  {
                    "answer": "kkkkkkkkkkkkkkkkkkkkkkkhhhhhhhhhhhhhhhhhhhhhhhhh",
                    "score": 0,
                    "date": 1648618833965,
                    "minimumPossibleScore": 0,
                    "maximumPossibleScore": 2
                  }
                ],
                "median": 0,
                "medians": [
                  0
                ]
              },
              "teacherScores": []
            },
            {
              "seen": false,
              "content": {
                "parent_id": null,
                "content_id": "60ffb20a27e9d08f148fbfe0",
                "name": "Audio",
                "type": "Audio",
                "fileType": "Audio",
                "h5p_id": null,
                "subcontent_id": null
              },
              "score": {
                "min": null,
                "max": null,
                "sum": 0,
                "scoreFrequency": 0,
                "mean": null,
                "scores": [],
                "answers": [],
                "median": null,
                "medians": []
              },
              "teacherScores": []
            }
          ]
        },
        {
          "user": {
            "user_id": "c1abdb68-5ca2-401d-a4ac-a845c05e9a12",
            "given_name": "student2021001",
            "family_name": "BBS KENIV JAM"
          },
          "scores": [
            {
              "seen": false,
              "content": {
                "parent_id": null,
                "content_id": "614d7854bbbecb046db012e1",
                "name": "eee",
                "type": "MemoryGame",
                "fileType": "H5P",
                "h5p_id": "614d7851eb70fa0015da40f9",
                "subcontent_id": null
              },
              "score": {
                "min": null,
                "max": null,
                "sum": 0,
                "scoreFrequency": 0,
                "mean": null,
                "scores": [],
                "answers": [],
                "median": null,
                "medians": []
              },
              "teacherScores": []
            }
          ]
        }
      ]
    }
  }
}
`
	t.Log(jsonData)
}

func TestAssessmentExternalService_ContentsToTree(t *testing.T) {
	ctx := context.Background()
	testOperator := &entity.Operator{
		UserID: "afdfc0d9-ada9-4e66-b225-20f956d1a399",
		OrgID:  "60c064cc-bbd8-4724-b3f6-b886dce4774f", // Badanamu HQ
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImQ2NTNmODgwLWI5MjAtNDM1Zi04ZjJkLTk4YjVkNDYyMWViOCIsImVtYWlsIjoic2Nob29sXzAzMDMyOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0OTA3OTI4MCwiaXNzIjoia2lkc2xvb3AifQ.uk1nBxFRcFVU20dwN5uVS3_4Oot6Jktppup-sEvOuye0Jf3_hZ4Do6H8_bsLpCTTpM4fKOididI9NZCtAZUxKQGB8-d2nEJd_wr5U-QE2tyOgCAPwcftP9Ra9J8jhDQGz30YuVO_-ieEnHcTxMaINIfaM0DUEpSgzLxcnn83xBFTTvGfT4CRGx5npfKoYMBDXqaFnUSfrHovLSc5cDvsoDveZ5xUEY4oy99Yc5MuPCmXdxTbygPdCiUn2dvUwUe5xWxC9kgk_4kJZsE8qbs9MQ1V4kK1jebpw9G6_O7fdldv2b5Aqh6lHDb2C8wEXjDCnu7U_RUf94foLXxeYtCmMQ",
	}
	roomInfo, err := external.GetAssessmentServiceProvider().GetScoresWithCommentsByRoomIDs(ctx, testOperator, []string{"62454aee4c6e70e130530dbc"})
	if err != nil {
		t.Error(err)
	}
	userScoresMap, contentTree, err := GetAssessmentExternalService().StudentScores(ctx, roomInfo["62454aee4c6e70e130530dbc"].ScoresByUser)
	if err != nil {
		t.Error(err)
	}
	t.Log(userScoresMap)
	t.Log(contentTree)
}
