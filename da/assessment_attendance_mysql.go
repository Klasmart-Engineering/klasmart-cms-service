package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IAssessmentAttendanceDA interface {
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentAttendance) error
	Query(ctx context.Context, condition dbo.Conditions, values interface{}) error
	QueryTx(ctx context.Context, db *dbo.DBContext, condition dbo.Conditions, values interface{}) error
	Uncheck(ctx context.Context, db *dbo.DBContext, assessmentID string) error
	Check(ctx context.Context, db *dbo.DBContext, assessmentID string, attendanceIDs []string) error
}

var (
	assessmentAttendanceDAInstance     IAssessmentAttendanceDA
	assessmentAttendanceDAInstanceOnce = sync.Once{}
)

func GetAssessmentAttendanceDA() IAssessmentAttendanceDA {
	assessmentAttendanceDAInstanceOnce.Do(func() {
		assessmentAttendanceDAInstance = &assessmentAttendanceDA{}
	})
	return assessmentAttendanceDAInstance
}

type assessmentAttendanceDA struct {
	dbo.BaseDA
}

func (*assessmentAttendanceDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.AssessmentAttendance) error {
	if len(items) == 0 {
		return nil
	}
	columns := []string{"id", "assessment_id", "attendance_id"}
	var values [][]interface{}
	for _, item := range items {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		values = append(values, []interface{}{item.ID, item.AssessmentID, item.AttendanceID})
	}
	template := SQLBatchInsert(entity.AssessmentAttendance{}.TableName(), columns, values)
	if err := tx.Exec(template.Format, template.Values...).Error; err != nil {
		log.Error(ctx, "batch insert assessments_attendances: batch insert failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

func (*assessmentAttendanceDA) DeleteByAssessmentID(ctx context.Context, tx *dbo.DBContext, assessmentID string) error {
	if err := tx.Where("assessment_id = ?", assessmentID).Delete(entity.AssessmentAttendance{}).Error; err != nil {
		log.Error(ctx, "delete attendances by id: delete failed from db",
			log.Err(err),
			log.String("assessment_id", assessmentID),
		)
		return err
	}
	return nil
}

// Uncheck all assessment attendances
func (a *assessmentAttendanceDA) Uncheck(ctx context.Context, tx *dbo.DBContext, assessmentID string) error {
	if err := tx.Model(&entity.AssessmentAttendance{}).Where("assessment_id = ?", assessmentID).
		Update("checked", false).
		Error; err != nil {
		log.Error(ctx, "uncheck assessment attendance: update failed",
			log.Err(err),
			log.String("assessment_id", assessmentID),
		)
		return err
	}
	return nil
}

func (a *assessmentAttendanceDA) Check(ctx context.Context, tx *dbo.DBContext, assessmentID string, attendanceIDs []string) error {
	if len(attendanceIDs) == 0 {
		return nil
	}
	if err := tx.Model(&entity.AssessmentAttendance{}).Where("assessment_id = ? and attendance_id in (?)", assessmentID, attendanceIDs).
		Update("checked", true).
		Error; err != nil {
		log.Error(ctx, "uncheck assessment attendance: update failed",
			log.Err(err),
			log.String("assessment_id", assessmentID),
			log.Strings("attendance_ids", attendanceIDs),
		)
		return err
	}
	return nil
}

type AssessmentAttendanceCondition struct {
	AssessmentIDs []string
	Checked       *bool
}

func (c *AssessmentAttendanceCondition) GetConditions() ([]string, []interface{}) {
	var (
		conditions []string
		values     []interface{}
	)
	if len(c.AssessmentIDs) > 0 {
		conditions = append(conditions, "(assessment_id in (?))")
		values = append(values, c.AssessmentIDs)
	}
	if c.Checked != nil {
		conditions = append(conditions, "(checked = ?)")
		values = append(values, *c.Checked)
	}
	return conditions, values
}

func (c *AssessmentAttendanceCondition) GetPager() *dbo.Pager { return nil }

func (c *AssessmentAttendanceCondition) GetOrderBy() string { return "" }
