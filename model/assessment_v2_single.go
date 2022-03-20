package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"strings"
)

type AssessmentGrainInit int

const (
	_ AssessmentGrainInit = iota
)

type AssessmentGrain struct {
	ctx        context.Context
	op         *entity.Operator
	InitRecord map[AssessmentsGrainInit]bool

	assessment *v2.Assessment
	ags        *AssessmentsGrain
}

func NewAssessmentSingleGrain(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) *AssessmentGrain {
	return &AssessmentGrain{
		ctx:        ctx,
		op:         op,
		InitRecord: make(map[AssessmentsGrainInit]bool),
		assessment: assessment,
		ags:        NewAssessmentsGrain(ctx, op, []*v2.Assessment{assessment}),
	}
}

func (ag *AssessmentGrain) getKey(value []string) string {
	return strings.Join(value, "_")
}
