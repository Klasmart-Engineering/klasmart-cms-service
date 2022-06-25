package model

import (
	"context"
	"fmt"
	"github.com/KL-Engineering/kidsloop-cms-service/external"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
)

func NewOfflineClassAssessment() IAssessmentProcessor {
	return &OfflineClassAssessment{}
}

type OfflineClassAssessment struct{}

func (o *OfflineClassAssessment) ProcessCompleteRate(ctx context.Context, assessmentUsers []*v2.AssessmentUser, roomData *external.RoomInfo, stuReviewMap map[string]*entity.ScheduleReview, reviewerFeedbackMap map[string]*v2.AssessmentReviewerFeedback) float64 {
	return 0
}

func (o *OfflineClassAssessment) ProcessTeacherName(assUserItem *v2.AssessmentUser, teacherMap map[string]*entity.IDName) (*entity.IDName, bool) {
	if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
		return nil, false
	}

	resultItem := &entity.IDName{
		ID:   assUserItem.UserID,
		Name: "",
	}

	if userItem, ok := teacherMap[assUserItem.UserID]; ok && userItem != nil {
		resultItem.Name = userItem.Name
	}

	return resultItem, true
}

// old

func (o *OfflineClassAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	return o.base.MatchSchedule()
}

func (o *OfflineClassAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	return o.base.MatchLessPlan()
}

func (o *OfflineClassAssessment) MatchProgram() (map[string]*entity.IDName, error) {
	onlineClass := NewOnlineClassAssessment(o.at, o.action)

	return onlineClass.MatchProgram()
}

func (o *OfflineClassAssessment) MatchSubject() (map[string][]*entity.IDName, error) {
	onlineClass := NewOnlineClassAssessment(o.at, o.action)

	return onlineClass.MatchSubject()
}

func (o *OfflineClassAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	return o.base.MatchTeacher()
}

func (o *OfflineClassAssessment) MatchClass() (map[string]*entity.IDName, error) {
	onlineClass := NewOnlineClassAssessment(o.at, o.action)

	return onlineClass.MatchClass()
}

func (o *OfflineClassAssessment) MatchOutcomes() (map[string]*v2.AssessmentOutcomeReply, error) {
	onlineClass := NewOnlineClassAssessment(o.at, o.action)

	return onlineClass.MatchOutcomes()
}

func (o *OfflineClassAssessment) MatchContents() ([]*v2.AssessmentContentReply, error) {
	libraryContents, err := o.at.FirstGetContentsFromSchedule()
	if err != nil {
		return nil, err
	}

	assessmentContentMap, err := o.at.FirstGetAssessmentContentMap()
	if err != nil {
		return nil, err
	}

	result := make([]*v2.AssessmentContentReply, 0)
	index := 0
	for _, item := range libraryContents {
		contentReplyItem := &v2.AssessmentContentReply{
			Number:          "0",
			ParentID:        "",
			ContentID:       item.ID,
			ContentName:     item.Name,
			Status:          v2.AssessmentContentStatusCovered,
			ContentType:     item.ContentType,
			FileType:        v2.AssessmentFileTypeNotChildSubContainer,
			MaxScore:        0,
			ReviewerComment: "",
			OutcomeIDs:      item.OutcomeIDs,
		}

		if item.ContentType == v2.AssessmentContentTypeLessonPlan {
			contentReplyItem.FileType = v2.AssessmentFileTypeHasChildContainer
			result = append(result, contentReplyItem)
			continue
		}

		index++
		contentReplyItem.Number = fmt.Sprintf("%d", index)

		if assessmentContentItem, ok := assessmentContentMap[item.ID]; ok {
			contentReplyItem.ReviewerComment = assessmentContentItem.ReviewerComment
			contentReplyItem.Status = assessmentContentItem.Status
		}

		result = append(result, contentReplyItem)
	}

	return result, nil
}

func (o *OfflineClassAssessment) MatchStudents(contents []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error) {
	//ctx := o.at.ctx

	assessmentUserMap, err := o.at.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	assessmentUsers, ok := assessmentUserMap[o.at.first.ID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	contentMapFromAssessment, err := o.at.FirstGetAssessmentContentMap()
	if err != nil {
		return nil, err
	}

	outcomeMapFromAssessment, err := o.at.FirstGetOutcomeFromAssessment()
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
			StudentID: item.UserID,
			//StudentName:   "",
			Status:        item.StatusByUser,
			ProcessStatus: item.StatusBySystem,
			Results:       nil,
		}

		for _, content := range contents {
			resultReply := &v2.AssessmentStudentResultReply{
				ContentID: content.ContentID,
			}

			userOutcomeReply := make([]*v2.AssessmentStudentResultOutcomeReply, 0)
			for _, outcomeID := range content.OutcomeIDs {
				var userOutcome *v2.AssessmentUserOutcome
				if assessmentContent, ok := contentMapFromAssessment[content.ContentID]; ok {
					key := o.at.GetKey([]string{
						item.ID,
						assessmentContent.ID,
						outcomeID,
					})
					userOutcome = outcomeMapFromAssessment[key]
				}
				userOutcomeReplyItem := &v2.AssessmentStudentResultOutcomeReply{
					OutcomeID: outcomeID,
				}
				if userOutcome != nil && userOutcome.Status != "" {
					userOutcomeReplyItem.Status = userOutcome.Status
				} else {
					if outcomeInfo, ok := outcomeMapFromContent[outcomeID]; ok && outcomeInfo.Assumed {
						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusAchieved
					} else {
						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusUnknown
					}
				}

				userOutcomeReply = append(userOutcomeReply, userOutcomeReplyItem)
			}
			resultReply.Outcomes = userOutcomeReply

			studentReply.Results = append(studentReply.Results, resultReply)
		}

		result = append(result, studentReply)
	}

	return result, nil
}
