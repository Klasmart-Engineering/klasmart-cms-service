package model

import (
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

func NewOfflineClassAssessmentPage(ags *AssessmentGrain) IAssessmentMatch {
	return &OfflineClassAssessment{
		ags:    ags,
		action: AssessmentMatchActionPage,
		base:   NewBaseAssessment(ags),
	}
}

func NewOfflineClassAssessmentDetail(ags *AssessmentGrain) IAssessmentMatch {
	return &OfflineClassAssessment{
		ags:    ags,
		action: AssessmentMatchActionDetail,
		base:   NewBaseAssessment(ags),
	}
}

type OfflineClassAssessment struct {
	EmptyAssessment

	base   BaseAssessment
	ags    *AssessmentGrain
	action AssessmentMatchAction
}

func (o *OfflineClassAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	return o.base.MatchSchedule()
}

func (o *OfflineClassAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	return o.base.MatchLessPlan()
}

func (o *OfflineClassAssessment) MatchProgram() (map[string]*entity.IDName, error) {
	onlineClass := NewOnlineClassAssessmentPage(o.ags)

	return onlineClass.MatchProgram()
}

func (o *OfflineClassAssessment) MatchSubject() (map[string][]*entity.IDName, error) {
	onlineClass := NewOnlineClassAssessmentPage(o.ags)

	return onlineClass.MatchSubject()
}

func (o *OfflineClassAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	onlineClass := NewOnlineClassAssessmentPage(o.ags)

	return onlineClass.MatchTeacher()
}

func (o *OfflineClassAssessment) MatchClass() (map[string]*entity.IDName, error) {
	onlineClass := NewOnlineClassAssessmentDetail(o.ags)

	return onlineClass.MatchClass()
}

func (o *OfflineClassAssessment) MatchOutcomes() (map[string]*v2.AssessmentOutcomeReply, error) {
	onlineClass := NewOnlineClassAssessmentDetail(o.ags)

	return onlineClass.MatchOutcomes()
}

func (o *OfflineClassAssessment) MatchContents() ([]*v2.AssessmentContentReply, error) {
	libraryContents, err := o.ags.SingleGetContentsFromSchedule()
	if err != nil {
		return nil, err
	}

	assessmentContentMap, err := o.ags.SingleGetAssessmentContentMap()
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
	ctx := o.ags.ctx

	userMap, err := o.ags.GetUserMap()
	if err != nil {
		return nil, err
	}

	assessmentUserMap, err := o.ags.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	assessmentUsers, ok := assessmentUserMap[o.ags.assessment.ID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	contentMapFromAssessment, err := o.ags.SingleGetAssessmentContentMap()
	if err != nil {
		return nil, err
	}

	outcomeMapFromAssessment, err := o.ags.SingleGetOutcomeFromAssessment()
	if err != nil {
		return nil, err
	}

	outcomeMapFromContent, err := o.ags.SingleGetOutcomeMapFromContent()
	if err != nil {
		return nil, err
	}

	result := make([]*v2.AssessmentStudentReply, 0, len(assessmentUsers))

	for _, item := range assessmentUsers {
		if item.UserType == v2.AssessmentUserTypeTeacher {
			continue
		}

		if item.StatusBySystem == v2.AssessmentUserStatusNotParticipate {
			continue
		}

		studentInfo, ok := userMap[item.UserID]
		if !ok {
			log.Warn(ctx, "not found user info from user service", log.Any("item", item), log.Any("userMap", userMap))
			studentInfo = &entity.IDName{
				ID:   item.UserID,
				Name: "",
			}
		}

		studentReply := &v2.AssessmentStudentReply{
			StudentID:   item.UserID,
			StudentName: studentInfo.Name,
			Status:      item.StatusByUser,
			Results:     nil,
		}

		for _, content := range contents {
			resultReply := &v2.AssessmentStudentResultReply{
				ContentID: content.ContentID,
			}

			userOutcomeReply := make([]*v2.AssessmentStudentResultOutcomeReply, 0)
			for _, outcomeID := range content.OutcomeIDs {
				var userOutcome *v2.AssessmentUserOutcome
				if assessmentContent, ok := contentMapFromAssessment[content.ContentID]; ok {
					key := o.ags.GetKey([]string{
						item.ID,
						assessmentContent.ID,
						outcomeID,
					})
					userOutcome = outcomeMapFromAssessment[key]
				}
				userOutcomeReplyItem := &v2.AssessmentStudentResultOutcomeReply{
					OutcomeID: outcomeID,
				}
				if userOutcome != nil {
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
