package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

type assessmentLiveRoom struct {
}

func getAssessmentLiveRoom() *assessmentLiveRoom {
	return &assessmentLiveRoom{}
}

//type AssessmentMaterialData struct {
//	LatestID string
//	FileType entity.FileType
//	//Source   SourceID
//}

type RoomUserInfo struct {
	UserID  string
	Results []*RoomUserResults
}
type RoomInfo struct {
	Initialized  bool
	Contents     []*RoomContent
	UserRoomInfo []*RoomUserInfo
}
type RoomContent struct {
	ID             string
	Number         string
	ParentID       string
	SubContentID   string
	SubContentType string
	FileType       external.FileType
	H5PID          string
	MaxScore       float64
	MaterialID     string
	Children       []*RoomContent

	LatestID string
}
type RoomUserResults struct {
	Score  float64
	Answer string
	//MaximumPossibleScore float64
	Seen          bool
	RoomContentID string
}

func (rc *RoomContent) GetID() string {
	if rc.SubContentID == "" {
		if rc.H5PID == "" {
			return rc.MaterialID
		}
		return rc.H5PID
	}
	return rc.SubContentID
}
func (rc *RoomContent) GetParentID() string {
	return rc.ParentID
}
func (rc *RoomContent) AppendChild(item interface{}) {
	rc.Children = append(rc.Children, item.(*RoomContent))
}
func (rc *RoomContent) GetInternalID() string {
	if rc.ParentID == "" {
		return rc.MaterialID
	} else {
		if rc.SubContentID != "" {
			return rc.SubContentID
		} else {
			return rc.H5PID
		}
	}
}

func (m *assessmentLiveRoom) Deconstruct(contents []*RoomContent) []*RoomContent {
	data := make([]TreeEntity, len(contents))
	for i, item := range contents {
		data[i] = item
	}

	treeData := GetTree(data)
	result := make([]*RoomContent, 0)
	for _, item := range treeData {
		result = append(result, item.(*RoomContent))
	}

	return result
}

func (m *assessmentLiveRoom) getRoomResultInfo(ctx context.Context, roomData []*external.H5PUserScores) (*RoomInfo, error) {
	result := &RoomInfo{
		Contents:     make([]*RoomContent, 0),
		UserRoomInfo: make([]*RoomUserInfo, 0),
	}

	if roomData == nil {
		log.Warn(ctx, "room data is null")
		return result, nil
	}

	contentMap := make(map[string]float64)

	for _, item := range roomData {
		if item.User == nil {
			log.Warn(ctx, "room user data is null", log.Any("roomDataItem", item))
			continue
		}

		if len(item.Scores) <= 0 {
			log.Warn(ctx, "room user scores data is null", log.Any("roomDataItem", item))
			continue
		}
		userItem := &RoomUserInfo{
			UserID:  item.User.UserID,
			Results: make([]*RoomUserResults, 0),
		}

		for _, scoreItem := range item.Scores {
			if scoreItem.Content == nil {
				log.Warn(ctx, "room user scores about content data is null", log.Any("roomDataItem", item))
				continue
			}
			contentKey := scoreItem.Content.GetInternalID()
			if _, ok := contentMap[contentKey]; !ok {
				roomContentItem := &RoomContent{
					ID:             contentKey, //scoreItem.Content.GetInternalID(),
					ParentID:       scoreItem.Content.ParentID,
					SubContentID:   scoreItem.Content.SubContentID,
					SubContentType: scoreItem.Content.Type,
					H5PID:          scoreItem.Content.H5PID,
					MaterialID:     scoreItem.Content.ContentID,
					MaxScore:       scoreItem.Score.Max,
					FileType:       scoreItem.Content.FileType,
				}
				result.Contents = append(result.Contents, roomContentItem)
				contentMap[contentKey] = 0
			}

			userResultItem := &RoomUserResults{
				Seen: scoreItem.Seen,
			}
			if scoreItem.Score != nil {
				userResultItem.RoomContentID = contentKey
				if len(scoreItem.TeacherScores) > 0 {
					userResultItem.Score = scoreItem.TeacherScores[len(scoreItem.TeacherScores)-1].Score
				} else if len(scoreItem.Score.Scores) > 0 {
					userResultItem.Score = scoreItem.Score.Scores[0]
				}

				if len(scoreItem.Score.Answers) > 0 {
					userResultItem.Answer = scoreItem.Score.Answers[0].Answer
					if contentMap[contentKey] < scoreItem.Score.Answers[0].MaximumPossibleScore {
						contentMap[contentKey] = scoreItem.Score.Answers[0].MaximumPossibleScore
					}
				}
			}

			userItem.Results = append(userItem.Results, userResultItem)
		}

		result.UserRoomInfo = append(result.UserRoomInfo, userItem)
	}

	for _, item := range result.Contents {
		item.MaxScore = contentMap[item.ID]
	}

	contentTree := getAssessmentLiveRoom().Deconstruct(result.Contents)
	m.setContentNumber(contentTree, "")
	result.Contents = contentTree

	return result, nil
}

