package model

import (
	"context"
	"fmt"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da/assessmentV2"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
	"time"
)

func NewOnlineClassAssessment() IAssessmentProcessor {
	return &OnlineClassAssessment{}
}

type OnlineClassAssessment struct{}

func (o *OnlineClassAssessment) Update(ctx context.Context, op *entity.Operator, assessment *v2.Assessment, req *v2.AssessmentUpdateReq) error {
	now := time.Now().Unix()
	at, err := NewAssessmentInit(ctx, op, assessment)
	if err != nil {
		return err
	}

	if err := at.initAssessmentUserWithIDTypeMap(); err != nil {
		return err
	}
	if err := at.initReviewerFeedbackMap(); err != nil {
		return err
	}

	userIDAndUserTypeMap := at.assessmentUserIDTypeMap
	reviewerFeedbackMap := at.reviewerFeedbackMap

	waitAddReviewerFeedbacks := make([]*v2.AssessmentReviewerFeedback, 0)
	waitUpdatedReviewerFeedbacks := make([]*v2.AssessmentReviewerFeedback, 0)

	waitUpdatedUsers := make([]*v2.AssessmentUser, 0)
	for _, item := range req.Students {
		existItem, ok := userIDAndUserTypeMap[GetAssessmentKey([]string{item.StudentID, v2.AssessmentUserTypeStudent.String()})]
		if !ok {
			log.Warn(ctx, "student not exist", log.Any("userIDAndUserTypeMap", userIDAndUserTypeMap), log.Any("reqItem", item))
			return constant.ErrInvalidArgs
		}
		if !item.Status.Valid() {
			log.Warn(ctx, "student status invalid", log.Any("userIDAndUserTypeMap", userIDAndUserTypeMap), log.Any("reqItem", item))
			return constant.ErrInvalidArgs
		}
		existItem.StatusByUser = item.Status

		if req.Action == v2.AssessmentActionComplete {
			if existItem.StatusBySystem == v2.AssessmentUserSystemStatusDone || existItem.StatusBySystem == v2.AssessmentUserSystemStatusResubmitted {
				existItem.StatusBySystem = v2.AssessmentUserSystemStatusCompleted
				existItem.CompletedAt = now
			}
		}
		existItem.UpdateAt = now

		if reviewerFeedbackItem, ok := reviewerFeedbackMap[existItem.ID]; ok {
			if reviewerFeedbackItem.ReviewerComment != item.ReviewerComment {
				reviewerFeedbackItem.ReviewerComment = item.ReviewerComment
				reviewerFeedbackItem.ReviewerID = op.UserID
				reviewerFeedbackItem.UpdateAt = now
				waitUpdatedReviewerFeedbacks = append(waitUpdatedReviewerFeedbacks, reviewerFeedbackItem)
			}
		} else if item.ReviewerComment != "" {
			reviewerFeedbackItem := &v2.AssessmentReviewerFeedback{
				ID:                utils.NewID(),
				AssessmentUserID:  existItem.ID,
				ReviewerID:        op.UserID,
				StudentFeedbackID: "",
				AssessScore:       0,
				ReviewerComment:   item.ReviewerComment,
				CreateAt:          now,
				UpdateAt:          0,
				DeleteAt:          0,
			}
			waitAddReviewerFeedbacks = append(waitAddReviewerFeedbacks, reviewerFeedbackItem)
		}
		waitUpdatedUsers = append(waitUpdatedUsers, existItem)
	}

	if err := at.initRoomData(); err != nil {
		return err
	}

	roomData := at.liveRoom
	canSetScoreContentMap := make(map[string]map[string]*AllowEditScoreContent)
	if roomData != nil && len(roomData.ScoresByUser) > 0 {
		canSetScoreContentMap, err = GetAssessmentExternalService().AllowEditScoreContent(ctx, roomData.ScoresByUser)
		if err != nil {
			return err
		}
	}

	//if waitUpdatedAssessment.AssessmentType == v2.AssessmentTypeReviewStudy {
	//	return a.updateReviewStudyAssessment(ctx, op, updateReviewStudyAssessmentInput{
	//		status:                      status,
	//		req:                         req,
	//		waitUpdatedAssessment:       waitUpdatedAssessment,
	//		waitUpdatedUsers:            waitUpdatedUsers,
	//		userIDAndUserTypeMap:        userIDAndUserTypeMap,
	//		canSetScoreContentMap:       canSetScoreContentMap,
	//		waitAddReviewerFeedbacks:    waitAddReviewerFeedbacks,
	//		waitUpdateReviewerFeedbacks: waitUpdatedReviewerFeedbacks,
	//	})
	//}

	if err := at.initContentsFromSchedule(); err != nil {
		return err
	}

	if err := at.initAssessmentContentMap(); err != nil {
		return err
	}

	scheduleContents := at.contentsFromSchedule
	assessmentContentMap := at.contentMapFromAssessment

	waitAddContentMap := make(map[string]*v2.AssessmentContent)
	for _, item := range scheduleContents {
		if _, ok := assessmentContentMap[item.ID]; !ok {
			waitAddContentMap[item.ID] = &v2.AssessmentContent{
				ID:           utils.NewID(),
				AssessmentID: assessment.ID,
				ContentID:    item.ID,
				ContentType:  item.ContentType,
				Status:       v2.AssessmentContentStatusNotCovered,
				CreateAt:     now,
			}
		}
	}

	waitUpdateContents := make([]*v2.AssessmentContent, 0, len(assessmentContentMap))
	for _, item := range req.Contents {
		if contentItem, ok := assessmentContentMap[item.ContentID]; ok {
			if !item.Status.Valid() {
				log.Warn(ctx, "content status is invalid", log.Any("item", item), log.Any("req.Contents", req.Contents))
				return constant.ErrInvalidArgs
			}
			contentItem.Status = item.Status
			contentItem.ReviewerComment = item.ReviewerComment
			contentItem.UpdateAt = now
			waitUpdateContents = append(waitUpdateContents, contentItem)
		} else {
			if waitAddContentItem, ok := waitAddContentMap[item.ContentID]; ok {
				if !item.Status.Valid() {
					log.Warn(ctx, "content status is invalid", log.Any("item", item), log.Any("req.Contents", req.Contents))
					return constant.ErrInvalidArgs
				}
				waitAddContentItem.ReviewerComment = item.ReviewerComment
				waitAddContentItem.Status = item.Status
				waitAddContentItem.CreateAt = now
			}
		}
	}

	waitAddContents := make([]*v2.AssessmentContent, 0, len(waitAddContentMap))
	for _, item := range waitAddContentMap {
		waitAddContents = append(waitAddContents, item)
	}
	allAssessmentContents := append(waitUpdateContents, waitAddContents...)

	// outcome
	contentOutcomeIDMap := make(map[string][]string, len(scheduleContents))
	for _, item := range scheduleContents {
		contentOutcomeIDMap[item.ID] = item.OutcomeIDs
	}

	outcomeIDs := make([]string, 0)
	for _, item := range scheduleContents {
		outcomeIDs = append(outcomeIDs, item.OutcomeIDs...)
	}
	outcomeIDs = utils.SliceDeduplicationExcludeEmpty(outcomeIDs)
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, op, dbo.MustGetDB(ctx), outcomeIDs)
	if err != nil {
		return err
	}
	outcomeMap := make(map[string]*entity.Outcome)
	for _, item := range outcomes {
		outcomeMap[item.ID] = item
	}

	if err := at.initOutcomeFromAssessment(); err != nil {
		return err
	}
	outcomeFromAssessmentMap := at.outcomeMapFromAssessment

	waitAddAssessmentOutcomeMap := make(map[string]*v2.AssessmentUserOutcome)
	for _, userItem := range userIDAndUserTypeMap {
		if userItem.UserType == v2.AssessmentUserTypeTeacher {
			continue
		}
		for _, contentItem := range allAssessmentContents {
			if outcomeIDs, ok := contentOutcomeIDMap[contentItem.ContentID]; ok {
				for _, outcomeID := range outcomeIDs {
					if outcomeItem, ok := outcomeMap[outcomeID]; ok {
						key := GetAssessmentKey([]string{userItem.ID, contentItem.ID, outcomeID})
						if _, ok := outcomeFromAssessmentMap[key]; ok {
							continue
						}
						waitAddOutcomeItem := &v2.AssessmentUserOutcome{
							ID:                  utils.NewID(),
							AssessmentUserID:    userItem.ID,
							AssessmentContentID: contentItem.ID,
							OutcomeID:           outcomeID,
							CreateAt:            now,
						}
						if outcomeItem.Assumed {
							waitAddOutcomeItem.Status = v2.AssessmentUserOutcomeStatusAchieved
						}

						waitAddAssessmentOutcomeMap[key] = waitAddOutcomeItem
					}
				}
			}
		}
	}

	// user comment,score, outcomes
	newScores := make([]*external.H5PSetScoreRequest, 0)

	contentReqMap := make(map[string]*v2.AssessmentUpdateContentReq)
	for _, item := range req.Contents {
		contentReqMap[item.ContentID] = item
	}

	allAssessmentContentMap := make(map[string]*v2.AssessmentContent)
	for _, item := range allAssessmentContents {
		allAssessmentContentMap[item.ContentID] = item
	}
	waitUpdateAssessmentOutcomes := make([]*v2.AssessmentUserOutcome, 0)

	for _, stuItem := range req.Students {
		if stuItem.Status == v2.AssessmentUserStatusNotParticipate {
			continue
		}
		// verify student data
		assessmentUserItem, ok := userIDAndUserTypeMap[GetAssessmentKey([]string{stuItem.StudentID, v2.AssessmentUserTypeStudent.String()})]
		if !ok {
			log.Warn(ctx, "student not exist", log.Any("userIDAndUserTypeMap", userIDAndUserTypeMap), log.Any("stuItem", stuItem))
			return constant.ErrInvalidArgs
		}

		for _, stuResult := range stuItem.Results {
			// verify student content data
			if assessmentContentItem, ok := allAssessmentContentMap[stuResult.ContentID]; ok {
				for _, outcomeItem := range stuResult.Outcomes {
					if !outcomeItem.Status.Valid() {
						log.Warn(ctx, "student outcome status invalid", log.Any("req", req), log.Any("outcomeItem", outcomeItem))
						return constant.ErrInvalidArgs
					}
					key := GetAssessmentKey([]string{assessmentUserItem.ID, assessmentContentItem.ID, outcomeItem.OutcomeID})
					if outcomeFromAssessmentItem, ok := outcomeFromAssessmentMap[key]; ok {
						outcomeFromAssessmentItem.Status = outcomeItem.Status
						outcomeFromAssessmentItem.UpdateAt = now
						waitUpdateAssessmentOutcomes = append(waitUpdateAssessmentOutcomes, outcomeFromAssessmentItem)
					} else if waitAddOutcomeItem, ok := waitAddAssessmentOutcomeMap[key]; ok {
						waitAddOutcomeItem.Status = outcomeItem.Status
					} else {
						log.Warn(ctx, "student outcome invalid", log.Any("outcomeFromAssessmentMap", outcomeFromAssessmentMap), log.Any("waitAddAssessmentOutcomeMap", waitAddAssessmentOutcomeMap), log.Any("stuItem", stuItem))
						continue
					}
				}
			}
			if contentItem, ok := contentReqMap[stuResult.ContentID]; ok {
				if stuContentMap, ok := canSetScoreContentMap[stuItem.StudentID]; ok {
					if canSetScoreContentItem, ok := stuContentMap[contentItem.ContentID]; ok {
						newScore := &external.H5PSetScoreRequest{
							RoomID:    assessment.ScheduleID,
							StudentID: stuItem.StudentID,
							Score:     stuResult.Score,
						}

						newScore.ContentID = canSetScoreContentItem.ContentID
						newScore.SubContentID = canSetScoreContentItem.SubContentID

						newScores = append(newScores, newScore)
					}
				}
			}
		}
	}

	// update student score
	if len(newScores) > 0 {
		if _, err := external.GetH5PRoomScoreServiceProvider().BatchSet(ctx, op, newScores); err != nil {
			log.Warn(ctx, "set student score error", log.Err(err), log.Any("newScores", newScores))
			return err
		}
	}

	var status v2.AssessmentStatus
	if req.Action == v2.AssessmentActionDraft {
		status = v2.AssessmentStatusInDraft
	} else if req.Action == v2.AssessmentActionComplete {
		status = v2.AssessmentStatusComplete
	} else {
		log.Warn(ctx, "req action is invalid", log.Any("req", req))
		return constant.ErrInvalidArgs
	}

	assessment.UpdateAt = now
	assessment.Status = status
	if status == v2.AssessmentStatusComplete {
		assessment.CompleteAt = now
	}

	waitAddAssessmentOutcomes := make([]*v2.AssessmentUserOutcome, 0, len(waitAddAssessmentOutcomeMap))
	for _, item := range waitAddAssessmentOutcomeMap {
		waitAddAssessmentOutcomes = append(waitAddAssessmentOutcomes, item)
	}

	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		if _, err := assessmentV2.GetAssessmentDA().UpdateTx(ctx, tx, assessment); err != nil {
			return err
		}

		if len(waitUpdatedUsers) > 0 {
			if _, err = assessmentV2.GetAssessmentUserDA().UpdateTx(ctx, tx, waitUpdatedUsers); err != nil {
				return err
			}
		}

		if len(waitAddContents) > 0 {
			if _, err = assessmentV2.GetAssessmentContentDA().InsertTx(ctx, tx, waitAddContents); err != nil {
				return err
			}
		}

		if len(waitUpdateContents) > 0 {
			if _, err = assessmentV2.GetAssessmentContentDA().UpdateTx(ctx, tx, waitUpdateContents); err != nil {
				return err
			}
		}

		if len(waitAddAssessmentOutcomes) > 0 {
			if _, err = assessmentV2.GetAssessmentUserOutcomeDA().InsertTx(ctx, tx, waitAddAssessmentOutcomes); err != nil {
				return err
			}
		}

		if len(waitUpdateAssessmentOutcomes) > 0 {
			if _, err = assessmentV2.GetAssessmentUserOutcomeDA().UpdateTx(ctx, tx, waitUpdateAssessmentOutcomes); err != nil {
				return err
			}
		}

		if len(waitAddReviewerFeedbacks) > 0 {
			if _, err := assessmentV2.GetAssessmentUserResultDA().InsertTx(ctx, tx, waitAddReviewerFeedbacks); err != nil {
				return err
			}
		}

		if len(waitUpdatedReviewerFeedbacks) > 0 {
			if _, err := assessmentV2.GetAssessmentUserResultDA().UpdateTx(ctx, tx, waitUpdatedReviewerFeedbacks); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (o *OnlineClassAssessment) ProcessCompleteRate(ctx context.Context, assessmentUsers []*v2.AssessmentUser, roomData *external.RoomInfo, stuReviewMap map[string]*entity.ScheduleReview, reviewerFeedbackMap map[string]*v2.AssessmentReviewerFeedback) float64 {
	return 0
}

func (o *OnlineClassAssessment) ProcessTeacherName(assUserItem *v2.AssessmentUser, teacherMap map[string]*entity.IDName) (*entity.IDName, bool) {
	if teacherID, ok := o.ProcessTeacherID(assUserItem); ok {
		resultItem := &entity.IDName{
			ID:   teacherID,
			Name: "",
		}

		if userItem, ok := teacherMap[teacherID]; ok && userItem != nil {
			resultItem.Name = userItem.Name
		}
		return resultItem, true
	}
	return nil, false
}

func (o *OnlineClassAssessment) ProcessTeacherID(assUserItem *v2.AssessmentUser) (string, bool) {
	if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
		return "", false
	}
	if assUserItem.StatusBySystem == v2.AssessmentUserSystemStatusNotStarted {
		return "", false
	}

	return assUserItem.UserID, true
}

