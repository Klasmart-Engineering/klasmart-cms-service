package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

func NewOfflineClassAssessment(ags *AssessmentsGrain) IAssessmentMatch {
	return &OfflineClassAssessment{
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

type OfflineClassAssessment struct {
	EmptyAssessment

	ctx context.Context
	op  *entity.Operator
	ags *AssessmentsGrain
}

func (o *OfflineClassAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	onlineClass := NewOnlineClassAssessment(o.ags)

	return onlineClass.MatchSchedule()
}

func (o *OfflineClassAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	onlineClass := NewOnlineClassAssessment(o.ags)

	return onlineClass.MatchLessPlan()
}

func (o *OfflineClassAssessment) MatchProgram() (map[string]*entity.IDName, error) {
	onlineClass := NewOnlineClassAssessment(o.ags)

	return onlineClass.MatchProgram()
}

func (o *OfflineClassAssessment) MatchSubject() (map[string][]*entity.IDName, error) {
	onlineClass := NewOnlineClassAssessment(o.ags)

	return onlineClass.MatchSubject()
}

func (o *OfflineClassAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	onlineClass := NewOnlineClassAssessment(o.ags)

	return onlineClass.MatchTeacher()
}
