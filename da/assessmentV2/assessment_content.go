package assessmentV2

import (
	"context"
	"database/sql"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"

	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
)

type IAssessmentContentDA interface {
	dbo.DataAccesser
	DeleteByAssessmentIDsTx(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) error
}

type assessmentContentDA struct {
	dbo.BaseDA
}

func (a *assessmentContentDA) DeleteByAssessmentIDsTx(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string) error {
	tx.ResetCondition()

	if err := tx.
		Where("assessment_id in (?)", assessmentIDs).
		Delete(&v2.AssessmentContent{}).Error; err != nil {
		log.Error(ctx, "delete assessment content failed",
			log.Strings("assessmentIDs", assessmentIDs),
		)
		return err
	}

	return nil
}

var (
	_assessmentContentOnce sync.Once
	_assessmentContentDA   IAssessmentContentDA
)

func GetAssessmentContentDA() IAssessmentContentDA {
	_assessmentContentOnce.Do(func() {
		_assessmentContentDA = &assessmentContentDA{}
	})
	return _assessmentContentDA
}

type AssessmentContentCondition struct {
	AssessmentID sql.NullString
	ContentType  sql.NullString

	DeleteAt sql.NullString
}

func (c AssessmentContentCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.AssessmentID.Valid {
		wheres = append(wheres, "assessment_id = ?")
		params = append(params, c.AssessmentID.String)
	}

	if c.ContentType.Valid {
		wheres = append(wheres, "content_type = ?")
		params = append(params, c.ContentType.String)
	}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "delete_at>0")
	} else {
		wheres = append(wheres, "(delete_at=0)")
	}

	return wheres, params
}

func (c AssessmentContentCondition) GetOrderBy() string {
	return ""
}

func (c AssessmentContentCondition) GetPager() *dbo.Pager {
	return nil
}