func (o *OnlineClassAssessment) ProcessContents(ctx context.Context, at *AssessmentInit) ([]*v2.AssessmentContentReply, error) {
	contentsFromSchedule := at.contentsFromSchedule
	contentMapFromAssessment := at.contentMapFromAssessment
	contentMapFromLiveRoom := at.contentMapFromLiveRoom

	result := make([]*v2.AssessmentContentReply, 0)

	index := 0
	for _, item := range contentsFromSchedule {
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
			ContentSubtype:  item.FileType.String(),
		}

		if item.ContentType == v2.AssessmentContentTypeLessonPlan {
			contentReplyItem.FileType = v2.AssessmentFileTypeHasChildContainer
			result = append(result, contentReplyItem)
			continue
		}

		index++
		contentReplyItem.Number = fmt.Sprintf("%d", index)

		if assessmentContentItem, ok := contentMapFromAssessment[item.ID]; ok {
			contentReplyItem.ReviewerComment = assessmentContentItem.ReviewerComment
			contentReplyItem.Status = assessmentContentItem.Status
		}

		if roomContentItem, ok := contentMapFromLiveRoom[item.ID]; ok {
			contentReplyItem.ContentSubtype = roomContentItem.Type
			contentReplyItem.H5PID = roomContentItem.H5PID
			contentReplyItem.MaxScore = roomContentItem.MaxScore
			contentReplyItem.H5PSubID = roomContentItem.SubContentID
			contentReplyItem.RoomProvideContentID = roomContentItem.ContentUniqueID

			if roomContentItem.FileType == external.FileTypeH5P {
				if roomContentItem.MaxScore > 0 {
					contentReplyItem.FileType = v2.AssessmentFileTypeSupportScoreStandAlone
				} else {
					contentReplyItem.FileType = v2.AssessmentFileTypeNotSupportScoreStandAlone
				}
			} else {
				contentReplyItem.FileType = v2.AssessmentFileTypeNotChildSubContainer
			}

			if len(roomContentItem.Children) > 0 {
				//contentReplyItem.IgnoreCalculateScore = true
				contentReplyItem.FileType = v2.AssessmentFileTypeHasChildContainer
				result = append(result, contentReplyItem)

				for i, child := range roomContentItem.Children {
					if child.SubContentID == "" {
						log.Warn(ctx, "sub content id is empty", log.Any("contentItem", item))
						continue
					}
					o.appendContent(ctx, child, contentReplyItem, &result, contentReplyItem.Number, i+1)
				}
			} else {
				result = append(result, contentReplyItem)
			}
		} else {
			result = append(result, contentReplyItem)
		}
	}

	return result, nil
}

