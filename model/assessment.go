package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IAssessmentModel interface {
	List(ctx context.Context, tx dbo.DBContext, cmd entity.ListAssessmentCommand) ([]*entity.AssessmentListView, error)
	Detail(ctx context.Context, tx dbo.DBContext, id string) (*entity.AssessmentDetailView, error)
	Add(ctx context.Context, tx dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error)
	PatchUpdate(ctx context.Context, tx dbo.DBContext, cmd entity.PatchUpdateAssessmentCommand) error
}
