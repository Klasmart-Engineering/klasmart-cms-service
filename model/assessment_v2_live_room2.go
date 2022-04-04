package model

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

type AssessmentExternalService struct {
}

func GetAssessmentExternalService() *AssessmentExternalService {
	return &AssessmentExternalService{}
}

type RoomContentTree struct {
	TreeID       string
	TreeParentID string
	Number       string

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
	return fmt.Sprintf("%s:%s", content.ContentID, content.ParentID)
}

func (aes *AssessmentExternalService) StudentCommonContentsToTree(ctx context.Context, roomContents []*external.ScoresByContent) ([]*RoomContentTree, error) {
	result := make([]*RoomContentTree, 0, len(roomContents))
	for _, item := range roomContents {
		if item.Content == nil {
			continue
		}
		resultItem := &RoomContentTree{
			H5PContent: external.H5PContent{
				ParentID:     item.Content.ParentID,
				ContentID:    item.Content.ContentID,
				Name:         item.Content.Name,
				Type:         item.Content.Type,
				FileType:     item.Content.FileType,
				H5PID:        item.Content.H5PID,
				SubContentID: item.Content.SubContentID,
			},
			Children: nil,
		}

		resultItem.TreeID = aes.ParseTreeID(item.Content)
		resultItem.TreeParentID = aes.ParseParentID(item.Content)

		if resultItem.TreeID == "" {
			log.Warn(ctx, "content tree id is empty", log.Any("roomItem", item))
			continue
		}

		result = append(result, resultItem)
	}

	tree := aes.deconstructUserRoomInfo(result)
	aes.setUserRoomItemNumber(tree, "")

	return result, nil
}

type RoomUserScore struct {
	ContentTreeID string
	Score         float64
	Seen          bool
	Answer        string
}

func (aes *AssessmentExternalService) StudentScoresMap(ctx context.Context, userScores []*external.H5PUserScores) (map[string][]*RoomUserScore, map[string]float64, error) {
	userScoreMap := make(map[string][]*RoomUserScore, len(userScores))
	contentMaxScoreMap := make(map[string]float64)

	for _, userScoreItem := range userScores {
		if userScoreItem.User == nil {
			log.Warn(ctx, "user is nil", log.Any("userScoreItem", userScoreItem))
			continue
		}
		if len(userScoreItem.Scores) <= 0 {
			continue
		}

		for _, scoreItem := range userScoreItem.Scores {
			if scoreItem.Score == nil {
				log.Warn(ctx, "user score item is nil", log.String("userID", userScoreItem.User.UserID), log.Any("scoreItem", scoreItem))
				continue
			}

			if scoreItem.Content == nil {
				log.Warn(ctx, "user content is nil", log.String("userID", userScoreItem.User.UserID), log.Any("scoreItem", scoreItem))
				continue
			}

			userScoreResultItem := &RoomUserScore{
				ContentTreeID: aes.ParseTreeID(scoreItem.Content),
				Seen:          scoreItem.Seen,
			}

			if len(scoreItem.TeacherScores) > 0 {
				userScoreResultItem.Score = scoreItem.TeacherScores[len(scoreItem.TeacherScores)-1].Score
			} else if len(scoreItem.Score.Scores) > 0 {
				userScoreResultItem.Score = scoreItem.Score.Scores[0]
			} else {
				userScoreResultItem.Score = 0
			}

			if len(scoreItem.Score.Answers) > 0 {
				userScoreResultItem.Answer = scoreItem.Score.Answers[0].Answer
				if contentMaxScoreMap[userScoreResultItem.ContentTreeID] < scoreItem.Score.Answers[0].MaximumPossibleScore {
					contentMaxScoreMap[userScoreResultItem.ContentTreeID] = scoreItem.Score.Answers[0].MaximumPossibleScore
				}
			}

			userScoreMap[userScoreItem.User.UserID] = append(userScoreMap[userScoreItem.User.UserID], userScoreResultItem)
		}
	}

	return userScoreMap, contentMaxScoreMap, nil
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

func (aes *AssessmentExternalService) calcRoomCompleteRate(ctx context.Context, userScore map[string][]*RoomUserScore, studentCount int, contentCount int) float64 {
	attemptedCount := 0

	for _, scores := range userScore {
		for _, scoreItem := range scores {
			if scoreItem.Seen {
				attemptedCount++
			}
		}
	}

	var result float64

	total := float64(studentCount * contentCount)
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
