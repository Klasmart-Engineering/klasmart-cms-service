package model

import (
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
	assessmentUserMap, err := o.ag.GetAssessmentUserMap()
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

	result := make(map[string]float64)
	//for _, item := range o.ag.assessments {
	//	if roomData, ok := roomDataMap[item.ScheduleID]; ok {
	//		result[item.ID] = getAssessmentLiveRoom().
	//			calcRoomCompleteRate(roomData, studentCount[item.ID])
	//	}
	//}

	return result, nil
}

func (o *OfflineStudyAssessment) MatchRemainingTime() (map[string]int64, error) {
	onlineStudy := NewOnlineStudyAssessmentPage(o.ag)

	return onlineStudy.MatchRemainingTime()
}
