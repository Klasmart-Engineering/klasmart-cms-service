package model

import (
	"context"
	"fmt"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
)

type AssessmentExternalService struct {
}

func GetAssessmentExternalService() *AssessmentExternalService {
	return &AssessmentExternalService{}
}

var canSetScoreMap = map[string]bool{
	"Essay":         true,
	"AudioRecorder": true,
	"SpeakTheWords": true,
}

type RoomContentTree struct {
	TreeID          string
	TreeParentID    string
	ContentUniqueID string

	Number   string
	MaxScore float64
	//LatestID     string

	external.H5PContent
	Children []*RoomContentTree
}

func (rc *RoomContentTree) GetID() string {
	return rc.TreeID
}
func (rc *RoomContentTree) GetParentID() string {
	return rc.TreeParentID
}
func (rc *RoomContentTree) AppendChild(item interface{}) {
	rc.Children = append(rc.Children, item.(*RoomContentTree))
}

func (aes *AssessmentExternalService) ParseTreeID(content *external.H5PContent) string {
	if content.ParentID == "" {
		return fmt.Sprintf("%s:%s", content.ContentID, content.H5PID)
	} else {
		return fmt.Sprintf("%s:%s", content.ContentID, content.SubContentID)
	}
}

func (aes *AssessmentExternalService) ParseParentID(content *external.H5PContent) string {
	if content.ParentID != "" {
		return fmt.Sprintf("%s:%s", content.ContentID, content.ParentID)
	}

	return content.ParentID
}

func (aes *AssessmentExternalService) ParseContentUniqueID(content *external.H5PContent) string {
	if content.ParentID == "" {
		return content.ContentID
	} else {
		return fmt.Sprintf("%s:%s", content.ContentID, content.SubContentID)
	}
}

type RoomUserScore struct {
	ContentUniqueID string
	Score           float64
	Seen            bool
	Answer          string
}

func (aes *AssessmentExternalService) StudentScores(ctx context.Context, userScores []*external.H5PUserScores) (map[string][]*RoomUserScore, []*RoomContentTree, error) {
	scoresData, err := aes.processStudentScoresData(ctx, userScores, true)
	if err != nil {
		return nil, nil, err
	}

	return scoresData.UserScores, scoresData.ContentTree, nil
}

type processStudentScoresOutput struct {
	UserScores  map[string][]*RoomUserScore
	ContentTree []*RoomContentTree
	ContentMap  map[string]*RoomContentTree
}

