package model

import (
	"context"
	"fmt"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da/assessmentV2"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
)

func NewReviewStudyAssessment() IAssessmentProcessor {
	return &ReviewStudyAssessment{}
}

type ReviewStudyAssessment struct{}

func (o *ReviewStudyAssessment) ProcessRemainingTime(ctx context.Context, dueAt int64, assessmentCreateAt int64) int64 {
	study := OfflineStudyAssessment{}
	return study.ProcessRemainingTime(ctx, dueAt, assessmentCreateAt)
}

func (o *ReviewStudyAssessment) Update(ctx context.Context, op *entity.Operator, assessment *v2.Assessment, req *v2.AssessmentUpdateReq) error {
	now := time.Now().Unix()

	at, err := NewAssessmentInit(ctx, op, assessment)
	if err != nil {
		return err
	}

	if err := at.initSchedule(); err != nil {
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

	remainingTime := o.ProcessRemainingTime(ctx, at.schedule.DueAt, assessment.CreateAt)
	if remainingTime > 0 {
		log.Warn(ctx, "assessment remaining time is greater than 0", log.Int64("remainingTime", remainingTime), log.Any("waitUpdateAssessment", assessment))
		return constant.ErrInvalidArgs
	}

	// user comment,score
	newScores := make([]*external.H5PSetScoreRequest, 0)

	contentReqMap := make(map[string]*v2.AssessmentUpdateContentReq)
	for _, item := range req.Contents {
		contentReqMap[item.ContentID] = item
	}

	for _, stuItem := range req.Students {
		if stuItem.Status == v2.AssessmentUserStatusNotParticipate {
			continue
		}
		// verify student data
		_, ok := userIDAndUserTypeMap[GetAssessmentKey([]string{stuItem.StudentID, v2.AssessmentUserTypeStudent.String()})]
		if !ok {
			log.Warn(ctx, "student not exist", log.Any("userIDAndUserTypeMap", userIDAndUserTypeMap), log.Any("stuItem", stuItem))
			return constant.ErrInvalidArgs
		}

		for _, stuResult := range stuItem.Results {
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

	// update student scores
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

	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		if _, err := assessmentV2.GetAssessmentDA().UpdateTx(ctx, tx, assessment); err != nil {
			return err
		}

		if len(waitUpdatedUsers) > 0 {
			if _, err := assessmentV2.GetAssessmentUserDA().UpdateTx(ctx, tx, waitUpdatedUsers); err != nil {
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

func (o *ReviewStudyAssessment) ProcessTeacherID(assUserItem *v2.AssessmentUser) (string, bool) {
	if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
		return "", false
	}

	return assUserItem.UserID, true
}

func (o *ReviewStudyAssessment) ProcessCompleteRate(ctx context.Context, assessmentUsers []*v2.AssessmentUser,
	roomData *external.RoomInfo, stuReviewMap map[string]*entity.ScheduleReview, reviewerFeedbackMap map[string]*v2.AssessmentReviewerFeedback) float64 {

	studentCount := 0
	for _, userItem := range assessmentUsers {
		if userItem.UserType == v2.AssessmentUserTypeStudent {
			studentCount++
		}
	}

	var contentTotalCount int
	for _, stuContentItem := range stuReviewMap {
		if stuContentItem.LiveLessonPlan == nil {
			log.Warn(ctx, "student content is empty", log.Any("stuContentItem", stuContentItem))
			continue
		}
		contentTotalCount += len(stuContentItem.LiveLessonPlan.LessonMaterials)
	}

	return GetAssessmentExternalService().calcRoomCompleteRateWhenUseDiffContent(ctx, roomData.ScoresByUser, contentTotalCount)
}

func (o *ReviewStudyAssessment) ProcessTeacherName(assUserItem *v2.AssessmentUser, teacherMap map[string]*entity.IDName) (*entity.IDName, bool) {
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

func (o *ReviewStudyAssessment) ProcessContents(ctx context.Context, at *AssessmentInit) ([]*v2.AssessmentContentReply, error) {
	return make([]*v2.AssessmentContentReply, 0), nil
}

func (o *ReviewStudyAssessment) ProcessDiffContents(ctx context.Context, at *AssessmentInit) []*v2.AssessmentDiffContentStudentsReply {
	assessmentUsers := at.assessmentUsers
	studentReviewMap := at.scheduleStuReviewMap
	contentMap := at.contentMapFromScheduleReview
	userScores := at.roomUserScoreMap
	roomContentTree := at.roomContentTree
	commentResultMap := at.commentResultMap

	studentRoomDataMap := make(map[string]map[string]*RoomUserScore)
	roomContentMap := make(map[string]*RoomContentTree)
	if len(userScores) > 0 {
		for _, contentItem := range roomContentTree {
			roomContentMap[contentItem.ContentUniqueID] = contentItem
		}
		for userID, scores := range userScores {
			studentRoomDataMap[userID] = make(map[string]*RoomUserScore)
			for _, scoreItem := range scores {
				studentRoomDataMap[userID][scoreItem.ContentUniqueID] = scoreItem
			}
		}
	} else {
		log.Warn(ctx, "not found room data", log.Any("studentReviewMap", studentReviewMap))
	}

	result := make([]*v2.AssessmentDiffContentStudentsReply, 0, len(studentReviewMap))
	for _, userItem := range assessmentUsers {
		if userItem.UserType != v2.AssessmentUserTypeStudent {
			continue
		}
		replyItem := &v2.AssessmentDiffContentStudentsReply{
			StudentID: userItem.UserID,
			//StudentName:     "",
			Status:          userItem.StatusByUser,
			ReviewerComment: "",
			Results:         make([]*v2.DiffContentStudentResultReply, 0),
		}

		if comment, ok := commentResultMap[userItem.ID]; ok {
			replyItem.ReviewerComment = comment
		} else {
			replyItem.ReviewerComment = commentResultMap[userItem.UserID]
		}

		if studentReviewItem, ok := studentReviewMap[userItem.UserID]; ok {
			if studentReviewItem.LiveLessonPlan == nil {
				log.Warn(ctx, "student review content empty", log.Any("studentReviewItem", studentReviewItem))
				continue
			}

			studentContentScoreMap, ok := studentRoomDataMap[userItem.UserID]
			if !ok {
				log.Warn(ctx, "not found user room data", log.String("userID", userItem.UserID), log.Any("studentRoomDataMap", studentRoomDataMap))
			}

			index := 0
			for _, contentItem := range studentReviewItem.LiveLessonPlan.LessonMaterials {
				index++
				reviewContentReplyItem := &v2.DiffContentStudentResultReply{
					Answer:    "",
					Score:     0,
					Attempted: false,
					Content: v2.AssessmentDiffContentReply{
						Number:               fmt.Sprintf("%d", index),
						ContentID:            contentItem.LessonMaterialID,
						ContentName:          contentItem.LessonMaterialName,
						ContentType:          v2.AssessmentContentTypeLessonMaterial,
						FileType:             v2.AssessmentFileTypeNotChildSubContainer,
						ParentID:             "",
						H5PID:                "",
						ContentSubtype:       "",
						MaxScore:             0,
						H5PSubID:             "",
						RoomProvideContentID: "",
					},
				}
				if contentInfo, ok := contentMap[contentItem.LessonMaterialID]; ok {
					reviewContentReplyItem.Content.ContentSubtype = contentInfo.FileType.String()
				}
				if userContentRoomData, ok := studentContentScoreMap[contentItem.LessonMaterialID]; ok {
					roomContentItem, ok := roomContentMap[userContentRoomData.ContentUniqueID]
					if !ok {
						log.Warn(ctx, "user content Data not found", log.Any("roomContentMap", roomContentMap), log.Any("userContentRoomData", userContentRoomData))
						continue
					}
					reviewContentReplyItem.Score = userContentRoomData.Score
					reviewContentReplyItem.Answer = userContentRoomData.Answer
					reviewContentReplyItem.Attempted = userContentRoomData.Seen
					reviewContentReplyItem.Content.H5PID = roomContentItem.H5PID
					reviewContentReplyItem.Content.ContentSubtype = roomContentItem.Type
					reviewContentReplyItem.Content.MaxScore = roomContentItem.MaxScore
					reviewContentReplyItem.Content.H5PSubID = roomContentItem.SubContentID

					if roomContentItem.FileType == external.FileTypeH5P {
						if roomContentItem.MaxScore > 0 {
							reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeSupportScoreStandAlone
						} else {
							reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeNotSupportScoreStandAlone
						}
					} else {
						reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeNotChildSubContainer
					}
					replyItem.Results = append(replyItem.Results, reviewContentReplyItem)
					if len(roomContentItem.Children) > 0 {
						reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeHasChildContainer

						for i, child := range roomContentItem.Children {
							o.appendStudentScore(child, studentContentScoreMap, contentItem, &replyItem.Results, reviewContentReplyItem.Content.Number, i+1)
						}
					}
				} else {
					replyItem.Results = append(replyItem.Results, reviewContentReplyItem)
				}
			}

		}

		result = append(result, replyItem)
	}

	return result
}

func (o *ReviewStudyAssessment) ProcessStudents(ctx context.Context, at *AssessmentInit, contentsReply []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error) {
	return make([]*v2.AssessmentStudentReply, 0), nil
}

func (o *ReviewStudyAssessment) appendStudentScore(roomContent *RoomContentTree, userContentScoreMap map[string]*RoomUserScore, materialItem *entity.ScheduleLiveLessonMaterial, result *[]*v2.DiffContentStudentResultReply, prefix string, index int) {
	replyItem := &v2.DiffContentStudentResultReply{
		Content: v2.AssessmentDiffContentReply{
			Number:      fmt.Sprintf("%s-%d", prefix, index),
			ContentID:   materialItem.LessonMaterialID,
			ContentName: materialItem.LessonMaterialName,
			ContentType: v2.AssessmentContentTypeUnknown,
			FileType:    v2.AssessmentFileTypeNotUnknown,
			ParentID:    materialItem.LessonMaterialID,
			H5PID:       roomContent.H5PID,
			//ReviewerComment:      "",
			ContentSubtype:       roomContent.Type,
			MaxScore:             roomContent.MaxScore,
			H5PSubID:             roomContent.SubContentID,
			RoomProvideContentID: "",
		},
	}
	if userScore, ok := userContentScoreMap[roomContent.ContentUniqueID]; ok {
		replyItem.Attempted = userScore.Seen
		replyItem.Score = userScore.Score
		replyItem.Answer = userScore.Answer
	}

	if roomContent.FileType == external.FileTypeH5P {
		if roomContent.MaxScore > 0 {
			replyItem.Content.FileType = v2.AssessmentFileTypeSupportScoreStandAlone
		} else {
			replyItem.Content.FileType = v2.AssessmentFileTypeNotSupportScoreStandAlone
		}
	}

	if len(roomContent.Children) > 0 {
		replyItem.Content.FileType = v2.AssessmentFileTypeHasChildContainer
	}

	*result = append(*result, replyItem)
	for i, item := range roomContent.Children {
		o.appendStudentScore(item, userContentScoreMap, materialItem, result, replyItem.Content.Number, i+1)
	}
}

// old
//
//func (o *ReviewStudyAssessment) MatchCompleteRate() map[string]float64 {
//	ctx := o.ali.ctx
//
//	assessmentUserMap := o.ali.assessmentUserMap
//
//	studentCount := make(map[string]int)
//	for key, users := range assessmentUserMap {
//		for _, userItem := range users {
//			if userItem.UserType == v2.AssessmentUserTypeStudent {
//				studentCount[key]++
//			}
//		}
//	}
//
//	roomDataMap := o.ali.liveRoomMap
//
//	// scheduleID,studentID
//	scheduleReviewMap := o.ali.scheduleStuReviewMap
//
//	result := make(map[string]float64)
//	for _, item := range o.assessmentMap {
//		studentReviewContentMap, ok := scheduleReviewMap[item.ScheduleID]
//		if !ok {
//			log.Warn(ctx, "not found student content in schedule", log.Any("schedule", item))
//			continue
//		}
//
//		roomData, ok := roomDataMap[item.ScheduleID]
//		if !ok {
//			continue
//		}
//
//		var contentTotalCount int
//		for _, stuContentItem := range studentReviewContentMap {
//			if stuContentItem.LiveLessonPlan == nil {
//				log.Warn(ctx, "student content is empty", log.Any("stuContentItem", stuContentItem))
//				continue
//			}
//			contentTotalCount += len(stuContentItem.LiveLessonPlan.LessonMaterials)
//		}
//
//		result[item.ID] = GetAssessmentExternalService().calcRoomCompleteRateWhenUseDiffContent(ctx, roomData.ScoresByUser, contentTotalCount)
//	}
//
//	return result
//}
//
//func (o *ReviewStudyAssessment) MatchAnyOneAttempted() (bool, error) {
//	return o.base.MatchAnyOneAttempted()
//}
//
//func (o *ReviewStudyAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
//	return o.base.MatchSchedule()
//}
//
//func (o *ReviewStudyAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
//	return o.base.MatchTeacher()
//}
//
//func (o *ReviewStudyAssessment) MatchClass() (map[string]*entity.IDName, error) {
//	return o.base.MatchClass()
//}
//
//func (o *ReviewStudyAssessment) MatchRemainingTime() (map[string]int64, error) {
//	onlineStudy := NewOnlineStudyAssessment(o.at, o.action)
//
//	return onlineStudy.MatchRemainingTime()
//}
//
//func (o *ReviewStudyAssessment) MatchDiffContentStudents() ([]*v2.AssessmentDiffContentStudentsReply, error) {
//	ctx := o.at.ctx
//	//op := adc.op
//
//	assessmentUserMap, err := o.at.GetAssessmentUserMap()
//	if err != nil {
//		return nil, err
//	}
//
//	assessmentUsers, ok := assessmentUserMap[o.at.first.ID]
//	if !ok {
//		log.Warn(ctx, "assessment users is empty", log.Any("assessment", o.at.first))
//		return nil, constant.ErrRecordNotFound
//	}
//
//	studentReviewMap, err := o.at.FirstGetScheduleReviewMap()
//	if err != nil {
//		return nil, err
//	}
//
//	contentMap, err := o.at.FirstGetContentsFromScheduleReview()
//	if err != nil {
//		return nil, err
//	}
//
//	roomDataMap, err := o.at.GetExternalAssessmentServiceData()
//	if err != nil {
//		return nil, err
//	}
//
//	commentResultMap, err := o.at.FirstGetCommentResultMap()
//	if err != nil {
//		return nil, err
//	}
//
//	roomData, ok := roomDataMap[o.at.first.ScheduleID]
//	studentRoomDataMap := make(map[string]map[string]*RoomUserScore)
//	roomContentMap := make(map[string]*RoomContentTree)
//	if ok {
//		userScores, roomContentTree, err := GetAssessmentExternalService().StudentScores(ctx, roomData.ScoresByUser)
//		if err != nil {
//			return nil, err
//		}
//
//		for _, contentItem := range roomContentTree {
//			roomContentMap[contentItem.ContentUniqueID] = contentItem
//		}
//		for userID, scores := range userScores {
//			studentRoomDataMap[userID] = make(map[string]*RoomUserScore)
//			for _, scoreItem := range scores {
//				studentRoomDataMap[userID][scoreItem.ContentUniqueID] = scoreItem
//			}
//		}
//	} else {
//		log.Warn(ctx, "not found room data", log.Any("roomDataMap", roomDataMap), log.Any("assessment", o.at.first))
//	}
//
//	log.Debug(ctx, "student room data", log.Any("studentRoomDataMap", studentRoomDataMap))
//
//	result := make([]*v2.AssessmentDiffContentStudentsReply, 0, len(studentReviewMap))
//	for _, userItem := range assessmentUsers {
//		if userItem.UserType != v2.AssessmentUserTypeStudent {
//			continue
//		}
//		replyItem := &v2.AssessmentDiffContentStudentsReply{
//			StudentID: userItem.UserID,
//			//StudentName:     "",
//			Status:          userItem.StatusByUser,
//			ReviewerComment: "",
//			Results:         make([]*v2.DiffContentStudentResultReply, 0),
//		}
//
//		if comment, ok := commentResultMap[userItem.ID]; ok {
//			replyItem.ReviewerComment = comment
//		} else {
//			replyItem.ReviewerComment = commentResultMap[userItem.UserID]
//		}
//
//		if studentReviewItem, ok := studentReviewMap[userItem.UserID]; ok {
//			if studentReviewItem.LiveLessonPlan == nil {
//				log.Warn(ctx, "student review content empty", log.Any("studentReviewItem", studentReviewItem))
//				continue
//			}
//
//			studentContentScoreMap, ok := studentRoomDataMap[userItem.UserID]
//			if !ok {
//				log.Warn(ctx, "not found user room data", log.String("userID", userItem.UserID), log.Any("studentRoomDataMap", studentRoomDataMap))
//			}
//
//			index := 0
//			for _, contentItem := range studentReviewItem.LiveLessonPlan.LessonMaterials {
//				index++
//				reviewContentReplyItem := &v2.DiffContentStudentResultReply{
//					Answer:    "",
//					Score:     0,
//					Attempted: false,
//					Content: v2.AssessmentDiffContentReply{
//						Number:               fmt.Sprintf("%d", index),
//						ContentID:            contentItem.LessonMaterialID,
//						ContentName:          contentItem.LessonMaterialName,
//						ContentType:          v2.AssessmentContentTypeLessonMaterial,
//						FileType:             v2.AssessmentFileTypeNotChildSubContainer,
//						ParentID:             "",
//						H5PID:                "",
//						ContentSubtype:       "",
//						MaxScore:             0,
//						H5PSubID:             "",
//						RoomProvideContentID: "",
//					},
//				}
//				if contentInfo, ok := contentMap[contentItem.LessonMaterialID]; ok {
//					reviewContentReplyItem.Content.ContentSubtype = contentInfo.FileType.String()
//				}
//				if userContentRoomData, ok := studentContentScoreMap[contentItem.LessonMaterialID]; ok {
//					roomContentItem, ok := roomContentMap[userContentRoomData.ContentUniqueID]
//					if !ok {
//						log.Warn(ctx, "user content Data not found", log.Any("roomContentMap", roomContentMap), log.Any("userContentRoomData", userContentRoomData))
//						continue
//					}
//					reviewContentReplyItem.Score = userContentRoomData.Score
//					reviewContentReplyItem.Answer = userContentRoomData.Answer
//					reviewContentReplyItem.Attempted = userContentRoomData.Seen
//					reviewContentReplyItem.Content.H5PID = roomContentItem.H5PID
//					reviewContentReplyItem.Content.ContentSubtype = roomContentItem.Type
//					reviewContentReplyItem.Content.MaxScore = roomContentItem.MaxScore
//					reviewContentReplyItem.Content.H5PSubID = roomContentItem.SubContentID
//
//					if roomContentItem.FileType == external.FileTypeH5P {
//						if roomContentItem.MaxScore > 0 {
//							reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeSupportScoreStandAlone
//						} else {
//							reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeNotSupportScoreStandAlone
//						}
//					} else {
//						reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeNotChildSubContainer
//					}
//					replyItem.Results = append(replyItem.Results, reviewContentReplyItem)
//					if len(roomContentItem.Children) > 0 {
//						reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeHasChildContainer
//
//						for i, child := range roomContentItem.Children {
//							o.appendStudentScore(child, studentContentScoreMap, contentItem, &replyItem.Results, reviewContentReplyItem.Content.Number, i+1)
//						}
//					}
//				} else {
//					replyItem.Results = append(replyItem.Results, reviewContentReplyItem)
//				}
//			}
//
//		}
//
//		result = append(result, replyItem)
//	}
//
//	return result, nil
//}
