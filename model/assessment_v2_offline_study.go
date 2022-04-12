package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da/assessmentV2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"time"
)

func NewOfflineStudyAssessmentPage(ag *AssessmentGrain) IAssessmentMatch {
	return &OfflineStudyAssessment{
		ag:     ag,
		action: AssessmentMatchActionPage,
		base:   NewBaseAssessment(ag),
	}
}

func NewOfflineStudyAssessmentDetail(ag *AssessmentGrain) IAssessmentMatch {
	return &OfflineStudyAssessment{
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

func (o *OfflineStudyAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	return o.base.MatchSchedule()
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
		stuTotal := studentCount[item.ID]
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
			result[item.ID] = float64(stuComplete) / float64(stuTotal)
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

		reviewerFeedback, ok := reviewerFeedbackMap[item.ID]
		if !ok {
			if len(studentResultItem.Outcomes) > 0 {
				resultItem.Results = append(resultItem.Results, studentResultItem)
			}

			result = append(result, resultItem)
			continue
		}

		resultItem.ReviewerComment = reviewerFeedback.ReviewerComment
		studentResultItem.AssessScore = v2.AssessmentUserAssessAverage
		if reviewerFeedback.AssessScore > 0 {
			studentResultItem.AssessScore = reviewerFeedback.AssessScore
		}
		studentResultItem.Attempted = true

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

		result = append(result, resultItem)
	}

	return result, nil
}

func (o *OfflineStudyAssessment) Update(req *v2.AssessmentUpdateReq) error {
	ctx := o.ag.ctx
	op := o.ag.op
	now := time.Now().Unix()

	log.Debug(ctx, "request data", log.Any("req", req))

	waitUpdateAssessment := o.ag.assessment

	// verify assessment status
	if !req.Action.Valid() {
		log.Warn(ctx, "req action invalid", log.Any("req", req))
		return constant.ErrInvalidArgs
	}
	if waitUpdateAssessment.Status == v2.AssessmentStatusComplete {
		log.Warn(ctx, "assessment is completed", log.Any("waitUpdateAssessment", waitUpdateAssessment))
		return ErrAssessmentHasCompleted
	}

	// verify assessment remaining time
	remainingTimeMap, err := o.MatchRemainingTime()
	if err != nil {
		return err
	}
	remainingTime, ok := remainingTimeMap[waitUpdateAssessment.ID]
	if !ok {
		log.Warn(ctx, "not found assessment remaining time", log.Any("waitUpdateAssessment", waitUpdateAssessment))
		return constant.ErrInvalidArgs
	}
	if remainingTime > 0 {
		log.Warn(ctx, "assessment remaining time is greater than 0", log.Int64("remainingTime", remainingTime), log.Any("waitUpdateAssessment", waitUpdateAssessment))
		return constant.ErrInvalidArgs
	}

	// prepare assessment user data
	assessmentUserMap, err := o.ag.GetAssessmentUserWithUserIDAndUserTypeMap()
	if err != nil {
		return err
	}
	waitUpdateAssessmentUsers := make([]*v2.AssessmentUser, 0, len(req.Students))
	waitUpdateFeedbackAssignments := make([]*entity.FeedbackAssignment, 0)
	assessmentUserIDs := make([]string, 0)
	reqStuResultMap := make(map[string]*v2.AssessmentStudentResultReq)
	reqStuOutcomeMap := make(map[string]*v2.AssessmentStudentResultOutcomeReq)
	reqReviewerCommentMap := make(map[string]string)
	assignmentMap := make(map[string]*v2.FeedbackAssignmentsReq)
	for _, item := range req.Students {
		if !item.Status.Valid() {
			log.Warn(ctx, "req student status is invalid", log.Any("studentItem", item))
			return constant.ErrInvalidArgs
		}

		stuKey := o.ag.GetKey([]string{
			item.StudentID,
			v2.AssessmentUserTypeStudent.String(),
		})

		assessmentUserItem, ok := assessmentUserMap[stuKey]
		if !ok {
			log.Warn(ctx, "not found student in assessment", log.Any("studentItem", item), log.Any("assessmentUserMap", assessmentUserMap))
			return constant.ErrInvalidArgs
		}
		assessmentUserItem.StatusByUser = item.Status
		assessmentUserItem.UpdateAt = now
		waitUpdateAssessmentUsers = append(waitUpdateAssessmentUsers, assessmentUserItem)
		assessmentUserIDs = append(assessmentUserIDs, assessmentUserItem.ID)

		if len(item.Results) <= 0 {
			continue
		}
		stuResult := item.Results[0]
		reqStuResultMap[assessmentUserItem.ID] = stuResult
		reqReviewerCommentMap[assessmentUserItem.ID] = item.ReviewerComment
		for _, outcomeReq := range stuResult.Outcomes {
			reqStuOutcomeMap[o.ag.GetKey([]string{assessmentUserItem.ID, outcomeReq.OutcomeID})] = outcomeReq
		}

		for _, assignItem := range stuResult.Assignments {
			assignmentMap[assignItem.ID] = assignItem
		}
	}

	// prepare assessment reviewer feedback
	waitUpdateReviewerFeedbacks := make([]*v2.AssessmentReviewerFeedback, 0, len(assessmentUserIDs))

	reviewerFeedbackCond := &assessmentV2.AssessmentUserResultCondition{
		AssessmentUserIDs: entity.NullStrings{
			Strings: assessmentUserIDs,
			Valid:   true,
		},
	}
	var reviewerFeedbacks []*v2.AssessmentReviewerFeedback
	err = assessmentV2.GetAssessmentUserResultDA().Query(ctx, reviewerFeedbackCond, &reviewerFeedbacks)
	if err != nil {
		return err
	}

	userOutcomeCond := &assessmentV2.AssessmentUserOutcomeCondition{
		AssessmentUserIDs: entity.NullStrings{
			Strings: assessmentUserIDs,
			Valid:   true,
		},
	}

	var userOutcomes []*v2.AssessmentUserOutcome
	err = assessmentV2.GetAssessmentUserOutcomeDA().Query(ctx, userOutcomeCond, &userOutcomes)
	if err != nil {
		log.Warn(ctx, "get assessment user outcomes error", log.Any("userOutcomeCond", userOutcomeCond), log.Any("req", req), log.Any("op", op))
		return err
	}
	waitUpdateUserOutcomes := make([]*v2.AssessmentUserOutcome, 0, len(userOutcomes))
	// user outcomes
	for _, item := range userOutcomes {
		if reqStuOutcome, ok := reqStuOutcomeMap[o.ag.GetKey([]string{item.AssessmentUserID, item.OutcomeID})]; ok {
			if !reqStuOutcome.Status.Valid() {
				log.Warn(ctx, "student outcome status invalid", log.Any("reqStuOutcome", reqStuOutcome), log.Any("req", req))
				return constant.ErrInvalidArgs
			}
			item.Status = reqStuOutcome.Status
			item.UpdateAt = now
			waitUpdateUserOutcomes = append(waitUpdateUserOutcomes, item)
		}
	}

	if len(reviewerFeedbacks) > 0 {
		feedbackIDs := make([]string, 0)
		for _, item := range reviewerFeedbacks {
			if reqStuResult, ok := reqStuResultMap[item.AssessmentUserID]; ok {
				if reqStuResult.AssessScore <= 0 {
					log.Warn(ctx, "student assessScore invalid in request", log.Any("req", req))
					return constant.ErrInvalidArgs
				}
				item.AssessScore = reqStuResult.AssessScore
				item.ReviewerID = op.UserID
			}
			if comment, ok := reqReviewerCommentMap[item.AssessmentUserID]; ok {
				item.ReviewerComment = comment
			}
			if req.Action == v2.AssessmentActionDraft {
				item.Status = v2.UserResultProcessStatusDraft
			} else if req.Action == v2.AssessmentActionComplete {
				item.Status = v2.UserResultProcessStatusComplete
				item.CompleteAt = now
			}
			item.UpdateAt = now

			waitUpdateReviewerFeedbacks = append(waitUpdateReviewerFeedbacks, item)
			feedbackIDs = append(feedbackIDs, item.StudentFeedbackID)
		}

		// assignment
		feedbackAssignCond := &da.FeedbackAssignmentCondition{
			FeedBackIDs: entity.NullStrings{
				Strings: feedbackIDs,
				Valid:   true,
			},
		}

		var feedbackAssigns []*entity.FeedbackAssignment
		err = da.GetFeedbackAssignmentDA().Query(ctx, feedbackAssignCond, &feedbackAssigns)
		if err != nil {
			log.Error(ctx, "get feedback assignment error", log.Any("feedbackAssignCond", feedbackAssignCond))
			return err
		}
		for _, item := range feedbackAssigns {
			if assignmentReq, ok := assignmentMap[item.ID]; ok && assignmentReq.ReviewAttachmentID != "" {
				item.ReviewAttachmentID = assignmentReq.ReviewAttachmentID
				item.UpdateAt = now
				waitUpdateFeedbackAssignments = append(waitUpdateFeedbackAssignments, item)
			}
		}
	}

	if req.Action == v2.AssessmentActionDraft {
		waitUpdateAssessment.Status = v2.AssessmentStatusInDraft
	} else if req.Action == v2.AssessmentActionComplete {
		waitUpdateAssessment.Status = v2.AssessmentStatusComplete
		waitUpdateAssessment.CompleteAt = now
	} else {
		log.Warn(ctx, "req action is invalid", log.Any("req", req))
		return constant.ErrInvalidArgs
	}

	waitUpdateAssessment.UpdateAt = now

	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, err := assessmentV2.GetAssessmentDA().UpdateTx(ctx, tx, waitUpdateAssessment)
		if err != nil {
			log.Error(ctx, "update assessment data error", log.Err(err), log.Any("waitUpdateAssessment", waitUpdateAssessment))
			return err
		}

		if len(waitUpdateAssessmentUsers) > 0 {
			_, err := assessmentV2.GetAssessmentUserDA().UpdateTx(ctx, tx, waitUpdateAssessmentUsers)
			if err != nil {
				log.Error(ctx, "update assessment user data error", log.Err(err), log.Any("waitUpdateAssessmentUsers", waitUpdateAssessmentUsers))
				return err
			}
		}
		if len(waitUpdateUserOutcomes) > 0 {
			_, err := assessmentV2.GetAssessmentUserOutcomeDA().UpdateTx(ctx, tx, waitUpdateUserOutcomes)
			if err != nil {
				log.Error(ctx, "update assessment user outcome data error", log.Err(err), log.Any("waitUpdateUserOutcomes", waitUpdateUserOutcomes))
				return err
			}
		}

		if len(waitUpdateReviewerFeedbacks) > 0 {
			_, err := assessmentV2.GetAssessmentUserResultDA().UpdateTx(ctx, tx, waitUpdateReviewerFeedbacks)
			if err != nil {
				log.Error(ctx, "update assessment reviewer feedback data error", log.Err(err), log.Any("waitUpdateReviewerFeedbacks", waitUpdateReviewerFeedbacks))
				return err
			}
		}

		if len(waitUpdateFeedbackAssignments) > 0 {
			_, err := da.GetFeedbackAssignmentDA().UpdateTx(ctx, tx, waitUpdateFeedbackAssignments)
			if err != nil {
				log.Error(ctx, "update feedback assignment data error", log.Err(err), log.Any("waitUpdateFeedbackAssignments", waitUpdateFeedbackAssignments))
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Error(ctx, "update data error", log.Err(err))
		return err
	}

	return nil
}
