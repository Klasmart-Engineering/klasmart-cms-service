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
	dbo.Querier
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentOutcome) error
	UpdateByAssessmentIDAndOutcomeID(ctx context.Context, tx *dbo.DBContext, item entity.AssessmentOutcome) error
	UncheckByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) error
	DeleteByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) error
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

type assessmentOutcomeDA struct {
	dbo.BaseDA
}

func (*assessmentOutcomeDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentOutcome) error {
	var (
		columns = []string{"id", "assessment_id", "outcome_id"}
		matrix  [][]interface{}
	)
	for _, item := range items {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		matrix = append(matrix, []interface{}{item.ID, item.AssessmentID, item.OutcomeID})
	}
	format, values := SQLBatchInsert(entity.AssessmentOutcome{}.TableName(), columns, matrix)
	if err := tx.Exec(format, values...).Error; err != nil {
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

func (*assessmentOutcomeDA) UpdateByAssessmentIDAndOutcomeID(ctx context.Context, tx *dbo.DBContext, item entity.AssessmentOutcome) error {
	changes := map[string]interface{}{
		"skip":          item.Skip,
		"none_achieved": item.NoneAchieved,
		"checked":       item.Checked,
	}
	if err := tx.
		Model(entity.AssessmentOutcome{}).
		Where("assessment_id = ? and outcome_id = ?", item.AssessmentID, item.OutcomeID).
		Updates(changes).
		Error; err != nil {
		log.Error(ctx, "UpdateByAssessmentIDAndOutcomeID: Updates: update failed",
			log.Err(err),
			log.Any("assessment_outcome", item),
		)
		return err
	}
	return nil
}

func (*assessmentOutcomeDA) UncheckByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) error {
	changes := map[string]interface{}{
		"checked": false,
	}
	if err := tx.
		Model(entity.AssessmentOutcome{}).
		Where("assessment_id = ?", assessmentID).
		Updates(changes).
		Error; err != nil {
		return err
	}
	return nil
}

type QueryAssessmentOutcomeConditions struct {
	AssessmentIDs []string `json:"assessment_ids"`
	Checked       *bool    `json:"checked"`
}

func (c *QueryAssessmentOutcomeConditions) GetConditions() ([]string, []interface{}) {
	b := NewSQLTemplate("")

	if c.AssessmentIDs != nil {
		if len(c.AssessmentIDs) == 0 {
			return FalseSQLTemplate().DBOConditions()
		}
		b.Appendf("assessment_id in (?)", c.AssessmentIDs)
	}
	if c.Checked != nil {
		b.Appendf("checked = ?", *c.Checked)
	}

	return b.DBOConditions()
}

func (c *QueryAssessmentOutcomeConditions) GetPager() *dbo.Pager { return nil }

func (c *QueryAssessmentOutcomeConditions) GetOrderBy() string { return "" }
