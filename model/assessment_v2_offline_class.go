package model

import (
	"context"
	"fmt"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da/assessmentV2"
	"time"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
)

func NewOfflineClassAssessment() IAssessmentProcessor {
	return &OfflineClassAssessment{}
}

type OfflineClassAssessment struct{}

func (o *OfflineClassAssessment) Update(ctx context.Context, op *entity.Operator, assessment *v2.Assessment, req *v2.AssessmentUpdateReq) error {
	now := time.Now().Unix()

	at, err := NewAssessmentInit(ctx, op, assessment)
	if err != nil {
		return err
	}

	if err := at.initAssessmentUserWithIDTypeMap(); err != nil {
		return err
	}

	userIDAndUserTypeMap := at.assessmentUserIDTypeMap

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

		waitUpdatedUsers = append(waitUpdatedUsers, existItem)
	}

	if err := at.initAssessmentContentMap(); err != nil {
		return err
	}

	//scheduleContents := at.contentsFromSchedule
	assessmentContentMap := at.contentMapFromAssessment

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
		}
	}

	// outcome

	if err := at.initOutcomeFromAssessment(); err != nil {
		return err
	}
	outcomeFromAssessmentMap := at.outcomeMapFromAssessment

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
			if assessmentContentItem, ok := assessmentContentMap[stuResult.ContentID]; ok {
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
					} else {
						log.Warn(ctx, "student outcome invalid", log.Any("outcomeFromAssessmentMap", outcomeFromAssessmentMap), log.Any("stuItem", stuItem))
						continue
					}
				}
			}
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
			if _, err = assessmentV2.GetAssessmentUserDA().UpdateTx(ctx, tx, waitUpdatedUsers); err != nil {
				return err
			}
		}

		if len(waitUpdateContents) > 0 {
			if _, err = assessmentV2.GetAssessmentContentDA().UpdateTx(ctx, tx, waitUpdateContents); err != nil {
				return err
			}
		}

		if len(waitUpdateAssessmentOutcomes) > 0 {
			if _, err = assessmentV2.GetAssessmentUserOutcomeDA().UpdateTx(ctx, tx, waitUpdateAssessmentOutcomes); err != nil {
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

func (o *OfflineClassAssessment) ProcessTeacherID(assUserItem *v2.AssessmentUser) (string, bool) {
	if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
		return "", false
	}

	return assUserItem.UserID, true
}

func (o *OfflineClassAssessment) ProcessContents(ctx context.Context, at *AssessmentInit) ([]*v2.AssessmentContentReply, error) {
	libraryContents := at.contentsFromSchedule

	assessmentContentMap := at.contentMapFromAssessment

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

func (o *OfflineClassAssessment) ProcessStudents(ctx context.Context, at *AssessmentInit, contents []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error) {
	assessmentUsers := at.assessmentUsers

	contentMapFromAssessment := at.contentMapFromAssessment

	outcomeMapFromAssessment := at.outcomeMapFromAssessment

	outcomeMapFromContent := at.outcomeMapFromContent

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
					key := GetAssessmentKey([]string{
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

func (o *OfflineClassAssessment) ProcessDiffContents(ctx context.Context, at *AssessmentInit) []*v2.AssessmentDiffContentStudentsReply {
	return make([]*v2.AssessmentDiffContentStudentsReply, 0)
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

func (o *OfflineClassAssessment) ProcessRemainingTime(ctx context.Context, dueAt int64, assessmentCreateAt int64) int64 {
	return 0
}

// old

//func (o *OfflineClassAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
//	return o.base.MatchSchedule()
//}
//
//func (o *OfflineClassAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
//	return o.base.MatchLessPlan()
//}
//
//func (o *OfflineClassAssessment) MatchProgram() (map[string]*entity.IDName, error) {
//	onlineClass := NewOnlineClassAssessment(o.at, o.action)
//
//	return onlineClass.MatchProgram()
//}
//
//func (o *OfflineClassAssessment) MatchSubject() (map[string][]*entity.IDName, error) {
//	onlineClass := NewOnlineClassAssessment(o.at, o.action)
//
//	return onlineClass.MatchSubject()
//}
//
//func (o *OfflineClassAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
//	return o.base.MatchTeacher()
//}
//
//func (o *OfflineClassAssessment) MatchClass() (map[string]*entity.IDName, error) {
//	onlineClass := NewOnlineClassAssessment(o.at, o.action)
//
//	return onlineClass.MatchClass()
//}
//
//func (o *OfflineClassAssessment) MatchOutcomes() (map[string]*v2.AssessmentOutcomeReply, error) {
//	onlineClass := NewOnlineClassAssessment(o.at, o.action)
//
//	return onlineClass.MatchOutcomes()
//}
//
//func (o *OfflineClassAssessment) MatchContents() ([]*v2.AssessmentContentReply, error) {
//	libraryContents, err := o.at.FirstGetContentsFromSchedule()
//	if err != nil {
//		return nil, err
//	}
//
//	assessmentContentMap, err := o.at.FirstGetAssessmentContentMap()
//	if err != nil {
//		return nil, err
//	}
//
//	result := make([]*v2.AssessmentContentReply, 0)
//	index := 0
//	for _, item := range libraryContents {
//		contentReplyItem := &v2.AssessmentContentReply{
//			Number:          "0",
//			ParentID:        "",
//			ContentID:       item.ID,
//			ContentName:     item.Name,
//			Status:          v2.AssessmentContentStatusCovered,
//			ContentType:     item.ContentType,
//			FileType:        v2.AssessmentFileTypeNotChildSubContainer,
//			MaxScore:        0,
//			ReviewerComment: "",
//			OutcomeIDs:      item.OutcomeIDs,
//		}
//
//		if item.ContentType == v2.AssessmentContentTypeLessonPlan {
//			contentReplyItem.FileType = v2.AssessmentFileTypeHasChildContainer
//			result = append(result, contentReplyItem)
//			continue
//		}
//
//		index++
//		contentReplyItem.Number = fmt.Sprintf("%d", index)
//
//		if assessmentContentItem, ok := assessmentContentMap[item.ID]; ok {
//			contentReplyItem.ReviewerComment = assessmentContentItem.ReviewerComment
//			contentReplyItem.Status = assessmentContentItem.Status
//		}
//
//		result = append(result, contentReplyItem)
//	}
//
//	return result, nil
//}
//
//func (o *OfflineClassAssessment) MatchStudents(contents []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error) {
//	//ctx := o.at.ctx
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
//	contentMapFromAssessment, err := o.at.FirstGetAssessmentContentMap()
//	if err != nil {
//		return nil, err
//	}
//
//	outcomeMapFromAssessment, err := o.at.FirstGetOutcomeFromAssessment()
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
//
//		studentReply := &v2.AssessmentStudentReply{
//			StudentID: item.UserID,
//			//StudentName:   "",
//			Status:        item.StatusByUser,
//			ProcessStatus: item.StatusBySystem,
//			Results:       nil,
//		}
//
//		for _, content := range contents {
//			resultReply := &v2.AssessmentStudentResultReply{
//				ContentID: content.ContentID,
//			}
//
//			userOutcomeReply := make([]*v2.AssessmentStudentResultOutcomeReply, 0)
//			for _, outcomeID := range content.OutcomeIDs {
//				var userOutcome *v2.AssessmentUserOutcome
//				if assessmentContent, ok := contentMapFromAssessment[content.ContentID]; ok {
//					key := o.at.GetKey([]string{
//						item.ID,
//						assessmentContent.ID,
//						outcomeID,
//					})
//					userOutcome = outcomeMapFromAssessment[key]
//				}
//				userOutcomeReplyItem := &v2.AssessmentStudentResultOutcomeReply{
//					OutcomeID: outcomeID,
//				}
//				if userOutcome != nil && userOutcome.Status != "" {
//					userOutcomeReplyItem.Status = userOutcome.Status
//				} else {
//					if outcomeInfo, ok := outcomeMapFromContent[outcomeID]; ok && outcomeInfo.Assumed {
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
//			studentReply.Results = append(studentReply.Results, resultReply)
//		}
//
//		result = append(result, studentReply)
//	}
//
//	return result, nil
//}
