package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IAssessmentModel interface {
	List(ctx context.Context, tx dbo.DBContext, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error)
	Detail(ctx context.Context, tx dbo.DBContext, id string) (*entity.AssessmentDetailView, error)
	Add(ctx context.Context, tx dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error)
	Update(ctx context.Context, tx dbo.DBContext, cmd entity.UpdateAssessmentCommand) error
}
