package model

import (
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

func NewOfflineStudyAssessmentPage(ag *AssessmentGrain) IAssessmentMatch {
	return &OnlineClassAssessment{
		ag:     ag,
		action: AssessmentMatchActionPage,
		base:   NewBaseAssessment(ag),
	}
}

func NewOfflineStudyAssessmentDetail(ag *AssessmentGrain) IAssessmentMatch {
	return &OnlineClassAssessment{
		ag:     ag,
		action: AssessmentMatchActionDetail,
		base:   NewBaseAssessment(ag),
	}
}

type OfflineStudyAssessment struct {
	EmptyAssessment

	base   BaseAssessment
	ag     *AssessmentGrain
	action AssessmentMatchAction
}

func (o *OfflineStudyAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	return o.base.MatchTeacher()
}

func (o *OfflineStudyAssessment) MatchClass() (map[string]*entity.IDName, error) {
	return o.base.MatchClass()
}

func (o *OfflineStudyAssessment) MatchCompleteRate() (map[string]float64, error) {
	ctx := o.ag.ctx

	assessmentUserMap, err := o.ag.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	reviewerFeedbackMap, err := o.ag.GetReviewerFeedbackMap()
	if err != nil {
		return nil, err
	}

	studentCount := make(map[string]int)
	studentCompleteCount := make(map[string]int)
	for key, users := range assessmentUserMap {
		for _, userItem := range users {
			if userItem.UserType == v2.AssessmentUserTypeStudent {
				studentCount[key]++

				if _, ok := reviewerFeedbackMap[userItem.ID]; ok {
					studentCompleteCount[key]++
				}
			}
		}
	}

	result := make(map[string]float64)
	for _, item := range o.ag.assessments {
		stuTotal := studentCompleteCount[item.ID]
		stuComplete := studentCompleteCount[item.ID]
		if stuTotal == 0 || stuComplete == 0 {
			result[item.ID] = 0
		} else if stuComplete > stuTotal {
			log.Warn(ctx, "Completion rate result greater than 1",
				log.Any("assessment", item),
				log.Any("assessmentUserMap", assessmentUserMap),
				log.Any("reviewerFeedbackMap", reviewerFeedbackMap))
			result[item.ID] = 1
		} else {
			result[item.ID] = float64(stuComplete / stuTotal)
		}
	}

	return result, nil
}

func (o *OfflineStudyAssessment) MatchRemainingTime() (map[string]int64, error) {
	onlineStudy := NewOnlineStudyAssessmentPage(o.ag)

	return onlineStudy.MatchRemainingTime()
}

func (o *OfflineStudyAssessment) MatchOutcomes() (map[string]*v2.AssessmentOutcomeReply, error) {
	outcomes, err := o.ag.SingleGetOutcomesFromSchedule()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*v2.AssessmentOutcomeReply, len(outcomes))
	if len(outcomes) <= 0 {
		return result, nil
	}

	for _, item := range outcomes {
		resultItem := &v2.AssessmentOutcomeReply{
			OutcomeID:          item.ID,
			OutcomeName:        item.Name,
			AssignedTo:         nil,
			Assumed:            item.Assumed,
			AssignedToLessPlan: false,
			AssignedToMaterial: false,
			ScoreThreshold:     item.ScoreThreshold,
		}
		result[item.ID] = resultItem
	}

	return result, nil
}

func (o *OfflineStudyAssessment) MatchStudents(contentsReply []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error) {
	ctx := o.ag.ctx
	op := o.ag.op

	assessmentUsers, err := o.ag.GetAssessmentUsers()
	if err != nil {
		return nil, err
	}
	userMap, err := o.ag.GetUserMap()
	if err != nil {
		return nil, err
	}

	feedbackCond := &da.ScheduleFeedbackCondition{
		ScheduleID: sql.NullString{
			String: o.ag.assessment.ScheduleID,
			Valid:  true,
		},
	}
	scheduleFeedbacks, err := GetScheduleFeedbackModel().Query(ctx, op, feedbackCond)
	if err != nil {
		return nil, err
	}
	scheduleFeedbackMap := make(map[string][]*entity.ScheduleFeedbackView, len(scheduleFeedbacks))
	for _, item := range scheduleFeedbacks {
		scheduleFeedbackMap[item.UserID] = append(scheduleFeedbackMap[item.UserID], item)
	}

	reviewerFeedbackMap, err := o.ag.amg.GetReviewerFeedbackMap()
	if err != nil {
		return nil, err
	}

	assessmentOutcomeMap, err := o.ag.SingleGetOutcomeFromAssessment()
	if err != nil {
		return nil, err
	}

	outcomesFromSchedule, err := o.ag.SingleGetOutcomesFromSchedule()
	if err != nil {
		return nil, err
	}

	result := make([]*v2.AssessmentStudentReply, 0, len(assessmentUsers))
	for _, item := range assessmentUsers {
		if item.UserType != v2.AssessmentUserTypeStudent {
			continue
		}

		resultItem := &v2.AssessmentStudentReply{
			StudentID: item.UserID,
			Status:    item.StatusByUser,
			Results:   make([]*v2.AssessmentStudentResultReply, 0),
		}

		if userInfo, ok := userMap[item.UserID]; ok && userInfo != nil {
			resultItem.StudentName = userInfo.Name
		}

		studentResultItem := new(v2.AssessmentStudentResultReply)
		if reviewerFeedback, ok := reviewerFeedbackMap[item.ID]; ok {
			resultItem.ReviewerComment = reviewerFeedback.ReviewerComment
			studentResultItem.AssessScore = reviewerFeedback.AssessScore
			studentResultItem.Attempted = true
		}

		studentResultItem.Outcomes = make([]*v2.AssessmentStudentResultOutcomeReply, 0, len(outcomesFromSchedule))
		for _, outcomeItem := range outcomesFromSchedule {
			aoKey := o.ag.GetKey([]string{
				item.ID,
				"",
				outcomeItem.ID,
			})

			stuOutcomeReplyItem := &v2.AssessmentStudentResultOutcomeReply{
				OutcomeID: outcomeItem.ID,
				Status:    v2.AssessmentUserOutcomeStatusUnknown,
			}
			if outcomeItem.Assumed {
				stuOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusAchieved
			}
			if assOutcomeItem, ok := assessmentOutcomeMap[aoKey]; ok && assOutcomeItem.Status != "" {
				stuOutcomeReplyItem.Status = assOutcomeItem.Status
			}

			studentResultItem.Outcomes = append(studentResultItem.Outcomes, stuOutcomeReplyItem)
		}

		if studentFeedbackItems, ok := scheduleFeedbackMap[item.UserID]; ok {
			studentResultItem.StudentFeedbacks = make([]*v2.StudentResultFeedBacksReply, 0, len(studentFeedbackItems))
			for _, feedbackItem := range studentFeedbackItems {
				feedbackReply := &v2.StudentResultFeedBacksReply{
					ID:          feedbackItem.ID,
					ScheduleID:  feedbackItem.ScheduleID,
					UserID:      feedbackItem.UserID,
					Comment:     feedbackItem.Comment,
					CreateAt:    feedbackItem.CreateAt,
					Assignments: feedbackItem.Assignments,
				}
				studentResultItem.StudentFeedbacks = append(studentResultItem.StudentFeedbacks, feedbackReply)
			}
		}

		resultItem.Results = append(resultItem.Results, studentResultItem)
	}

	return result, nil
}
