package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IOutcomeAttendanceDA interface {
	GetAttendanceIDsByOutcomeID(ctx context.Context, tx *dbo.DBContext, outcomeID string) ([]string, error)
	BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.OutcomeAttendance) error
	//DeleteByOutcomeID(ctx context.Context, tx dbo.DBContext, outcomeID string) error
	BatchDeleteByOutcomeIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string) error
}

var (
	outcomeAttendanceDAInstance     IOutcomeAttendanceDA
	outcomeAttendanceDAInstanceOnce = sync.Once{}
)

func GetOutcomeAttendanceDA() IOutcomeAttendanceDA {
	outcomeAttendanceDAInstanceOnce.Do(func() {
		outcomeAttendanceDAInstance = &outcomeAttendanceDA{}
	})
	return outcomeAttendanceDAInstance
}

type outcomeAttendanceDA struct{}

func (*outcomeAttendanceDA) GetAttendanceIDsByOutcomeID(ctx context.Context, tx *dbo.DBContext, outcomeID string) ([]string, error) {
	var items []entity.OutcomeAttendance
	if err := tx.Where("outcome_id = ?", outcomeID).Find(&items).Error; err != nil {
		return nil, err
	}
	var ids []string
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids, nil
}

func (*outcomeAttendanceDA) BatchInsert(ctx context.Context, tx *dbo.DBContext, items []*entity.OutcomeAttendance) error {
	if len(items) == 0 {
		return nil
	}
	columns := []string{"id", "outcome_id", "attendance_id"}
	var values [][]interface{}
	for _, item := range items {
		if item.ID == "" {
			item.ID = utils.NewID()
		}
		values = append(values, []interface{}{item.ID, item.OutcomeID, item.AttendanceID})
	}
	template := SQLBatchInsert(entity.OutcomeAttendance{}.TableName(), columns, values)
	if err := tx.Exec(template.Format, template.Values...).Error; err != nil {
		log.Error(ctx, "batch insert outcomes_attendances: batch insert failed",
			log.Err(err),
			log.Any("items", items),
		)
		return err
	}
	return nil
}

func (*outcomeAttendanceDA) BatchDeleteByOutcomeIDs(ctx context.Context, tx *dbo.DBContext, outcomeIDs []string) error {
	if len(outcomeIDs) == 0 {
		return nil
	}
	if err := tx.Where("outcome_id in ?", outcomeIDs).Delete(entity.OutcomeAttendance{}).Error; err != nil {
		log.Error(ctx, "batch delete attendances by outcome ids: batch delete failed",
			log.Err(err),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return err
	}
	return nil
}
