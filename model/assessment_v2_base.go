package model

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

type IAssessmentMatch interface {
	MatchSchedule() (map[string]*entity.Schedule, error)
	MatchTeacher() (map[string][]*entity.IDName, error)
	MatchLessPlan() (map[string]*v2.AssessmentContentView, error)
	MatchProgram() (map[string]*entity.IDName, error)
	MatchSubject() (map[string][]*entity.IDName, error)
	MatchClass() (map[string]*entity.IDName, error)
	MatchCompleteRate() (map[string]float64, error)
	MatchRemainingTime() (map[string]int64, error)
}

func NewEmptyAssessment() IAssessmentMatch {
	return &EmptyAssessment{}
}

//func NewEmptyAssessmentDetail(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) IAssessmentMatch {
//	return &EmptyAssessment{
//		ctx: ctx,
//		op:  op,
//		ags: NewAssessmentsGrain(ctx, op, []*v2.Assessment{assessment}),
//	}
//}

type EmptyAssessment struct {
}

func (o *EmptyAssessment) MatchClass() (map[string]*entity.IDName, error) {
	return make(map[string]*entity.IDName), nil
}

func (o *EmptyAssessment) MatchCompleteRate() (map[string]float64, error) {
	return make(map[string]float64), nil
}

func (o *EmptyAssessment) MatchRemainingTime() (map[string]int64, error) {
	return make(map[string]int64), nil
}

func (o *EmptyAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	return make(map[string]*entity.Schedule), nil
}

func (o *EmptyAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	return make(map[string]*v2.AssessmentContentView), nil
}

func (o *EmptyAssessment) MatchProgram() (map[string]*entity.IDName, error) {
	return make(map[string]*entity.IDName), nil
}

func (o *EmptyAssessment) MatchSubject() (map[string][]*entity.IDName, error) {
	return make(map[string][]*entity.IDName), nil
}

func (o *EmptyAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	return make(map[string][]*entity.IDName), nil
}

func GetAssessmentMatch(assessmentType v2.AssessmentType, ags *AssessmentsGrain) IAssessmentMatch {
	var match IAssessmentMatch
	switch assessmentType {
	case v2.AssessmentTypeOnlineClass:
		match = NewOnlineClassAssessment(ags)
	case v2.AssessmentTypeOfflineClass:
		match = NewOfflineClassAssessment(ags)
	case v2.AssessmentTypeOnlineStudy:
		match = NewOnlineStudyAssessment(ags)
	case v2.AssessmentTypeReviewStudy:
		match = NewReviewStudyAssessment(ags)
	default:
		match = NewEmptyAssessment()
	}

	return match
}

func ConvertAssessmentPageReply(assessments []*v2.Assessment, match IAssessmentMatch) ([]*v2.AssessmentQueryReply, error) {
	result := make([]*v2.AssessmentQueryReply, len(assessments))

	scheduleMap, err := match.MatchSchedule()
	if err != nil {
		return nil, err
	}
	lessPlanMap, err := match.MatchLessPlan()
	if err != nil {
		return nil, err
	}
	teacherMap, err := match.MatchTeacher()
	if err != nil {
		return nil, err
	}
	programMap, err := match.MatchProgram()
	if err != nil {
		return nil, err
	}
	subjectMap, err := match.MatchSubject()
	if err != nil {
		return nil, err
	}
	classMap, err := match.MatchClass()
	if err != nil {
		return nil, err
	}
	completeRateMap, err := match.MatchCompleteRate()
	if err != nil {
		return nil, err
	}
	remainingMap, err := match.MatchRemainingTime()
	if err != nil {
		return nil, err
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

		replyItem.Teachers = teacherMap[item.ID]

		schedule, ok := scheduleMap[item.ScheduleID]
		if !ok {
			continue
		}
		if lessPlanItem, ok := lessPlanMap[item.ID]; ok {
			replyItem.LessonPlan = &entity.IDName{
				ID:   lessPlanItem.ID,
				Name: lessPlanItem.Name,
			}
		}

		replyItem.Program = programMap[item.ID]
		replyItem.Subjects = subjectMap[item.ID]
		replyItem.DueAt = schedule.DueAt
		replyItem.ClassInfo = classMap[item.ID]
		replyItem.RemainingTime = remainingMap[item.ID]
		replyItem.CompleteRate = completeRateMap[item.ID]
	}

	return result, nil
}