func (aes *AssessmentExternalService) processStudentScoresData(ctx context.Context, userScores []*external.H5PUserScores, isNeedContentTree bool) (*processStudentScoresOutput, error) {
	userScoreMap := make(map[string][]*RoomUserScore, len(userScores))
	contentMaxScoreMap := make(map[string]float64)
	contents := make([]*RoomContentTree, 0)

	for _, userScoreItem := range userScores {
		if userScoreItem.User == nil {
			log.Warn(ctx, "user is nil", log.Any("userScoreItem", userScoreItem))
			continue
		}
		if len(userScoreItem.Scores) <= 0 {
			continue
		}

		// content id
		userContentScoreMap := make(map[string]*RoomUserScore)
		for _, scoreItem := range userScoreItem.Scores {
			if scoreItem.Score == nil {
				log.Warn(ctx, "user score item is nil", log.String("userID", userScoreItem.User.UserID), log.Any("scoreItem", scoreItem))
				continue
			}

			if scoreItem.Content == nil {
				log.Warn(ctx, "user content is nil", log.String("userID", userScoreItem.User.UserID), log.Any("scoreItem", scoreItem))
				continue
			}

			contentUniqueID := aes.ParseContentUniqueID(scoreItem.Content)

			if _, ok := contentMaxScoreMap[contentUniqueID]; !ok {
				resultItem := &RoomContentTree{
					TreeID:          aes.ParseTreeID(scoreItem.Content),
					TreeParentID:    aes.ParseParentID(scoreItem.Content),
					ContentUniqueID: aes.ParseContentUniqueID(scoreItem.Content),
					H5PContent: external.H5PContent{
						ParentID:     scoreItem.Content.ParentID,
						ContentID:    scoreItem.Content.ContentID,
						Name:         scoreItem.Content.Name,
						Type:         scoreItem.Content.Type,
						FileType:     scoreItem.Content.FileType,
						H5PID:        scoreItem.Content.H5PID,
						SubContentID: scoreItem.Content.SubContentID,
					},
					Children: nil,
				}
				contents = append(contents, resultItem)
				contentMaxScoreMap[contentUniqueID] = 0
			}

			if userContentScoreItem, ok := userContentScoreMap[contentUniqueID]; ok {
				if !userContentScoreItem.Seen {
					aes.setStudentScore(userContentScoreItem, scoreItem, contentMaxScoreMap)
				}
			} else {
				userScoreResultItem := &RoomUserScore{
					ContentUniqueID: contentUniqueID,
				}
				aes.setStudentScore(userScoreResultItem, scoreItem, contentMaxScoreMap)

				userScoreMap[userScoreItem.User.UserID] = append(userScoreMap[userScoreItem.User.UserID], userScoreResultItem)
				userContentScoreMap[contentUniqueID] = userScoreResultItem
			}
		}
	}

	result := &processStudentScoresOutput{
		UserScores: userScoreMap,
		ContentMap: make(map[string]*RoomContentTree),
	}

	for _, item := range contents {
		item.MaxScore = contentMaxScoreMap[item.ContentUniqueID]
		result.ContentMap[item.ContentUniqueID] = item
	}

	if isNeedContentTree {
		result.ContentTree = aes.deconstructUserRoomInfo(contents)
	}

	return result, nil
}

func (aes *AssessmentExternalService) setStudentScore(userScoreResultItem *RoomUserScore, scoreItem *external.H5PUserContentScore, contentMaxScoreMap map[string]float64) {
	if scoreItem.Seen {
		userScoreResultItem.Seen = true
		if len(scoreItem.TeacherScores) > 0 {
			userScoreResultItem.Score = scoreItem.TeacherScores[len(scoreItem.TeacherScores)-1].Score
		} else if len(scoreItem.Score.Scores) > 0 {
			userScoreResultItem.Score = scoreItem.Score.Scores[0]
		} else {
			userScoreResultItem.Score = 0
		}
		if len(scoreItem.Score.Answers) > 0 {
			userScoreResultItem.Answer = scoreItem.Score.Answers[0].Answer
			if contentMaxScoreMap[userScoreResultItem.ContentUniqueID] < scoreItem.Score.Answers[0].MaximumPossibleScore {
				contentMaxScoreMap[userScoreResultItem.ContentUniqueID] = scoreItem.Score.Answers[0].MaximumPossibleScore
			}
		}
	}
}

func (aes *AssessmentExternalService) StudentCommentMap(ctx context.Context, teacherComments []*external.H5PTeacherCommentsByStudent) (map[string]string, error) {
	result := make(map[string]string)

	for _, commentItem := range teacherComments {
		if commentItem.User == nil {
			log.Warn(ctx, "user is nil", log.Any("commentItem", commentItem))
			continue
		}

		if len(commentItem.TeacherComments) <= 0 {
			continue
		}

		latestComment := commentItem.TeacherComments[len(commentItem.TeacherComments)-1]

		if latestComment == nil {
			continue
		}

		result[commentItem.User.UserID] = latestComment.Comment
	}

	return result, nil
}

func (aes *AssessmentExternalService) setUserRoomItemNumber(userRoomInfos []*RoomContentTree, prefix string) {
	for i, lm := range userRoomInfos {
		if len(prefix) > 0 {
			lm.Number = fmt.Sprintf("%s-%d", prefix, i+1)
		} else {
			lm.Number = fmt.Sprintf("%d", i+1)
		}
		if len(lm.Children) > 0 {
			aes.setUserRoomItemNumber(lm.Children, lm.Number)
		}
	}
}

