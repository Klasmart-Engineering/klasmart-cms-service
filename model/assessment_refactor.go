package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

type IAssessmentMatch interface {
	FillPage(assessments []*v2.Assessment) ([]*v2.AssessmentQueryReply, error)
}

func NewOnlineClassAssessment(ctx context.Context, op *entity.Operator) IAssessmentMatch {
	return &OnlineClassAssessment{
		ctx: ctx,
		op:  op,
	}
}

type OnlineClassAssessment struct {
	ctx context.Context
	op  *entity.Operator
}

func (o *OnlineClassAssessment) MatchTeacher(ags *AssessmentsGrain) ([]*v2.AssessmentQueryReply, error) {
	// key: assessmentID
	assessmentUsers, err := ags.GetAssessmentUsers()
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(assessmentUsers))
	deDupMap := make(map[string]struct{})

	for _, auItem := range assessmentUsers {
		if _, ok := deDupMap[auItem.UserID]; !ok && auItem.StatusByUser == v2.AssessmentUserStatusParticipate {
			deDupMap[auItem.UserID] = struct{}{}
			userIDs = append(userIDs, auItem.UserID)
		}
	}

	userMap, err := ags.GetUserMap()
	if err != nil {
		return err
	}

	apc.assTeacherMap = make(map[string][]*entity.IDName, len(apc.assessments))
	for _, item := range apc.assessments {
		if assUserItems, ok := assessmentUserMap[item.ID]; ok {
			for _, assUserItem := range assUserItems {
				if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
					continue
				}
				if item.AssessmentType == v2.AssessmentTypeOnlineClass && assUserItem.StatusByUser == v2.AssessmentUserStatusNotParticipate {
					continue
				}

				if userItem, ok := userMap[assUserItem.UserID]; ok && userItem != nil {
					apc.assTeacherMap[item.ID] = append(apc.assTeacherMap[item.ID], userItem)
				}
			}
		}
	}

	return nil
}

func (o *OnlineClassAssessment) FillPage(assessments []*v2.Assessment) ([]*v2.AssessmentQueryReply, error) {
	ctx := o.ctx
	op := o.op

	result := make([]*v2.AssessmentQueryReply, len(assessments))
	ags := NewAssessmentsGrain(ctx, op, assessments)

	assessmentUsers, err := ags.GetAssessmentUsers()
	if err != nil {
		return nil, err
	}

	for _, item := range assessmentUsers {
		if item.UserType == v2.AssessmentUserTypeStudent &&
			item.StatusBySystem == v2.AssessmentUserStatusNotParticipate {
			continue
		}
		apc.assessmentUserMap[item.AssessmentID] = append(apc.assessmentUserMap[item.AssessmentID], item)
	}

	userMap, err := ags.GetUserMap(assessmentUsers)
	if err != nil {
		return nil, err
	}

	apc.assTeacherMap = make(map[string][]*entity.IDName, len(apc.assessments))
	for _, item := range apc.assessments {
		if assUserItems, ok := assessmentUserMap[item.ID]; ok {
			for _, assUserItem := range assUserItems {
				if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
					continue
				}
				if item.AssessmentType == v2.AssessmentTypeOnlineClass && assUserItem.StatusByUser == v2.AssessmentUserStatusNotParticipate {
					continue
				}

				if userItem, ok := userMap[assUserItem.UserID]; ok && userItem != nil {
					apc.assTeacherMap[item.ID] = append(apc.assTeacherMap[item.ID], userItem)
				}
			}
		}
	}

	for i, item := range assessments {
		replyItem := &v2.AssessmentQueryReply{
			ID:         item.ID,
			Title:      item.Title,
			ClassEndAt: item.ClassEndAt,
			CompleteAt: item.CompleteAt,
			Status:     item.Status,
		}
		result[i] = replyItem

		replyItem.Teachers = apc.assTeacherMap[item.ID]

		schedule, ok := apc.assScheduleMap[item.ID]
		if !ok {
			log.Warn(ctx, "not found assessment schedule", log.Any("assScheduleMap", apc.assScheduleMap), log.Any("assessmentItem", item))
			continue
		}
		if lessPlanItem, ok := apc.assLessPlanMap[item.ID]; ok {
			replyItem.LessonPlan = &entity.IDName{
				ID:   lessPlanItem.ID,
				Name: lessPlanItem.Name,
			}
		}

		replyItem.Program = apc.assProgramMap[item.ID]
		replyItem.Subjects = apc.assSubjectMap[item.ID]
		replyItem.DueAt = schedule.DueAt
		replyItem.ClassInfo = apc.assClassMap[item.ID]
		replyItem.RemainingTime = apc.assRemainingTimeMap[item.ID]
		replyItem.CompleteRate = apc.assCompleteRateMap[item.ID]
	}

	return result, nil
}
