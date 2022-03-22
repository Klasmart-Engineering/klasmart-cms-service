package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

func NewReviewStudyAssessment(ags *AssessmentsGrain) IAssessmentMatch {
	return &ReviewStudyAssessment{
		ctx: ags.ctx,
		op:  ags.op,
		ags: ags,
	}
}

//func NewOfflineClassAssessmentDetail(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) IAssessmentMatch {
//	return &OnlineClassAssessment{
//		ctx: ctx,
//		op:  op,
//		ags: NewAssessmentsGrain(ctx, op, []*v2.Assessment{assessment}),
//	}
//}

type ReviewStudyAssessment struct {
	EmptyAssessment

	ctx context.Context
	op  *entity.Operator
	ags *AssessmentsGrain
}

func (o *ReviewStudyAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	onlineClass := NewOnlineClassAssessment(o.ags)

	return onlineClass.MatchSchedule()
}

func (o *ReviewStudyAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	assessmentUserMap, err := o.ags.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	userMap, err := o.ags.GetUserMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*entity.IDName, len(o.ags.assessments))
	for _, item := range o.ags.assessments {
		if assUserItems, ok := assessmentUserMap[item.ID]; ok {
			for _, assUserItem := range assUserItems {
				if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
					continue
				}

				if userItem, ok := userMap[assUserItem.UserID]; ok && userItem != nil {
					result[item.ID] = append(result[item.ID], userItem)
				}
			}
		}
	}

	return result, nil
}

func (o *ReviewStudyAssessment) MatchClass() (map[string]*entity.IDName, error) {
	relationMap, err := o.ags.GetScheduleRelationMap()
	if err != nil {
		return nil, err
	}

	classMap, err := o.ags.GetClassMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*entity.IDName, len(o.ags.assessments))
	for _, item := range o.ags.assessments {
		if srItems, ok := relationMap[item.ScheduleID]; ok {
			for _, srItem := range srItems {
				if srItem.RelationType == entity.ScheduleRelationTypeClassRosterClass {
					result[item.ID] = classMap[srItem.RelationID]
					break
				}
			}
		}
	}

	return result, nil
}

func (o *ReviewStudyAssessment) MatchCompleteRate() (map[string]float64, error) {
	assessmentUserMap, err := o.ags.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	studentCount := make(map[string]int)
	for key, users := range assessmentUserMap {
		for _, userItem := range users {
			if userItem.UserType == v2.AssessmentUserTypeStudent {
				studentCount[key]++
			}
		}
	}

	roomDataMap, err := o.ags.GetRoomData()
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64)
	for _, item := range o.ags.assessments {
		if roomData, ok := roomDataMap[item.ScheduleID]; ok {
			result[item.ID] = getAssessmentLiveRoom().
				calcRoomCompleteRate(roomData, studentCount[item.ID])
		}
	}

	return result, nil
}