func (o *OnlineClassAssessment) appendContent(ctx context.Context, roomContent *RoomContentTree, materialItem *v2.AssessmentContentReply, result *[]*v2.AssessmentContentReply, prefix string, index int) {
	replyItem := &v2.AssessmentContentReply{
		Number:               fmt.Sprintf("%s-%d", prefix, index),
		ParentID:             materialItem.ContentID,
		ContentID:            roomContent.ContentUniqueID,
		ContentName:          materialItem.ContentName,
		ReviewerComment:      "",
		Status:               materialItem.Status,
		OutcomeIDs:           materialItem.OutcomeIDs,
		ContentType:          v2.AssessmentContentTypeUnknown,
		ContentSubtype:       roomContent.Type,
		FileType:             v2.AssessmentFileTypeNotUnknown,
		MaxScore:             roomContent.MaxScore,
		H5PID:                roomContent.H5PID,
		H5PSubID:             roomContent.SubContentID,
		RoomProvideContentID: roomContent.ContentUniqueID,
	}

	if roomContent.FileType == external.FileTypeH5P {
		if roomContent.MaxScore > 0 {
			replyItem.FileType = v2.AssessmentFileTypeSupportScoreStandAlone
		} else {
			replyItem.FileType = v2.AssessmentFileTypeNotSupportScoreStandAlone
		}
	}

	if len(roomContent.Children) > 0 {
		replyItem.FileType = v2.AssessmentFileTypeHasChildContainer
	}

	*result = append(*result, replyItem)
	for i, item := range roomContent.Children {
		if item.SubContentID == "" {
			log.Warn(ctx, "sub content id is empty", log.Any("contentItem", item))
			continue
		}
		o.appendContent(ctx, item, materialItem, result, replyItem.Number, i+1)
	}
}

