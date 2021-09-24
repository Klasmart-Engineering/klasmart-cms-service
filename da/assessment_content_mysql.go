package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var (
	assessmentContentDAInstance     IAssessmentContentDA
	assessmentContentDAInstanceOnce = sync.Once{}
)

func GetAssessmentContentDA() IAssessmentContentDA {
	assessmentContentDAInstanceOnce.Do(func() {
		assessmentContentDAInstance = &assessmentContentDA{}
	})
	return assessmentContentDAInstance
}

type IAssessmentContentDA interface {
	dbo.DataAccesser
	GetLessonPlan(ctx context.Context, tx *dbo.DBContext, assessmentID string) (*entity.AssessmentContent, error)
	GetLessonMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string) ([]*entity.AssessmentContent, error)
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContent) error
	UpdatePartial(ctx context.Context, tx *dbo.DBContext, args *UpdatePartialAssessmentContentArgs) error
	UncheckAllLessonMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string) error
	BatchCheckLessonMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string, materialIDs []string) error
}

type assessmentContentDA struct {
	dbo.BaseDA
}

func (d *assessmentContentDA) GetLessonPlan(ctx context.Context, tx *dbo.DBContext, assessmentID string) (*entity.AssessmentContent, error) {
	tx.ResetCondition()

	var plans []*entity.AssessmentContent
	if err := d.QueryTx(ctx, tx, &QueryAssessmentContentConditions{
		AssessmentID: entity.NullString{
			String: assessmentID,
			Valid:  true,
		},
		ContentType: entity.NullContentType{
			Value: entity.ContentTypePlan,
			Valid: true,
		},
	}, &plans); err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, dbo.ErrRecordNotFound
	}
	return plans[0], nil
}

func (d *assessmentContentDA) GetLessonMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string) ([]*entity.AssessmentContent, error) {
	var materials []*entity.AssessmentContent
	if err := d.QueryTx(ctx, tx, &QueryAssessmentContentConditions{
		AssessmentID: entity.NullString{
			String: assessmentID,
			Valid:  true,
		},
		ContentType: entity.NullContentType{
			Value: entity.ContentTypeMaterial,
			Valid: true,
		},
	}, &materials); err != nil {
		return nil, err
	}
	return materials, nil
}

func (d *assessmentContentDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContent) error {
	if len(items) == 0 {
		return nil
	}
	_, err := d.InsertTx(ctx, tx, &items)
	if err != nil {
		log.Error(ctx, "BatchInsert: SQLBatchInsert: batch insert assessment contents failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

func (*assessmentContentDA) UpdatePartial(ctx context.Context, tx *dbo.DBContext, args *UpdatePartialAssessmentContentArgs) error {
	tx.ResetCondition()

	changes := map[string]interface{}{
		"content_comment": args.ContentComment,
		"checked":         args.Checked,
	}
	if err := tx.
		Model(entity.AssessmentContent{}).
		Where("assessment_id = ? and content_id = ?", args.AssessmentID, args.ContentID).
		Updates(changes).
		Error; err != nil {
		log.Error(ctx, "UpdatePartial: updates failed",
			log.Err(err),
			log.Any("args", args),
		)
		return err
	}
	return nil
}

func (*assessmentContentDA) UncheckAllLessonMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string) error {
	tx.ResetCondition()

	changes := map[string]interface{}{
		"checked": false,
	}
	if err := tx.
		Model(entity.AssessmentContent{}).
		Where("assessment_id = ? and content_type = ?", assessmentID, entity.ContentTypeMaterial).
		Updates(changes).
		Error; err != nil {
		log.Error(ctx, "UncheckAllLessonMaterials: uncheck failed",
			log.Err(err),
			log.String("assessment_id", assessmentID),
		)
		return err
	}
	return nil
}

func (*assessmentContentDA) BatchCheckLessonMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string, materialIDs []string) error {
	tx.ResetCondition()

	changes := map[string]interface{}{
		"checked": true,
	}
	if err := tx.
		Model(entity.AssessmentContent{}).
		Where("content_type = ? and assessment_id = ? and content_id in (?)",
			entity.ContentTypeMaterial, assessmentID, materialIDs).
		Updates(changes).
		Error; err != nil {
		log.Error(ctx, "BatchCheckLessonMaterials: check failed",
			log.Err(err),
			log.String("assessment_id", assessmentID),
			log.Strings("material_ids", materialIDs),
		)
		return err
	}
	return nil
}

// region utils

type QueryAssessmentContentConditions struct {
	AssessmentID  entity.NullString
	AssessmentIDs entity.NullStrings
	ContentIDs    entity.NullStrings
	ContentType   entity.NullContentType
	Checked       entity.NullBool
}

func (c *QueryAssessmentContentConditions) GetConditions() ([]string, []interface{}) {
	t := NewSQLTemplate("")
	if c.AssessmentID.Valid {
		t.Appendf("assessment_id = ?", c.AssessmentID.String)
	}
	if c.AssessmentIDs.Valid {
		t.Appendf("assessment_id in (?)", c.AssessmentIDs.Strings)
	}
	if c.ContentIDs.Valid {
		t.Appendf("content_id in (?)", c.ContentIDs.Strings)
	}
	if c.ContentType.Valid {
		t.Appendf("content_type = ?", c.ContentType.Value)
	}
	if c.Checked.Valid {
		t.Appendf("checked = ?", c.Checked.Bool)
	}
	return t.DBOConditions()
}

func (c *QueryAssessmentContentConditions) GetPager() *dbo.Pager {
	return nil
}

func (c *QueryAssessmentContentConditions) GetOrderBy() string {
	return ""
}

type UpdatePartialAssessmentContentArgs struct {
	AssessmentID   string
	ContentID      string
	ContentComment string
	Checked        bool
}

// endregion
