package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IAssessmentModel interface {
	Detail(ctx context.Context, tx dbo.DBContext, id string) (*entity.AssessmentDetailView, error)
	List(ctx context.Context, tx dbo.DBContext, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error)
	Add(ctx context.Context, tx dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error)
	Update(ctx context.Context, tx dbo.DBContext, cmd entity.UpdateAssessmentCommand) error
}

var (
	assessmentModelInstance     IAssessmentModel
	assessmentModelInstanceOnce = sync.Once{}
)

func GetAssessmentModel() IAssessmentModel {
	assessmentModelInstanceOnce.Do(func() {
		assessmentModelInstance = &assessmentModel{}
	})
	return assessmentModelInstance
}

type assessmentModel struct{}

func (a *assessmentModel) Detail(ctx context.Context, tx dbo.DBContext, id string) (*entity.AssessmentDetailView, error) {
	//item, err := da.GetAssessmentDA().Get(ctx, tx, id)
	//if err != nil {
	//	log.Error(ctx, "get assessment detail: get from da failed",
	//		log.Err(err),
	//		log.String("id", "id"),
	//	)
	//	return nil, err
	//}
	// TODO: fill external fields
	panic("implement me")
}

func (a *assessmentModel) List(ctx context.Context, tx dbo.DBContext, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error) {
	panic("implement me")
}

func (a *assessmentModel) Add(ctx context.Context, tx dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error) {
	panic("implement me")
}

func (a *assessmentModel) Update(ctx context.Context, tx dbo.DBContext, cmd entity.UpdateAssessmentCommand) error {
	panic("implement me")
}