func (o *OnlineClassAssessment) ProcessDiffContents(ctx context.Context, at *AssessmentInit) []*v2.AssessmentDiffContentStudentsReply {
	return make([]*v2.AssessmentDiffContentStudentsReply, 0)
}

func (o *OnlineClassAssessment) ProcessStudents(ctx context.Context, at *AssessmentInit, contentsReply []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error) {
	assessmentUsers := at.assessmentUsers
	commentResultMap := at.commentResultMap
	assessmentOutcomeMap := at.outcomeMapFromAssessment
	userScoresMap := at.roomUserScoreMap
	contentMapFromAssessment := at.contentMapFromAssessment
	outcomeMapFromContent := at.outcomeMapFromContent

	roomUserResultMap := make(map[string]*RoomUserScore)
	for userID, scores := range userScoresMap {
		for _, scoreItem := range scores {
			key := GetAssessmentKey([]string{
				userID,
				scoreItem.ContentUniqueID,
			})
			roomUserResultMap[key] = scoreItem
		}
	}

	contentScoreMap, studentScoreMap := at.summaryRoomScores(userScoresMap, contentsReply)

	result := make([]*v2.AssessmentStudentReply, 0, len(assessmentUsers))

	for _, item := range assessmentUsers {
		if item.UserType == v2.AssessmentUserTypeTeacher {
			continue
		}
		if item.StatusBySystem == v2.AssessmentUserSystemStatusNotStarted {
			continue
		}

		studentReply := &v2.AssessmentStudentReply{
			StudentID: item.UserID,
			//StudentName:   "",
			Status:        item.StatusByUser,
			ProcessStatus: item.StatusBySystem,
			Results:       nil,
		}

		if comment, ok := commentResultMap[item.ID]; ok {
			studentReply.ReviewerComment = comment
		} else {
			studentReply.ReviewerComment = commentResultMap[item.UserID]
		}

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
				studentScoreKey := GetAssessmentKey([]string{
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
					key := GetAssessmentKey([]string{
						item.ID,
						assessmentContent.ID,
						outcomeID,
					})
					userOutcome = assessmentOutcomeMap[key]
				}
				userOutcomeReplyItem := &v2.AssessmentStudentResultOutcomeReply{
					OutcomeID: outcomeID,
				}
				if at.assessment.Status == v2.AssessmentStatusInDraft ||
					at.assessment.Status == v2.AssessmentStatusComplete {
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

			roomKey := GetAssessmentKey([]string{
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

func (o *OnlineClassAssessment) ProcessRemainingTime(ctx context.Context, dueAt int64, assessmentCreateAt int64) int64 {
	return 0
}

// old

//func (o *OnlineClassAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
//	return o.base.MatchSchedule()
//}
//
//func (o *OnlineClassAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
//	return o.base.MatchLessPlan()
//}
//
//func (o *OnlineClassAssessment) MatchProgram() (map[string]*entity.IDName, error) {
//	scheduleMap, err := o.at.GetScheduleMap()
//	if err != nil {
//		return nil, err
//	}
//
//	programMap, err := o.at.GetProgramMap()
//	if err != nil {
//		return nil, err
//	}
//
//	result := make(map[string]*entity.IDName, len(o.at.assessments))
//	for _, item := range o.at.assessments {
//		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
//			result[item.ID] = programMap[schedule.ProgramID]
//		}
//	}
//
//	return result, nil
//}
//
//func (o *OnlineClassAssessment) MatchSubject() (map[string][]*entity.IDName, error) {
//	relationMap, err := o.at.GetScheduleRelationMap()
//	if err != nil {
//		return nil, err
//	}
//
//	subjectMap, err := o.at.GetSubjectMap()
//	if err != nil {
//		return nil, err
//	}
//
//	result := make(map[string][]*entity.IDName, len(o.at.assessments))
//	for _, item := range o.at.assessments {
//		if srItems, ok := relationMap[item.ScheduleID]; ok {
//			for _, srItem := range srItems {
//				if srItem.RelationType != entity.ScheduleRelationTypeSubject {
//					continue
//				}
//				if subItem, ok := subjectMap[srItem.RelationID]; ok && subItem != nil {
//					result[item.ID] = append(result[item.ID], subItem)
//				}
//			}
//		}
//	}
//
//	return result, nil
//}
//
//func (o *OnlineClassAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
//	assessmentUserMap, err := o.at.GetAssessmentUserMap()
//	if err != nil {
//		return nil, err
//	}
//
//	userMap := make(map[string]*entity.IDName)
//	if o.action == AssessmentMatchActionPage {
//		userMap, err = o.at.GetTeacherMap()
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	result := make(map[string][]*entity.IDName, len(o.at.assessments))
//	for _, item := range o.at.assessments {
//		if assUserItems, ok := assessmentUserMap[item.ID]; ok {
//			for _, assUserItem := range assUserItems {
//				if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
//					continue
//				}
//				if assUserItem.StatusBySystem == v2.AssessmentUserSystemStatusNotStarted {
//					continue
//				}
//				resultItem := &entity.IDName{
//					ID:   assUserItem.UserID,
//					Name: "",
//				}
//
//				if userItem, ok := userMap[assUserItem.UserID]; ok && userItem != nil {
//					resultItem.Name = userItem.Name
//				}
//				result[item.ID] = append(result[item.ID], resultItem)
//			}
//		}
//	}
//
//	return result, nil
//}
//
//func (o *OnlineClassAssessment) MatchClass() (map[string]*entity.IDName, error) {
//	if o.action == AssessmentMatchActionPage {
//		return o.EmptyAssessment.MatchClass()
//	}
//	return o.base.MatchClass()
//}
//
//func (o *OnlineClassAssessment) MatchOutcomes() (map[string]*v2.AssessmentOutcomeReply, error) {
//	contents, err := o.at.FirstGetContentsFromSchedule()
//	if err != nil {
//		return nil, err
//	}
//
//	outcomeMap, err := o.at.FirstGetOutcomeMapFromContent()
//	if err != nil {
//		return nil, err
//	}
//
//	result := make(map[string]*v2.AssessmentOutcomeReply, len(outcomeMap))
//
//	for _, item := range outcomeMap {
//		result[item.ID] = &v2.AssessmentOutcomeReply{
//			OutcomeID:          item.ID,
//			OutcomeName:        item.Name,
//			AssignedTo:         nil,
//			Assumed:            item.Assumed,
//			AssignedToLessPlan: false,
//			AssignedToMaterial: false,
//			ScoreThreshold:     item.ScoreThreshold,
//		}
//	}
//
//	for _, materialItem := range contents {
//		for _, outcomeID := range materialItem.OutcomeIDs {
//			if outcomeItem, ok := result[outcomeID]; ok {
//				if materialItem.ContentType == v2.AssessmentContentTypeLessonPlan {
//					outcomeItem.AssignedToLessPlan = true
//				}
//				if materialItem.ContentType == v2.AssessmentContentTypeLessonMaterial {
//					outcomeItem.AssignedToMaterial = true
//				}
//			}
//		}
//	}
//
//	for _, outcomeItem := range result {
//		if outcomeItem.AssignedToLessPlan {
//			outcomeItem.AssignedTo = append(outcomeItem.AssignedTo, v2.AssessmentOutcomeAssignTypeLessonPlan)
//		}
//		if outcomeItem.AssignedToMaterial {
//			outcomeItem.AssignedTo = append(outcomeItem.AssignedTo, v2.AssessmentOutcomeAssignTypeLessonMaterial)
//		}
//	}
//
//	return result, nil
//}
//
//func (o *OnlineClassAssessment) MatchStudents(contentsReply []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error) {
//	ctx := o.at.ctx
//
//	assessmentUserMap, err := o.at.GetAssessmentUserMap()
//	if err != nil {
//		return nil, err
//	}
//
//	assessmentUsers, ok := assessmentUserMap[o.at.first.ID]
//	if !ok {
//		return nil, constant.ErrRecordNotFound
//	}
//
//	commentResultMap, err := o.at.FirstGetCommentResultMap()
//	if err != nil {
//		return nil, err
//	}
//
//	assessmentOutcomeMap, err := o.at.FirstGetOutcomeFromAssessment()
//	if err != nil {
//		return nil, err
//	}
//
//	userScoresMap, _, err := o.at.FirstGetRoomData()
//	if err != nil {
//		return nil, err
//	}
//
//	roomUserResultMap := make(map[string]*RoomUserScore)
//	for userID, scores := range userScoresMap {
//		for _, scoreItem := range scores {
//			key := o.at.GetKey([]string{
//				userID,
//				scoreItem.ContentUniqueID,
//			})
//			roomUserResultMap[key] = scoreItem
//		}
//	}
//
//	contentScoreMap, studentScoreMap := o.at.summaryRoomScores(userScoresMap, contentsReply)
//
//	contentMapFromAssessment, err := o.at.FirstGetAssessmentContentMap()
//	if err != nil {
//		return nil, err
//	}
//
//	outcomeMapFromContent, err := o.at.FirstGetOutcomeMapFromContent()
//	if err != nil {
//		return nil, err
//	}
//
//	result := make([]*v2.AssessmentStudentReply, 0, len(assessmentUsers))
//
//	for _, item := range assessmentUsers {
//		if item.UserType == v2.AssessmentUserTypeTeacher {
//			continue
//		}
//		if item.StatusBySystem == v2.AssessmentUserSystemStatusNotStarted {
//			continue
//		}
//
//		studentReply := &v2.AssessmentStudentReply{
//			StudentID: item.UserID,
//			//StudentName:   "",
//			Status:        item.StatusByUser,
//			ProcessStatus: item.StatusBySystem,
//			Results:       nil,
//		}
//
//		if comment, ok := commentResultMap[item.ID]; ok {
//			studentReply.ReviewerComment = comment
//		} else {
//			studentReply.ReviewerComment = commentResultMap[item.UserID]
//		}
//
//		for _, content := range contentsReply {
//			resultReply := &v2.AssessmentStudentResultReply{
//				ContentID: content.ContentID,
//			}
//
//			contentID := content.ContentID
//			if content.ContentType == v2.AssessmentContentTypeUnknown {
//				contentID = content.ParentID
//			}
//
//			var studentContentScore float32
//			if contentScoreItem, ok := contentScoreMap[contentID]; ok && contentScoreItem != 0 {
//				studentScoreKey := o.at.GetKey([]string{
//					item.UserID,
//					contentID,
//				})
//				if studentScoreItem, ok := studentScoreMap[studentScoreKey]; ok {
//					studentContentScore = float32(studentScoreItem / contentScoreItem)
//				}
//			}
//
//			userOutcomeReply := make([]*v2.AssessmentStudentResultOutcomeReply, 0)
//			for _, outcomeID := range content.OutcomeIDs {
//				var userOutcome *v2.AssessmentUserOutcome
//				if assessmentContent, ok := contentMapFromAssessment[contentID]; ok {
//					key := o.at.GetKey([]string{
//						item.ID,
//						assessmentContent.ID,
//						outcomeID,
//					})
//					userOutcome = assessmentOutcomeMap[key]
//				}
//				userOutcomeReplyItem := &v2.AssessmentStudentResultOutcomeReply{
//					OutcomeID: outcomeID,
//				}
//				if o.at.first.Status == v2.AssessmentStatusInDraft ||
//					o.at.first.Status == v2.AssessmentStatusComplete {
//					if userOutcome != nil && userOutcome.Status != "" {
//						userOutcomeReplyItem.Status = userOutcome.Status
//					} else {
//						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusUnknown
//					}
//				} else {
//					outcomeInfo, ok := outcomeMapFromContent[outcomeID]
//					if !ok {
//						log.Warn(ctx, "outcome not found in content", log.Any("outcomeMapFromContent", outcomeMapFromContent), log.String("outcomeID", outcomeID))
//						continue
//					}
//					if outcomeInfo.Assumed || studentContentScore >= outcomeInfo.ScoreThreshold {
//						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusAchieved
//					} else {
//						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusUnknown
//					}
//				}
//
//				userOutcomeReply = append(userOutcomeReply, userOutcomeReplyItem)
//			}
//			resultReply.Outcomes = userOutcomeReply
//
//			roomKey := o.at.GetKey([]string{
//				item.UserID,
//				content.RoomProvideContentID,
//			})
//			if roomResultItem, ok := roomUserResultMap[roomKey]; ok {
//				resultReply.Answer = roomResultItem.Answer
//				resultReply.Score = roomResultItem.Score
//				resultReply.Attempted = roomResultItem.Seen
//			}
//
//			studentReply.Results = append(studentReply.Results, resultReply)
//		}
//
//		result = append(result, studentReply)
//	}
//
//	return result, nil
//}
//
//func (o *OnlineClassAssessment) MatchAnyOneAttempted() (bool, error) {
//	return o.base.MatchAnyOneAttempted()
//}
