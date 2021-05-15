package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IContentAssessmentModel interface {
	GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string)
	List(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.ListContentAssessmentsArgs)
	Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.UpdateContentAssessmentArgs)
}

var (
	contentAssessmentModelInstance     IContentAssessmentModel
	contentAssessmentModelInstanceOnce = sync.Once{}
)

func GetContentAssessmentModel() IContentAssessmentModel {
	contentAssessmentModelInstanceOnce.Do(func() {
		contentAssessmentModelInstance = &contentAssessmentModel{}
	})
	return contentAssessmentModelInstance
}

type contentAssessmentModel struct{}

func (m *contentAssessmentModel) GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string) {
}

func (m *contentAssessmentModel) List(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.ListContentAssessmentsArgs) {
	panic("implement me")
}

func (m *contentAssessmentModel) Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.UpdateContentAssessmentArgs) {
	panic("implement me")
}
