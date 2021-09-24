package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IAssessmentContentOutcomeDA interface {
	dbo.DataAccesser
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContentOutcome) error
	UpdateNoneAchieved(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContentOutcome) error
}

var (
	assessmentContentOutcomeDAInstance     IAssessmentContentOutcomeDA
	assessmentContentOutcomeDAInstanceOnce = sync.Once{}
)

func GetAssessmentContentOutcomeDA() IAssessmentContentOutcomeDA {
	assessmentContentOutcomeDAInstanceOnce.Do(func() {
		assessmentContentOutcomeDAInstance = &assessmentContentOutcomeDA{}
	})
	return assessmentContentOutcomeDAInstance
}

type assessmentContentOutcomeDA struct {
	dbo.BaseDA
}

func (as *assessmentContentOutcomeDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContentOutcome) error {
	if len(items) == 0 {
		return nil
	}
	_, err := as.Insert(ctx, &items)
	if err != nil {
		log.Error(ctx, "BatchInsert: SQLBatchInsert: batch insert assessment content outcomes failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

type QueryAssessmentContentOutcomeConditions struct {
	AssessmentIDs                 entity.NullStrings
	OutocmeIDs                    entity.NullStrings
	ContentIDs                    entity.NullStrings
	AssessmentIDAndOutcomeIDPairs entity.NullAssessmentOutcomeKeys
}

func (c *QueryAssessmentContentOutcomeConditions) GetConditions() ([]string, []interface{}) {
	t := NewSQLTemplate("")
	if c.AssessmentIDs.Valid {
		t.Appendf("assessment_id in (?)", c.AssessmentIDs.Strings)
	}
	if c.ContentIDs.Valid {
		t.Appendf("content_id in (?)", c.ContentIDs.Strings)
	}
	if c.OutocmeIDs.Valid {
		t.Appendf("outcome_id in (?)", c.OutocmeIDs.Strings)
	}
	if c.AssessmentIDAndOutcomeIDPairs.Valid {
		temp := NewSQLTemplate("")
		for _, pair := range c.AssessmentIDAndOutcomeIDPairs.Value {
			temp.Appendf("(assessment_id = ? and outcome_id = ?)", pair.AssessmentID, pair.OutcomeID)
		}
		t.AppendResult(temp.Or())
	}
	return t.DBOConditions()
}

func (c *QueryAssessmentContentOutcomeConditions) GetPager() *dbo.Pager {
	return nil
}

func (c *QueryAssessmentContentOutcomeConditions) GetOrderBy() string {
	return ""
}

func (*assessmentContentOutcomeDA) UpdateNoneAchieved(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContentOutcome) error {
	tx.Reset()

	falseChanges := map[string]interface{}{"none_achieved": false}
	trueChanges := map[string]interface{}{"none_achieved": true}
	falseCond := NewSQLTemplate("")
	trueCond := NewSQLTemplate("")
	for _, item := range items {
		if item.NoneAchieved {
			trueCond.Appendf("(assessment_id = ? and content_id = ? and outcome_id = ?)", item.AssessmentID, item.ContentID, item.OutcomeID)
		} else {
			falseCond.Appendf("(assessment_id = ? and content_id = ? and outcome_id = ?)", item.AssessmentID, item.ContentID, item.OutcomeID)
		}
	}

	// update true
	trueFormat, trueValues := trueCond.Or()
	if trueFormat != "" {
		if err := tx.
			Model(entity.AssessmentContentOutcome{}).
			Where(trueFormat, trueValues...).
			Updates(trueChanges).
			Error; err != nil {
			log.Error(ctx, "update none achieved: update true failed",
				log.Err(err),
				log.String("format", trueFormat),
				log.Any("values", trueValues),
				log.Any("items", items),
			)
			return err
		}
	}

	// update false
	falseFormat, falseValues := falseCond.Or()
	if falseFormat != "" {
		if err := tx.
			Model(entity.AssessmentContentOutcome{}).
			Where(falseFormat, falseValues...).
			Updates(falseChanges).
			Error; err != nil {
			log.Error(ctx, "update none achieved: update false failed",
				log.Err(err),
				log.String("format", falseFormat),
				log.Any("values", falseValues),
				log.Any("items", items),
			)
			return err
		}
	}

	return nil
}
