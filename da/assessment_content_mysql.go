package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IAssessmentContentDA interface {
	dbo.DataAccesser
	GetPlan(ctx context.Context, tx *dbo.DBContext, assessmentID string) (*entity.AssessmentContent, error)
	GetMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string) ([]*entity.AssessmentContent, error)
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContent) error
	UncheckAllMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string) error
	BatchCheckMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string, materialIDs []string) error
	UpdatePartial(ctx context.Context, tx *dbo.DBContext, args UpdatePartialAssessmentContentArgs) error
}

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

type assessmentContentDA struct {
	dbo.BaseDA
}

func (d *assessmentContentDA) GetPlan(ctx context.Context, tx *dbo.DBContext, assessmentID string) (*entity.AssessmentContent, error) {
	var (
		plans       []*entity.AssessmentContent
		contentType = entity.ContentType(entity.ContentTypePlan)
	)
	if err := d.QueryTx(ctx, tx, &QueryAssessmentContentConditions{
		AssessmentID: &assessmentID,
		ContentType:  &contentType,
	}, &plans); err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, dbo.ErrRecordNotFound
	}
	return plans[0], nil
}

func (d *assessmentContentDA) GetMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string) ([]*entity.AssessmentContent, error) {
	var (
		materials   []*entity.AssessmentContent
		contentType = entity.ContentType(entity.ContentTypeMaterial)
	)
	if err := d.QueryTx(ctx, tx, &QueryAssessmentContentConditions{
		AssessmentID: &assessmentID,
		ContentType:  &contentType,
	}, &materials); err != nil {
		return nil, err
	}
	return materials, nil
}

func (*assessmentContentDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentContent) error {
	var (
		columns = []string{"id", "assessment_id", "content_id", "content_name", "content_type", "checked"}
		matrix  [][]interface{}
	)
	for _, item := range items {
		matrix = append(matrix, []interface{}{
			item.ID,
			item.AssessmentID,
			item.ContentID,
			item.ContentName,
			item.ContentType,
			item.Checked,
		})
	}
	format, values := SQLBatchInsert(entity.AssessmentContent{}.TableName(), columns, matrix)
	if err := tx.Exec(format, values...).Error; err != nil {
		log.Error(ctx, "BatchInsert: SQLBatchInsert: batch insert assessment contents failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

func (*assessmentContentDA) UncheckAllMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string) error {
	changes := map[string]interface{}{
		"checked": false,
	}
	if err := tx.
		Model(entity.AssessmentContent{}).
		Where("assessment_id = ? and content_type = ?", assessmentID, entity.ContentTypeMaterial).
		Updates(changes).
		Error; err != nil {
		log.Error(ctx, "UncheckAllMaterials: uncheck failed",
			log.Err(err),
			log.String("assessment_id", assessmentID),
		)
		return err
	}
	return nil
}

func (*assessmentContentDA) BatchCheckMaterials(ctx context.Context, tx *dbo.DBContext, assessmentID string, materialIDs []string) error {
	changes := map[string]interface{}{
		"checked": true,
	}
	if err := tx.
		Model(entity.AssessmentContent{}).
		Where("content_type = ? and assessment_id = ? and content_id in (?)",
			entity.ContentTypeMaterial, assessmentID, materialIDs).
		Updates(changes).
		Error; err != nil {
		log.Error(ctx, "BatchCheckMaterials: check failed",
			log.Err(err),
			log.String("assessment_id", assessmentID),
			log.Strings("material_ids", materialIDs),
		)
		return err
	}
	return nil
}

func (*assessmentContentDA) UpdatePartial(ctx context.Context, tx *dbo.DBContext, args UpdatePartialAssessmentContentArgs) error {
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

type QueryAssessmentContentConditions struct {
	AssessmentID *string
	ContentIDs   []string
	ContentType  *entity.ContentType
	Checked      *bool
}

func (c *QueryAssessmentContentConditions) GetConditions() ([]string, []interface{}) {
	t := NewSQLTemplate("")
	if c.AssessmentID != nil {
		t.Appendf("assessment_id = ?", *c.AssessmentID)
	}
	if c.ContentIDs != nil {
		if len(c.ContentIDs) == 0 {
			return FalseSQLTemplate().DBOConditions()
		}
		t.Appendf("content_id in (?)", c.ContentIDs)
	}
	if c.ContentType != nil {
		t.Appendf("content_type = ?", *c.ContentType)
	}
	if c.Checked != nil {
		t.Appendf("checked = ?", *c.Checked)
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
