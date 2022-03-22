package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"time"
)

func NewOnlineStudyAssessment(ags *AssessmentsGrain) IAssessmentMatch {
	return &OnlineStudyAssessment{
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

type OnlineStudyAssessment struct {
	EmptyAssessment

	ctx context.Context
	op  *entity.Operator
	ags *AssessmentsGrain
}

func (o *OnlineStudyAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	onlineClass := NewOnlineClassAssessment(o.ags)

	return onlineClass.MatchSchedule()
}

func (o *OnlineStudyAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	onlineClass := NewOnlineClassAssessment(o.ags)

	return onlineClass.MatchLessPlan()
}

func (o *OnlineStudyAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
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

func (o *OnlineStudyAssessment) MatchClass() (map[string]*entity.IDName, error) {
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

func (o *OnlineStudyAssessment) MatchCompleteRate() (map[string]float64, error) {
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

func (o *OnlineStudyAssessment) MatchRemainingTime() (map[string]int64, error) {
	scheduleMap, err := o.ags.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, item := range o.ags.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			var remainingTime int64
			if schedule.DueAt != 0 {
				remainingTime = schedule.DueAt - time.Now().Unix()
			} else {
				remainingTime = time.Unix(item.CreateAt, 0).Add(config.Get().Assessment.DefaultRemainingTime).Unix() - time.Now().Unix()
			}
			if remainingTime < 0 {
				remainingTime = 0
			}
			result[item.ID] = remainingTime
		}
	}

	return result, nil
}
