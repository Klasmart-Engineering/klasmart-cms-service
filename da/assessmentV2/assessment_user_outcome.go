package assessmentV2

import (
	"context"
	"database/sql"
	"sync"

	"github.com/KL-Engineering/common-log/log"

	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"

	"github.com/KL-Engineering/dbo"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IAssessmentUserOutcomeDA interface {
	dbo.DataAccesser
	DeleteByAssessmentUserIDsTx(ctx context.Context, tx *dbo.DBContext, assessmentUserIDs []string) error
	GetContentOutcomeByAssessmentID(ctx context.Context, assessmentID string) ([]*AssessmentContentOutcomeDBView, error)
}

type assessmentUserOutcomeDA struct {
	dbo.BaseDA
}

type AssessmentContentOutcomeDBView struct {
	ContentID string `gorm:"content_id"`
	OutcomeID string `gorm:"outcome_id"`
}

func (a *assessmentUserOutcomeDA) GetContentOutcomeByAssessmentID(ctx context.Context, assessmentID string) ([]*AssessmentContentOutcomeDBView, error) {
	tx := dbo.MustGetDB(ctx)
	tx.ResetCondition()

	sql := `
SELECT  assessments_contents_v2.content_id,assessments_users_outcomes_v2.outcome_id FROM assessments_contents_v2
left join assessments_users_outcomes_v2 on assessments_contents_v2.id = assessments_users_outcomes_v2.assessment_content_id
where assessments_contents_v2.assessment_id = ?
group by assessments_contents_v2.content_id,assessments_users_outcomes_v2.outcome_id
`
	var result = make([]*AssessmentContentOutcomeDBView, 0)
	err := tx.Raw(sql, assessmentID).Scan(&result).Error
	if err != nil {
		log.Error(ctx, "get content outcome from assessment error", log.Err(err), log.String("assessmentID", assessmentID))
		return nil, err
	}

	return result, nil
}

func (a *assessmentUserOutcomeDA) DeleteByAssessmentUserIDsTx(ctx context.Context, tx *dbo.DBContext, assessmentUserIDs []string) error {
	tx.ResetCondition()

	if err := tx.Unscoped().
		Where("assessment_user_id in (?)", assessmentUserIDs).
		Delete(&v2.AssessmentUserOutcome{}).Error; err != nil {
		log.Error(ctx, "delete assessment attendance outcome failed",
			log.Strings("assessmentUserIDs", assessmentUserIDs),
		)
		return err
	}

	return nil
}

var (
	_assessmentUserOutcomeOnce sync.Once
	_assessmentUserOutcomeDA   IAssessmentUserOutcomeDA
)

func GetAssessmentUserOutcomeDA() IAssessmentUserOutcomeDA {
	_assessmentUserOutcomeOnce.Do(func() {
		_assessmentUserOutcomeDA = &assessmentUserOutcomeDA{}
	})
	return _assessmentUserOutcomeDA
}

type AssessmentUserOutcomeCondition struct {
	AssessmentUserIDs entity.NullStrings
	IDs               entity.NullStrings
	Status            sql.NullString
}

func (c AssessmentUserOutcomeCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.AssessmentUserIDs.Valid {
		wheres = append(wheres, "assessment_user_id in (?)")
		params = append(params, c.AssessmentUserIDs.Strings)
	}

	if c.IDs.Valid {
		wheres = append(wheres, "id in (?)")
		params = append(params, c.IDs.Strings)
	}
	if c.Status.Valid {
		wheres = append(wheres, "status = ?")
		params = append(params, c.Status.String)
	}

	return wheres, params
}

func (c AssessmentUserOutcomeCondition) GetOrderBy() string {
	return ""
}

func (c AssessmentUserOutcomeCondition) GetPager() *dbo.Pager {
	return nil
}
