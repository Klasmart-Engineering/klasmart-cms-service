package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IAssessmentContentOutcomeDA interface {
	dbo.DataAccesser
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContentOutcome) error
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

func (*assessmentContentOutcomeDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContentOutcome) error {
	var (
		columns = []string{"id", "assessment_id", "content_id", "outcome_id"}
		matrix  [][]interface{}
	)
	for _, item := range items {
		matrix = append(matrix, []interface{}{
			item.ID,
			item.AssessmentID,
			item.ContentID,
			item.OutcomeID,
		})
	}
	format, values := SQLBatchInsert(entity.AssessmentContentOutcome{}.TableName(), columns, matrix)
	if err := tx.Exec(format, values...).Error; err != nil {
		log.Error(ctx, "BatchInsert: SQLBatchInsert: batch insert assessment content outcomes failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

type QueryAssessmentContentOutcomeConditions struct {
	AssessmentIDs []string
	ContentIDs    []string
}

func (c *QueryAssessmentContentOutcomeConditions) GetConditions() ([]string, []interface{}) {
	t := NewSQLTemplate("")
	if c.AssessmentIDs != nil {
		if len(c.AssessmentIDs) == 0 {
			return FalseSQLTemplate().DBOConditions()
		}
		t.Appendf("assessment_id in (?)", c.AssessmentIDs)
	}
	if c.ContentIDs != nil {
		if len(c.ContentIDs) == 0 {
			return FalseSQLTemplate().DBOConditions()
		}
		t.Appendf("content_id in (?)", c.ContentIDs)
	}
	return t.DBOConditions()
}

func (c *QueryAssessmentContentOutcomeConditions) GetPager() *dbo.Pager {
	return nil
}

func (c *QueryAssessmentContentOutcomeConditions) GetOrderBy() string {
	return ""
}
