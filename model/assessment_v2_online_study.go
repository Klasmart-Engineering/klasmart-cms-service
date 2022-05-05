package model

import (
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"time"
)

func NewOnlineStudyAssessmentPage(at *AssessmentTool) IAssessmentMatch {
	return &OnlineStudyAssessment{
		at:     at,
		action: AssessmentMatchActionPage,
		base:   NewBaseAssessment(at, AssessmentMatchActionPage),
	}
}

func NewOnlineStudyAssessmentDetail(at *AssessmentTool) IAssessmentMatch {
	return &OnlineStudyAssessment{
		at:     at,
		action: AssessmentMatchActionDetail,
		base:   NewBaseAssessment(at, AssessmentMatchActionDetail),
	}
}

type OnlineStudyAssessment struct {
	EmptyAssessment

	base   BaseAssessment
	at     *AssessmentTool
	action AssessmentMatchAction
}

func (o *OnlineStudyAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	return o.base.MatchSchedule()
}

func (o *OnlineStudyAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	return o.base.MatchLessPlan()
}

func (o *OnlineStudyAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	return o.base.MatchTeacher()
}

func (o *OnlineStudyAssessment) MatchClass() (map[string]*entity.IDName, error) {
	return o.base.MatchClass()
}

func (o *OnlineStudyAssessment) MatchCompleteRate() (map[string]float64, error) {
	ctx := o.at.ctx

	assessmentUserMap, err := o.at.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	studentCount := make(map[string]int)
	for key, users := range assessmentUserMap {
		for _, userItem := range users {
			if userItem.UserType == v2.AssessmentUserTypeStudent {
				studentCount[key]++
			}
		}
	}

	roomDataMap, err := o.at.GetRoomStudentScoresAndComments()
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64)
	for _, item := range o.at.assessments {
		if roomData, ok := roomDataMap[item.ScheduleID]; ok {
			result[item.ID] = GetAssessmentExternalService().calcRoomCompleteRateWhenUseSomeContent(ctx, roomData.ScoresByUser, studentCount[item.ID])
		}
	}

	return result, nil
}

func (o *OnlineStudyAssessment) MatchRemainingTime() (map[string]int64, error) {
	scheduleMap, err := o.at.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, item := range o.at.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			var remainingTime int64
			if schedule.DueAt != 0 {
				remainingTime = schedule.DueAt - time.Now().Unix()
			} else {
				remainingTime = time.Unix(item.CreateAt, 0).Add(config.Get().Assessment.DefaultRemainingTime).Unix() - time.Now().Unix()
			}
			if remainingTime < 0 {
				remainingTime = 0
			}
			result[item.ID] = remainingTime
		}
	}

	return result, nil
}

func (o *OnlineStudyAssessment) MatchOutcomes() (map[string]*v2.AssessmentOutcomeReply, error) {
	onlineClass := NewOnlineClassAssessmentPage(o.at)

	return onlineClass.MatchOutcomes()
}

func (o *OnlineStudyAssessment) MatchContents() ([]*v2.AssessmentContentReply, error) {
	onlineClass := NewOnlineClassAssessmentPage(o.at)

	return onlineClass.MatchContents()
}

