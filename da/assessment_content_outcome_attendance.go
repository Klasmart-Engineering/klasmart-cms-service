package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IAssessmentContentOutcomeAttendanceDA interface {
	dbo.DataAccesser
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContentOutcomeAttendance) error
	BatchDelete(ctx context.Context, tx *dbo.DBContext, keys []*DeleteAssessmentContentOutcomeAttendanceKey) error
}

var (
	assessmentContentOutcomeAttendanceDAInstance     IAssessmentContentOutcomeAttendanceDA
	assessmentContentOutcomeAttendanceDAInstanceOnce = sync.Once{}
)

func GetAssessmentContentOutcomeAttendanceDA() IAssessmentContentOutcomeAttendanceDA {
	assessmentContentOutcomeAttendanceDAInstanceOnce.Do(func() {
		assessmentContentOutcomeAttendanceDAInstance = &assessmentContentOutcomeAttendanceDA{}
	})
	return assessmentContentOutcomeAttendanceDAInstance
}

type assessmentContentOutcomeAttendanceDA struct {
	dbo.BaseDA
}

type QueryAssessmentContentOutcomeAttendanceCondition struct {
	AssessmentIDs                 entity.NullStrings
	ContentIDs                    entity.NullStrings
	OutcomeIDs                    entity.NullStrings
	AttendanceIDs                 entity.NullStrings
	ContentIDAndOutcomeIDPairs    NullContentIDAndOutcomeIDKeys
	AssessmentIDAndOutcomeIDPairs entity.NullAssessmentOutcomeKeys
}

type ContentIDAndOutcomeIDKey struct {
	ContentID string
	OutcomeID string
}

type NullContentIDAndOutcomeIDKeys struct {
	Value []*ContentIDAndOutcomeIDKey
	Valid bool
}

func (c QueryAssessmentContentOutcomeAttendanceCondition) GetConditions() ([]string, []interface{}) {
	t := NewSQLTemplate("")
	if c.AssessmentIDs.Valid {
		t.Appendf("assessment_id in (?)", c.AssessmentIDs.Strings)
	}
	if c.ContentIDs.Valid {
		t.Appendf("content_id in (?)", c.OutcomeIDs.Strings)
	}
	if c.OutcomeIDs.Valid {
		t.Appendf("outcome_id in (?)", c.OutcomeIDs.Strings)
	}
	if c.AttendanceIDs.Valid {
		t.Appendf("attendance_id in (?)", c.AttendanceIDs.Strings)
	}
	if c.ContentIDAndOutcomeIDPairs.Valid {
		temp := NewSQLTemplate("")
		for _, key := range c.ContentIDAndOutcomeIDPairs.Value {
			temp.Appendf("(content_id = ? and outcome_id = ?)", key.ContentID, key.OutcomeID)
		}
		t.AppendResult(temp.Or())
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

func (c QueryAssessmentContentOutcomeAttendanceCondition) GetPager() *dbo.Pager {
	return nil
}

func (c QueryAssessmentContentOutcomeAttendanceCondition) GetOrderBy() string {
	return ""
}

func (*assessmentContentOutcomeAttendanceDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContentOutcomeAttendance) error {
	if len(items) == 0 {
		return nil
	}
	var (
		columns = []string{"id", "assessment_id", "content_id", "outcome_id", "attendance_id"}
		matrix  [][]interface{}
	)
	for _, item := range items {
		matrix = append(matrix, []interface{}{
			item.ID,
			item.AssessmentID,
			item.ContentID,
			item.OutcomeID,
			item.AttendanceID,
		})
	}
	format, values := SQLBatchInsert(entity.AssessmentContentOutcomeAttendance{}.TableName(), columns, matrix)
	if err := tx.Exec(format, values...).Error; err != nil {
		log.Error(ctx, "BatchInsert: SQLBatchInsert: batch insert assessment content outcome attendances failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

type DeleteAssessmentContentOutcomeAttendanceKey struct {
	AssessmentID string
	ContentID    string
	OutcomeID    string
}

func (*assessmentContentOutcomeAttendanceDA) BatchDelete(ctx context.Context, tx *dbo.DBContext, keys []*DeleteAssessmentContentOutcomeAttendanceKey) error {
	if len(keys) == 0 {
		return nil
	}
	t := &SQLTemplate{}
	for _, key := range keys {
		t.Appendf("(assessment_id = ? and content_id = ? and outcome_id = ?)", key.AssessmentID, key.ContentID, key.OutcomeID)
	}
	format, values := t.Or()
	if err := tx.
		Where(format, values...).
		Delete(entity.AssessmentContentOutcomeAttendance{}).Error; err != nil {
		log.Error(ctx, "batch delete assessment content outcome attendances failed",
			log.Err(err),
			log.Any("keys", keys),
		)
		return err
	}
	return nil
}
