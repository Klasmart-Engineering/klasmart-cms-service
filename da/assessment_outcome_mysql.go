package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
	"sync"
)

type IAssessmentOutcomeDA interface {
	GetListByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string, checked *bool) ([]*entity.AssessmentOutcome, error)
	BatchGetByAssessmentIDs(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string, checked *bool) ([]*entity.AssessmentOutcome, error)
	BatchGetMapByKeys(ctx context.Context, tx *dbo.DBContext, keys []entity.AssessmentOutcomeKey, checked *bool) (map[entity.AssessmentOutcomeKey]*entity.AssessmentOutcome, error)
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

type assessmentOutcomeDA struct{}

func (*assessmentOutcomeDA) GetListByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string, checked *bool) ([]*entity.AssessmentOutcome, error) {
	db := tx.Where("assessment_id = ?", assessmentID)
	if checked != nil {
		db = db.Where("checked = ?", *checked)
	}
	var items []*entity.AssessmentOutcome
	if err := db.Find(&items).Error; err != nil {
		log.Error(ctx, "GetListByAssessmentID: db.Find: find failed",
			log.Err(err),
			log.Any("checked", checked),
			log.String("assessment_id", assessmentID),
		)
		return nil, err
	}
	return items, nil
}

func (*assessmentOutcomeDA) BatchGetByAssessmentIDs(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string, checked *bool) ([]*entity.AssessmentOutcome, error) {
	db := tx.Where("assessment_id in (?)", assessmentIDs)
	if checked != nil {
		db = db.Where("checked = ?", *checked)
	}
	var items []*entity.AssessmentOutcome
	if err := db.Find(&items).Error; err != nil {
		log.Error(ctx, "batch get assessment outcome by assessment ids: find failed",
			log.Err(err),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return nil, err
	}
	return items, nil
}

func (*assessmentOutcomeDA) BatchGetMapByKeys(ctx context.Context, tx *dbo.DBContext, keys []entity.AssessmentOutcomeKey, checked *bool) (map[entity.AssessmentOutcomeKey]*entity.AssessmentOutcome, error) {
	var (
		template string
		formats  []string
		values   []interface{}
	)
	for _, key := range keys {
		formats = append(formats, "(assessment_id = ? and outcome_id = ?)")
		values = append(values, key.AssessmentID, key.OutcomeID)
	}
	template = strings.Join(formats, " or ")
	db := tx.Where(template, values...)
	if checked != nil {
		db = db.Where("checked = ?", *checked)
	}
	var items []*entity.AssessmentOutcome
	if err := db.Find(&items).Error; err != nil {
		log.Error(ctx, "batch get assessment outcome by keys: find failed",
			log.Err(err),
			log.Any("checked", checked),
			log.String("template", template),
			log.Any("values", values),
			log.Any("keys", keys),
		)
		return nil, err
	}
	result := map[entity.AssessmentOutcomeKey]*entity.AssessmentOutcome{}
	for _, item := range items {
		result[entity.AssessmentOutcomeKey{
			AssessmentID: item.AssessmentID,
			OutcomeID:    item.OutcomeID,
		}] = item
	}
	return result, nil
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
		Where("assessment_id = ? and outcome_id = ?", assessmentID).
		Updates(changes).
		Error; err != nil {
		return err
	}
	return nil
}
