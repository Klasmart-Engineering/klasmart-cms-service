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
	GetAttendanceIDsByAssessmentID(ctx context.Context, tx dbo.DBContext, assessmentID string) ([]string, error)
	BatchInsert(ctx context.Context, tx dbo.DBContext, items []*entity.AssessmentAttendance) error
	DeleteByAssessmentID(ctx context.Context, tx dbo.DBContext, assessmentID string) error
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

type assessmentAttendanceDA struct{}

func (*assessmentAttendanceDA) GetAttendanceIDsByAssessmentID(ctx context.Context, tx dbo.DBContext, assessmentID string) ([]string, error) {
	var items []entity.AssessmentAttendance
	if err := tx.Where("assessment_id = ?", assessmentID).Find(&items).Error; err != nil {
		return nil, err
	}
	var ids []string
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids, nil
}

func (*assessmentAttendanceDA) BatchInsert(ctx context.Context, tx dbo.DBContext, items []*entity.AssessmentAttendance) error {
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

func (*assessmentAttendanceDA) DeleteByAssessmentID(ctx context.Context, tx dbo.DBContext, assessmentID string) error {
	if err := tx.Where("assessment_id", assessmentID).Delete(entity.AssessmentAttendance{}).Error; err != nil {
		log.Error(ctx, "delete attendances by id: delete failed from db",
			log.Err(err),
			log.String("assessment_id", assessmentID),
		)
		return err
	}
	return nil
}
