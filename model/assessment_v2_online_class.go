package model

import (
	"fmt"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
)

func NewOnlineClassAssessment(at *AssessmentTool, action AssessmentMatchAction) IAssessmentMatch {
	return &OnlineClassAssessment{
		at:     at,
		action: action,
		base:   NewBaseAssessment(at, action),
	}
}

type OnlineClassAssessment struct {
	EmptyAssessment

	base   BaseAssessment
	at     *AssessmentTool
	action AssessmentMatchAction
}

func (o *OnlineClassAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	return o.base.MatchSchedule()
}

func (o *OnlineClassAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	return o.base.MatchLessPlan()
}

func (o *OnlineClassAssessment) MatchProgram() (map[string]*entity.IDName, error) {
	scheduleMap, err := o.at.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	programMap, err := o.at.GetProgramMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*entity.IDName, len(o.at.assessments))
	for _, item := range o.at.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			result[item.ID] = programMap[schedule.ProgramID]
		}
	}

	return result, nil
}

func (o *OnlineClassAssessment) MatchSubject() (map[string][]*entity.IDName, error) {
	relationMap, err := o.at.GetScheduleRelationMap()
	if err != nil {
		return nil, err
	}

	subjectMap, err := o.at.GetSubjectMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*entity.IDName, len(o.at.assessments))
	for _, item := range o.at.assessments {
		if srItems, ok := relationMap[item.ScheduleID]; ok {
			for _, srItem := range srItems {
				if srItem.RelationType != entity.ScheduleRelationTypeSubject {
					continue
				}
				if subItem, ok := subjectMap[srItem.RelationID]; ok && subItem != nil {
					result[item.ID] = append(result[item.ID], subItem)
				}
			}
		}
	}

	return result, nil
}

func (o *OnlineClassAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	assessmentUserMap, err := o.at.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]*entity.IDName)
	if o.action == AssessmentMatchActionPage {
		userMap, err = o.at.GetTeacherMap()
		if err != nil {
			return nil, err
		}
	}

	result := make(map[string][]*entity.IDName, len(o.at.assessments))
	for _, item := range o.at.assessments {
		if assUserItems, ok := assessmentUserMap[item.ID]; ok {
			for _, assUserItem := range assUserItems {
				if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
					continue
				}
				if assUserItem.StatusBySystem == v2.AssessmentUserSystemStatusNotStarted {
					continue
				}
				resultItem := &entity.IDName{
					ID:   assUserItem.UserID,
					Name: "",
				}

				if userItem, ok := userMap[assUserItem.UserID]; ok && userItem != nil {
					resultItem.Name = userItem.Name
				}
				result[item.ID] = append(result[item.ID], resultItem)
			}
		}
	}

	return result, nil
}

func (o *OnlineClassAssessment) MatchClass() (map[string]*entity.IDName, error) {
	if o.action == AssessmentMatchActionPage {
		return o.EmptyAssessment.MatchClass()
	}
	return o.base.MatchClass()
}

func (o *OnlineClassAssessment) MatchOutcomes() (map[string]*v2.AssessmentOutcomeReply, error) {
	contents, err := o.at.FirstGetContentsFromSchedule()
	if err != nil {
		return nil, err
	}

	outcomeMap, err := o.at.FirstGetOutcomeMapFromContent()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*v2.AssessmentOutcomeReply, len(outcomeMap))

	for _, item := range outcomeMap {
		result[item.ID] = &v2.AssessmentOutcomeReply{
			OutcomeID:          item.ID,
			OutcomeName:        item.Name,
			AssignedTo:         nil,
			Assumed:            item.Assumed,
			AssignedToLessPlan: false,
			AssignedToMaterial: false,
			ScoreThreshold:     item.ScoreThreshold,
		}
	}

	for _, materialItem := range contents {
		for _, outcomeID := range materialItem.OutcomeIDs {
			if outcomeItem, ok := result[outcomeID]; ok {
				if materialItem.ContentType == v2.AssessmentContentTypeLessonPlan {
					outcomeItem.AssignedToLessPlan = true
				}
				if materialItem.ContentType == v2.AssessmentContentTypeLessonMaterial {
					outcomeItem.AssignedToMaterial = true
				}
			}
		}
	}

	for _, outcomeItem := range result {
		if outcomeItem.AssignedToLessPlan {
			outcomeItem.AssignedTo = append(outcomeItem.AssignedTo, v2.AssessmentOutcomeAssignTypeLessonPlan)
		}
		if outcomeItem.AssignedToMaterial {
			outcomeItem.AssignedTo = append(outcomeItem.AssignedTo, v2.AssessmentOutcomeAssignTypeLessonMaterial)
		}
	}

	return result, nil
}

func (o *OnlineClassAssessment) MatchContents() ([]*v2.AssessmentContentReply, error) {
	libraryContents, err := o.at.FirstGetContentsFromSchedule()
	if err != nil {
		return nil, err
	}

	assessmentContentMap, err := o.at.FirstGetAssessmentContentMap()
	if err != nil {
		return nil, err
	}

	roomContentMap, err := o.at.FirstGetContentMapFromLiveRoom()
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
			ContentSubtype:  item.FileType.String(),
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

		if roomContentItem, ok := roomContentMap[item.ID]; ok {
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
						log.Warn(o.at.ctx, "sub content id is empty", log.Any("contentItem", item))
						continue
					}
					o.appendContent(child, contentReplyItem, &result, contentReplyItem.Number, i+1)
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

func (o *OnlineClassAssessment) MatchStudents(contentsReply []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error) {
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

	userScoresMap, _, err := o.at.FirstGetRoomData()
	if err != nil {
		return nil, err
	}

	roomUserResultMap := make(map[string]*RoomUserScore)
	for userID, scores := range userScoresMap {
		for _, scoreItem := range scores {
			key := o.at.GetKey([]string{
				userID,
				scoreItem.ContentUniqueID,
			})
			roomUserResultMap[key] = scoreItem
		}
	}

	contentScoreMap, studentScoreMap := o.at.summaryRoomScores(userScoresMap, contentsReply)

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

func (o *OnlineClassAssessment) appendContent(roomContent *RoomContentTree, materialItem *v2.AssessmentContentReply, result *[]*v2.AssessmentContentReply, prefix string, index int) {
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
			log.Warn(o.at.ctx, "sub content id is empty", log.Any("contentItem", item))
			continue
		}
		o.appendContent(item, materialItem, result, replyItem.Number, i+1)
	}
}

func (o *OnlineClassAssessment) MatchAnyOneAttempted() (bool, error) {
	return o.base.MatchAnyOneAttempted()
}
