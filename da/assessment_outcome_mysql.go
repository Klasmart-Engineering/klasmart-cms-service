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
	UpdateByAssessmentIDAndOutcomeID(ctx context.Context, tx *dbo.DBContext, item *entity.AssessmentOutcome) error
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

func (as *assessmentOutcomeDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentOutcome) error {
	if len(items) == 0 {
		log.Debug(ctx, "batch insert assessment outcome: no items")
		return nil
	}
	var models []entity.AssessmentOutcome
	for _, item := range items {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		models = append(models, entity.AssessmentOutcome{ID: item.ID, AssessmentID: item.AssessmentID,
			OutcomeID: item.OutcomeID, Skip: item.Skip, NoneAchieved: item.NoneAchieved, Checked: item.Checked})
	}
	_, err := as.InsertTx(ctx, tx, &models)
	if err != nil {
		log.Error(ctx, "batch insert assessments_outcomes: batch insert failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

func (*assessmentOutcomeDA) DeleteByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) error {
	tx.Reset()

	if err := tx.Where("assessment_id", assessmentID).Delete(entity.AssessmentOutcome{}).Error; err != nil {
		log.Error(ctx, "delete outcomes by id: delete failed from db",
			log.Err(err),
			log.String("assessment_id", assessmentID),
		)
		return err
	}
	return nil
}

func (*assessmentOutcomeDA) UpdateByAssessmentIDAndOutcomeID(ctx context.Context, tx *dbo.DBContext, item *entity.AssessmentOutcome) error {
	tx.Reset()

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
	tx.Reset()

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
	AssessmentIDs entity.NullStrings               `json:"assessment_ids"`
	Keys          entity.NullAssessmentOutcomeKeys `json:"keys"`
	Checked       entity.NullBool                  `json:"checked"`
}

func (c *QueryAssessmentOutcomeConditions) GetConditions() ([]string, []interface{}) {
	b := NewSQLTemplate("")

	if c.AssessmentIDs.Valid {
		b.Appendf("assessment_id in (?)", c.AssessmentIDs.Strings)
	}
	if c.Checked.Valid {
		b.Appendf("checked = ?", c.Checked.Bool)
	}
	if c.Keys.Valid {
		temp := NewSQLTemplate("")
		for _, key := range c.Keys.Value {
			temp.Appendf("(assessment_id = ? and outcome_id = ?)", key.AssessmentID, key.OutcomeID)
		}
		b.AppendResult(temp.Or())
	}

	return b.DBOConditions()
}

func (c *QueryAssessmentOutcomeConditions) GetPager() *dbo.Pager { return nil }

func (c *QueryAssessmentOutcomeConditions) GetOrderBy() string { return "" }