func (aes *AssessmentExternalService) deconstructUserRoomInfo(userRoomInfos []*RoomContentTree) []*RoomContentTree {
	data := make([]TreeEntity, len(userRoomInfos))
	for i, item := range userRoomInfos {
		data[i] = item
	}

	treeData := GetTree(data)
	result := make([]*RoomContentTree, 0)
	for _, item := range treeData {
		result = append(result, item.(*RoomContentTree))
	}

	return result
}

func (aes *AssessmentExternalService) calcRoomCompleteRateWhenUseSomeContent(ctx context.Context, userScores []*external.H5PUserScores, studentCount int) float64 {
	contentCount := 0
	contentMap := make(map[string]struct{})
	attemptedMap := make(map[string]struct{})
	for _, item := range userScores {
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

			contentKey := aes.ParseContentUniqueID(scoreItem.Content)
			if _, ok := contentMap[contentKey]; !ok {
				contentCount++
				contentMap[contentKey] = struct{}{}
			}

			if scoreItem.Seen {
				userContentKey := fmt.Sprintf("%s:%s", item.User.UserID, contentKey)
				attemptedMap[userContentKey] = struct{}{}
			}
		}
	}

	var result float64

	total := float64(studentCount * contentCount)
	attemptedCount := len(attemptedMap)
	if total > 0 {
		result = float64(attemptedCount) / total

		if result > 1 {
			log.Warn(ctx, "calcRoomCompleteRate greater than 1",
				log.Int("studentCount", studentCount),
				log.Int("contentCount", contentCount),
				log.Int("attemptedCount", attemptedCount),
			)

			result = 1
		}
	}

	return result
}

func (aes *AssessmentExternalService) calcRoomCompleteRateWhenUseDiffContent(ctx context.Context, userScores []*external.H5PUserScores, contentTotalCount int) float64 {
	attemptedCount := 0
	childCount := 0
	for _, item := range userScores {
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

			// At present, the review types are all h5p types
			if scoreItem.Content.FileType != external.FileTypeH5P {
				continue
			}

			if scoreItem.Seen {
				attemptedCount++
			}

			if scoreItem.Content.ParentID != "" {
				childCount++
			}
		}
	}

	var result float64

	// number of attempts /（parent count + child count）
	if contentTotalCount > 0 {
		result = float64(attemptedCount) / float64(contentTotalCount+childCount)

		if result > 1 {
			log.Warn(ctx, "calcRoomCompleteRate greater than 1",
				log.Int("contentTotalCount", contentTotalCount),
				log.Int("attemptedCount", attemptedCount),
			)

			result = 1
		}
	}

	return result
}

type AllowEditScoreContent struct {
	ContentID    string
	SubContentID string
	Attempted    bool
}

func (aes *AssessmentExternalService) AllowEditScoreContent(ctx context.Context, userScores []*external.H5PUserScores) (map[string]map[string]*AllowEditScoreContent, error) {
	scoresData, err := aes.processStudentScoresData(ctx, userScores, false)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]*AllowEditScoreContent)
	for stuID, stuScores := range scoresData.UserScores {
		result[stuID] = make(map[string]*AllowEditScoreContent)
		for _, scoreItem := range stuScores {
			contentItem, ok := scoresData.ContentMap[scoreItem.ContentUniqueID]
			if !ok {
				continue
			}

			if contentItem.MaxScore > 0 && canSetScoreMap[contentItem.Type] {
				result[stuID][scoreItem.ContentUniqueID] = &AllowEditScoreContent{
					ContentID:    contentItem.ContentID,
					SubContentID: contentItem.SubContentID,
				}
			}

		}
	}

	log.Debug(ctx, "can set score info", log.Any("result", result))
	return result, nil
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
