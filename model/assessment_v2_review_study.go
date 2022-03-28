package model

import (
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func NewReviewStudyAssessmentPage(ags *AssessmentGrain) IAssessmentMatch {
	return &ReviewStudyAssessment{
		ags:    ags,
		action: AssessmentMatchActionPage,
		base:   NewBaseAssessment(ags),
	}
}

func NewReviewStudyAssessmentDetail(ags *AssessmentGrain) IAssessmentMatch {
	return &ReviewStudyAssessment{
		ags:    ags,
		action: AssessmentMatchActionDetail,
		base:   NewBaseAssessment(ags),
	}
}

type ReviewStudyAssessment struct {
	EmptyAssessment

	base   BaseAssessment
	ags    *AssessmentGrain
	action AssessmentMatchAction
}

func (o *ReviewStudyAssessment) MatchAnyOneAttempted() (bool, error) {
	return o.base.MatchAnyOneAttempted()
}

func (o *ReviewStudyAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	return o.base.MatchSchedule()
}

func (o *ReviewStudyAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	return NewOnlineStudyAssessmentPage(o.ags).MatchTeacher()
}

func (o *ReviewStudyAssessment) MatchClass() (map[string]*entity.IDName, error) {
	return o.base.MatchClass()
}

func (o *ReviewStudyAssessment) MatchCompleteRate() (map[string]float64, error) {
	return NewOnlineStudyAssessmentPage(o.ags).MatchCompleteRate()
}

func (o *ReviewStudyAssessment) MatchDiffContentStudents() ([]*v2.AssessmentDiffContentStudentsReply, error) {
	ctx := o.ags.ctx
	//op := adc.op

	assessmentUserMap, err := o.ags.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	assessmentUsers, ok := assessmentUserMap[o.ags.assessment.ID]
	if !ok {
		log.Warn(ctx, "assessment users is empty", log.Any("assessment", o.ags.assessment))
		return nil, constant.ErrRecordNotFound
	}

	userInfoMap, err := o.ags.GetUserMap()
	if err != nil {
		return nil, err
	}

	studentReviewMap, contentMap, err := o.ags.SingleGetContentsFromScheduleReview()
	if err != nil {
		return nil, err
	}

	roomDataMap, _ := o.ags.GetRoomData()
	roomData, ok := roomDataMap[o.ags.assessment.ScheduleID]
	if !ok {
		log.Warn(ctx, "not found room data", log.Any("roomDataMap", roomDataMap), log.Any("assessment", o.ags.assessment))
	}
	studentRoomDataMap := make(map[string]map[string]*UserRoomInfo)
	for _, item := range roomData {
		if item.User == nil {
			log.Warn(ctx, "room user data is empty")
			continue
		}
		userScoresTree, err := getAssessmentLiveRoom().getUserResultInfo(ctx, item)
		if err != nil {
			continue
		}

		studentRoomDataMap[item.User.UserID] = make(map[string]*UserRoomInfo)
		for _, userScoreItem := range userScoresTree {
			studentRoomDataMap[item.User.UserID][userScoreItem.MaterialID] = userScoreItem
		}
	}

	result := make([]*v2.AssessmentDiffContentStudentsReply, 0, len(studentReviewMap))
	for _, userItem := range assessmentUsers {
		if userItem.UserType != v2.AssessmentUserTypeStudent {
			continue
		}
		replyItem := &v2.AssessmentDiffContentStudentsReply{
			StudentID:       userItem.UserID,
			StudentName:     "",
			Status:          userItem.StatusByUser,
			ReviewerComment: "",
			Results:         make([]*v2.DiffContentStudentResultReply, 0),
		}

		if userInfo, ok := userInfoMap[userItem.UserID]; ok {
			replyItem.StudentName = userInfo.Name
		}

		if studentReviewItem, ok := studentReviewMap[userItem.UserID]; ok {
			if studentReviewItem.LiveLessonPlan == nil {
				log.Warn(ctx, "student review content empty", log.Any("studentReviewItem", studentReviewItem))
				continue
			}

			studentRoomDataItem, _ := studentRoomDataMap[userItem.UserID]

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
				if studentRoomDataItem != nil {
					if userContentRoomData, ok := studentRoomDataItem[contentItem.LessonMaterialID]; ok {
						reviewContentReplyItem.Score = userContentRoomData.Score
						reviewContentReplyItem.Answer = userContentRoomData.Answer
						reviewContentReplyItem.Attempted = userContentRoomData.Seen
						reviewContentReplyItem.Content.H5PID = userContentRoomData.H5PID
						reviewContentReplyItem.Content.ContentSubtype = userContentRoomData.SubContentType
						reviewContentReplyItem.Content.MaxScore = userContentRoomData.MaxScore
						reviewContentReplyItem.Content.H5PSubID = userContentRoomData.SubContentID

						if userContentRoomData.FileType == external.FileTypeH5P {
							if userContentRoomData.MaxScore > 0 {
								reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeSupportScoreStandAlone
							} else {
								reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeNotSupportScoreStandAlone
							}
						} else {
							reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeNotChildSubContainer
						}
						replyItem.Results = append(replyItem.Results, reviewContentReplyItem)
						if len(userContentRoomData.Children) > 0 {
							reviewContentReplyItem.Content.FileType = v2.AssessmentFileTypeHasChildContainer

							for i, child := range userContentRoomData.Children {
								o.appendStudentScore(child, contentItem, &replyItem.Results, reviewContentReplyItem.Content.Number, i+1)
							}
						}
					} else {
						replyItem.Results = append(replyItem.Results, reviewContentReplyItem)
					}
				} else {
					replyItem.Results = append(replyItem.Results, reviewContentReplyItem)
				}
			}

		}

		result = append(result, replyItem)
	}

	return result, nil
}

func (o *ReviewStudyAssessment) appendStudentScore(roomContent *UserRoomInfo, materialItem *entity.ScheduleLiveLessonMaterial, result *[]*v2.DiffContentStudentResultReply, prefix string, index int) {
	replyItem := &v2.DiffContentStudentResultReply{
		Answer:    roomContent.Answer,
		Score:     roomContent.Score,
		Attempted: roomContent.Seen,
		Content: v2.AssessmentDiffContentReply{
			Number:      fmt.Sprintf("%s-%d", prefix, index),
			ContentID:   materialItem.LessonMaterialID,
			ContentName: materialItem.LessonMaterialName,
			ContentType: v2.AssessmentContentTypeUnknown,
			FileType:    v2.AssessmentFileTypeNotUnknown,
			ParentID:    materialItem.LessonMaterialID,
			H5PID:       roomContent.H5PID,
			//ReviewerComment:      "",
			ContentSubtype:       roomContent.SubContentType,
			MaxScore:             roomContent.MaxScore,
			H5PSubID:             roomContent.SubContentID,
			RoomProvideContentID: "",
		},
	}

	if roomContent.FileType == external.FileTypeH5P {
		if canScoringMap[roomContent.SubContentType] || roomContent.MaxScore > 0 {
			replyItem.Content.FileType = v2.AssessmentFileTypeSupportScoreStandAlone
		} else {
			replyItem.Content.FileType = v2.AssessmentFileTypeNotSupportScoreStandAlone
		}
	}

	*result = append(*result, replyItem)
	for i, item := range roomContent.Children {
		o.appendStudentScore(item, materialItem, result, replyItem.Content.Number, i+1)
	}
}