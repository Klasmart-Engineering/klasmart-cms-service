package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IAssessmentOutcomeDA interface {
	GetOutcomeIDsByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) ([]string, error)
	BatchGetByAssessmentIDAndOutcomeIDs(ctx context.Context, tx *dbo.DBContext, assessmentID string, outcomeIDs []string) ([]*entity.AssessmentOutcome, error)
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentOutcome) error
	DeleteByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) error
	UpdateSkipField(ctx context.Context, tx *dbo.DBContext, assessmentID string, outcomeID string, skip bool) error
}

var (
	assessmentOutcomeDAInstance     IAssessmentOutcomeDA
	assessmentOutcomeDAInstanceOnce = sync.Once{}
)

func GetAssessmentOutcomeDA() IAssessmentOutcomeDA {
	assessmentOutcomeDAInstanceOnce.Do(func() {
		assessmentOutcomeDAInstance = &assessmentOutcomeDA{}
	})
	return assessmentOutcomeDAInstance
}

type assessmentOutcomeDA struct{}

func (*assessmentOutcomeDA) GetOutcomeIDsByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) ([]string, error) {
	var items []entity.AssessmentOutcome
	if err := tx.Where("assessment_id = ?", assessmentID).Find(&items).Error; err != nil {
		return nil, err
	}
	var ids []string
	for _, item := range items {
		ids = append(ids, item.OutcomeID)
	}
	return ids, nil
}

func (*assessmentOutcomeDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentOutcome) error {
	columns := []string{"id", "assessment_id", "outcome_id"}
	var values [][]interface{}
	for _, item := range items {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		values = append(values, []interface{}{item.ID, item.AssessmentID, item.OutcomeID})
	}
	template := SQLBatchInsert(entity.AssessmentOutcome{}.TableName(), columns, values)
	if err := tx.Exec(template.Format, template.Values...).Error; err != nil {
		log.Error(ctx, "batch insert assessments_outcomes: batch insert failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

func (*assessmentOutcomeDA) DeleteByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) error {
	if err := tx.Where("assessment_id", assessmentID).Delete(entity.AssessmentOutcome{}).Error; err != nil {
		log.Error(ctx, "delete outcomes by id: delete failed from db",
			log.Err(err),
			log.String("assessment_id", assessmentID),
		)
		return err
	}
	return nil
}

func (d *assessmentOutcomeDA) UpdateSkipField(ctx context.Context, tx *dbo.DBContext, assessmentID string, outcomeID string, skip bool) error {
	if err := tx.
		Model(entity.AssessmentOutcome{}).
		Where("assessment_id = ? and outcome_id = ?", assessmentID, outcomeID).
		Update("skip", skip).
		Error; err != nil {
		return err
	}
	return nil
}

func (d *assessmentOutcomeDA) BatchGetByAssessmentIDAndOutcomeIDs(ctx context.Context, tx *dbo.DBContext, assessmentID string, outcomeIDs []string) ([]*entity.AssessmentOutcome, error) {
	var items []*entity.AssessmentOutcome
	if err := tx.
		Where("assessment_id = ?", assessmentID).
		Where("outcome_id in (?)", outcomeIDs).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
