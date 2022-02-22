package assessmentV2

import (
	"context"
	"database/sql"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IAssessmentUserOutcomeDA interface {
	dbo.DataAccesser
	DeleteByAssessmentUserIDsTx(ctx context.Context, tx *dbo.DBContext, assessmentUserIDs []string) error
}

type assessmentUserOutcomeDA struct {
	dbo.BaseDA
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
