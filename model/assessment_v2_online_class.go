package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

func NewOnlineClassAssessment(ags *AssessmentsGrain) IAssessmentMatch {
	return &OnlineClassAssessment{
		ctx: ags.ctx,
		op:  ags.op,
		ags: ags,
	}
}

//func NewOnlineClassAssessmentDetail(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) IAssessmentMatch {
//	return &OnlineClassAssessment{
//		ctx: ctx,
//		op:  op,
//		ags: NewAssessmentsGrain(ctx, op, []*v2.Assessment{assessment}),
//	}
//}

type OnlineClassAssessment struct {
	EmptyAssessment

	ctx context.Context
	op  *entity.Operator
	ags *AssessmentsGrain
}

func (o *OnlineClassAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	scheduleMap, err := o.ags.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	return scheduleMap, nil
}

func (o *OnlineClassAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	scheduleMap, err := o.ags.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	lessPlanMap, err := o.ags.GetLessPlanMap()
	if err != nil {
		return nil, err
	}

	// key:assessmentID
	result := make(map[string]*v2.AssessmentContentView, len(o.ags.assessments))
	for _, item := range o.ags.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			if lessPlanItem, ok := lessPlanMap[schedule.LessonPlanID]; ok && lessPlanItem != nil {
				result[item.ID] = lessPlanItem
			}
		}
	}

	return result, nil
}

func (o *OnlineClassAssessment) MatchProgram() (map[string]*entity.IDName, error) {
	scheduleMap, err := o.ags.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	programMap, err := o.ags.GetProgramMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*entity.IDName, len(o.ags.assessments))
	for _, item := range o.ags.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			result[item.ID] = programMap[schedule.ProgramID]
		}
	}

	return result, nil
}

func (o *OnlineClassAssessment) MatchSubject() (map[string][]*entity.IDName, error) {
	relationMap, err := o.ags.GetScheduleRelationMap()
	if err != nil {
		return nil, err
	}

	subjectMap, err := o.ags.GetSubjectMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*entity.IDName, len(o.ags.assessments))
	for _, item := range o.ags.assessments {
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
				if assUserItem.StatusBySystem == v2.AssessmentUserStatusNotParticipate {
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
