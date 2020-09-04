package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IAssessmentModel interface {
	Detail(ctx context.Context, tx *dbo.DBContext, id string) (*entity.AssessmentDetailView, error)
	List(ctx context.Context, tx *dbo.DBContext, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error)
	Add(ctx context.Context, tx *dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error)
	Update(ctx context.Context, tx *dbo.DBContext, cmd entity.UpdateAssessmentCommand) error
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

func (a *assessmentModel) Detail(ctx context.Context, tx *dbo.DBContext, id string) (*entity.AssessmentDetailView, error) {
	var result entity.AssessmentDetailView
	dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		item, err := da.GetAssessmentDA().Get(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "get assessment detail: get from da failed",
				log.Err(err),
				log.String("id", "id"),
			)
			return err
		}
		result = entity.AssessmentDetailView{
			ID:    item.ID,
			Title: item.Title,
			//Attendances:           nil,
			//Subject:               entity.AssessmentSubject{},
			//Teacher:               entity.AssessmentTeacher{},
			ClassEndTime: item.ClassEndTime,
			ClassLength:  item.ClassLength,
			//NumberOfActivities:    0,
			//NumberOfOutcomes:      0,
			CompleteTime:          item.CompleteTime,
			OutcomeAttendanceMaps: nil,
		}
		assessmentIDs, err := da.GetAssessmentAttendanceDA().GetAttendanceIDsByAssessmentID(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "get attendance ids by assessment id: get from da failed",
				log.Err(err),
				log.String("id", "id"),
			)
			return err
		}
		_ = assessmentIDs
		return nil
	})
	// TODO: fill external fields
	panic("implement me")
}

func (a *assessmentModel) List(ctx context.Context, tx *dbo.DBContext, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error) {
	panic("implement me")
}

func (a *assessmentModel) Add(ctx context.Context, tx *dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error) {
	panic("implement me")
}

func (a *assessmentModel) Update(ctx context.Context, tx *dbo.DBContext, cmd entity.UpdateAssessmentCommand) error {
	panic("implement me")
}