func (o *OnlineStudyAssessment) MatchStudents(contentsReply []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error) {
	ctx := o.at.ctx

	assessmentUserMap, err := o.at.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	assessmentUsers, ok := assessmentUserMap[o.at.first.ID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	commentResultMap, err := o.at.FirstGetCommentResultMap()
	if err != nil {
		return nil, err
	}

	assessmentOutcomeMap, err := o.at.FirstGetOutcomeFromAssessment()
	if err != nil {
		return nil, err
	}

	userScoreMap, _, err := o.at.FirstGetRoomData()
	if err != nil {
		return nil, err
	}

	roomUserResultMap := make(map[string]*RoomUserScore)
	for userID, scores := range userScoreMap {
		for _, resultItem := range scores {
			key := o.at.GetKey([]string{
				userID,
				resultItem.ContentUniqueID,
			})
			roomUserResultMap[key] = resultItem
		}
	}

	contentScoreMap, studentScoreMap := o.at.summaryRoomScores(userScoreMap, contentsReply)

	contentMapFromAssessment, err := o.at.FirstGetAssessmentContentMap()
	if err != nil {
		return nil, err
	}

	outcomeMapFromContent, err := o.at.FirstGetOutcomeMapFromContent()
	if err != nil {
		return nil, err
	}

	result := make([]*v2.AssessmentStudentReply, 0, len(assessmentUsers))

	for _, item := range assessmentUsers {
		if item.UserType == v2.AssessmentUserTypeTeacher {
			continue
		}

		studentReply := &v2.AssessmentStudentReply{
			StudentID:     item.UserID,
			StudentName:   "",
			Status:        item.StatusByUser,
			ProcessStatus: item.StatusBySystem,
			Results:       nil,
		}
		studentReply.ReviewerComment = commentResultMap[item.UserID]

		for _, content := range contentsReply {
			resultReply := &v2.AssessmentStudentResultReply{
				ContentID: content.ContentID,
			}

			contentID := content.ContentID
			if content.ContentType == v2.AssessmentContentTypeUnknown {
				contentID = content.ParentID
			}

			var studentContentScore float32
			if contentScoreItem, ok := contentScoreMap[contentID]; ok && contentScoreItem != 0 {
				studentScoreKey := o.at.GetKey([]string{
					item.UserID,
					contentID,
				})
				if studentScoreItem, ok := studentScoreMap[studentScoreKey]; ok {
					studentContentScore = float32(studentScoreItem / contentScoreItem)
				}
			}

			userOutcomeReply := make([]*v2.AssessmentStudentResultOutcomeReply, 0)
			for _, outcomeID := range content.OutcomeIDs {
				var userOutcome *v2.AssessmentUserOutcome
				if assessmentContent, ok := contentMapFromAssessment[contentID]; ok {
					key := o.at.GetKey([]string{
						item.ID,
						assessmentContent.ID,
						outcomeID,
					})
					userOutcome = assessmentOutcomeMap[key]
				}
				userOutcomeReplyItem := &v2.AssessmentStudentResultOutcomeReply{
					OutcomeID: outcomeID,
				}
				if o.at.first.Status == v2.AssessmentStatusInDraft ||
					o.at.first.Status == v2.AssessmentStatusComplete {
					if userOutcome != nil && userOutcome.Status != "" {
						userOutcomeReplyItem.Status = userOutcome.Status
					} else {
						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusUnknown
					}
				} else {
					outcomeInfo, ok := outcomeMapFromContent[outcomeID]
					if !ok {
						log.Warn(ctx, "outcome not found in content", log.Any("outcomeMapFromContent", outcomeMapFromContent), log.String("outcomeID", outcomeID))
						continue
					}
					if outcomeInfo.Assumed || studentContentScore >= outcomeInfo.ScoreThreshold {
						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusAchieved
					} else {
						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusUnknown
					}
				}

				userOutcomeReply = append(userOutcomeReply, userOutcomeReplyItem)
			}
			resultReply.Outcomes = userOutcomeReply

			roomKey := o.at.GetKey([]string{
				item.UserID,
				content.RoomProvideContentID,
			})
			if roomResultItem, ok := roomUserResultMap[roomKey]; ok {
				resultReply.Answer = roomResultItem.Answer
				resultReply.Score = roomResultItem.Score
				resultReply.Attempted = roomResultItem.Seen
			}

			studentReply.Results = append(studentReply.Results, resultReply)
		}

		result = append(result, studentReply)
	}

	return result, nil
}

func (o *OnlineStudyAssessment) MatchAnyOneAttempted() (bool, error) {
	return o.base.MatchAnyOneAttempted()
}