func (m *assessmentLiveRoom) calcRoomCompleteRate(roomData []*external.H5PUserScores, studentCount int) float64 {
	attemptedCount := 0
	contentCount := 0
	contentMap := make(map[string]struct{})

	for _, item := range roomData {
		if item.User == nil {
			continue
		}

		if len(item.Scores) <= 0 {
			continue
		}

		for _, scoreItem := range item.Scores {
			if scoreItem.Content == nil {
				continue
			}

			contentKey := scoreItem.Content.GetInternalID()
			if _, ok := contentMap[contentKey]; !ok {
				contentCount++
				contentMap[contentKey] = struct{}{}
			}
			if scoreItem.Seen {
				attemptedCount++
			}

		}
	}

	total := float64(studentCount * contentCount)
	if total > 0 {
		return float64(attemptedCount) / total
	}

	return 0
}

func (m *assessmentLiveRoom) setContentNumber(treedLessonMaterials []*RoomContent, prefix string) {
	for i, lm := range treedLessonMaterials {
		if len(prefix) > 0 {
			lm.Number = fmt.Sprintf("%s-%d", prefix, i+1)
		} else {
			lm.Number = fmt.Sprintf("%d", i+1)
		}
		if len(lm.Children) > 0 {
			m.setContentNumber(lm.Children, lm.Number)
		}
	}
}

func (m *assessmentLiveRoom) batchGetRoomCommentMap(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	commentMap, err := external.GetH5PRoomCommentServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		log.Warn(ctx, "batch get room comment map failed",
			log.Strings("room_ids", roomIDs),
			log.Any("operator", operator),
		)

		return result, nil
	}

	for roomID, users := range commentMap {
		result[roomID] = make(map[string]string, len(users))
		for _, u := range users {
			if len(u.TeacherComments) <= 0 {
				continue
			}

			latestComment := u.TeacherComments[len(u.TeacherComments)-1]

			result[roomID][latestComment.Student.UserID] = latestComment.Comment
		}
	}
	log.Debug(ctx, "batch get room comment map",
		log.Any("result", result),
		log.Strings("room_ids", roomIDs),
	)
	return result, nil
}

var canScoringMap = map[string]bool{
	"Accordion":                    false,
	"AdvancedBlanks":               true,
	"AdventCalendar":               false,
	"Agamotto":                     false,
	"ArithmeticQuiz":               true,
	"Audio":                        false,
	"AudioRecorder":                false,
	"BookMaker":                    false,
	"BranchingScenario":            true,
	"Chart":                        false,
	"Collage":                      false,
	"Column":                       true,
	"CoursePresentationKID":        true,
	"Dialogcards":                  false,
	"Dictation":                    true,
	"DocumentationTool":            false,
	"DragQuestion":                 true,
	"DragText":                     true,
	"Essay":                        true,
	"Blanks":                       true,
	"ImageMultipleHotspotQuestion": true,
	"ImageHotspotQuestion":         true,
	"FindTheWords":                 true,
	"Flashcards":                   true,
	"GuessTheAnswer":               false,
	"IFrameEmbed":                  false,
	"ImageHotspots":                false,
	"ImagePair":                    true,
	"ImageSequencing":              true,
	"ImageJuxtaposition":           false,
	"ImageSlider":                  false,
	"ImpressPresentation":          false,
	"InteractiveBook":              true,
	"InteractiveVideo":             true,
	"JigsawPuzzleKID":              true,
	"KewArCode":                    false,
	"MarkTheWords":                 true,
	"MemoryGame":                   true,
	"MultiChoice":                  true,
	"PersonalityQuiz":              false,
	"Questionnaire":                false,
	"QuestionSet":                  true,
	"Shape":                        false,
	"SingleChoiceSet":              true,
	"SpeakTheWords":                true,
	"SpeqkTheWordsSet":             true,
	"Summary":                      true,
	"Timeline":                     false,
	"TrueFalse":                    true,
	"TwitterUserFeed":              false,
	"ThreeImage":                   false,
}

type TreeEntity interface {
	GetID() string
	GetParentID() string
	AppendChild(item interface{})
}

func GetTree(treeArray []TreeEntity) []TreeEntity {
	tag := make(map[int]bool)
	tag2 := make(map[string]bool)
	result := make([]TreeEntity, 0)

	for i := 0; i < len(treeArray); i++ {
		for j := 0; j < len(treeArray); j++ {
			if !tag[j] && treeArray[i].GetID() == treeArray[j].GetParentID() {
				tag[j] = true
				treeArray[i].AppendChild(treeArray[j])
				tag2[treeArray[j].GetID()] = true
			}
		}
	}

	for _, item := range treeArray {
		if !tag2[item.GetID()] {
			result = append(result, item)
		}
	}
	return result
}