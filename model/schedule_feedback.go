package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
	"sync"
	"time"
)

type IScheduleFeedbackModel interface {
	Add(ctx context.Context, op *entity.Operator, input *entity.ScheduleFeedbackAddInput) (string, error)
	ExistByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) (bool, error)
	Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleFeedbackCondition) ([]*entity.ScheduleFeedbackView, error)
	GetNewest(ctx context.Context, op *entity.Operator, condition *da.ScheduleFeedbackCondition) (*entity.ScheduleFeedbackView, error)
}

type scheduleFeedbackModel struct {
}

func (s *scheduleFeedbackModel) GetNewest(ctx context.Context, op *entity.Operator, condition *da.ScheduleFeedbackCondition) (*entity.ScheduleFeedbackView, error) {
	//var dataList []*entity.ScheduleFeedback
	//err := da.GetScheduleFeedbackDA().Query(ctx, condition, dataList)
	//if err != nil {
	//	log.Error(ctx, "query error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
	//	return nil, err
	//}
	//if len(dataList) <= 0 {
	//	log.Warn(ctx, "not found", log.Any("op", op), log.Any("condition", condition))
	//	return nil, constant.ErrRecordNotFound
	//}
	return nil, nil
}

func (s *scheduleFeedbackModel) Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleFeedbackCondition) ([]*entity.ScheduleFeedbackView, error) {
	var dataList []*entity.ScheduleFeedback
	err := da.GetScheduleFeedbackDA().Query(ctx, condition, dataList)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}
	result := make([]*entity.ScheduleFeedbackView, len(dataList))
	for i, item := range dataList {
		result[i] = &entity.ScheduleFeedbackView{ScheduleFeedback: *item}
	}
	return result, nil
}

func (s *scheduleFeedbackModel) ExistByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) (bool, error) {
	condition := &da.ScheduleFeedbackCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
	}
	count, err := da.GetScheduleFeedbackDA().Count(ctx, condition, &entity.ScheduleFeedback{})
	if err != nil {
		log.Error(ctx, "insert error", log.Err(err), log.Any("op", op), log.String("scheduleID", scheduleID))
		return false, err
	}
	return count > 0, nil
}

func (s *scheduleFeedbackModel) Add(ctx context.Context, op *entity.Operator, input *entity.ScheduleFeedbackAddInput) (string, error) {
	err := s.verifyScheduleFeedback(ctx, op, input)
	if err != nil {
		return "", err
	}

	data := &entity.ScheduleFeedback{
		ID:            utils.NewID(),
		ScheduleID:    input.ScheduleID,
		UserID:        op.UserID,
		AssignmentUrl: input.AssignmentUrl,
		Comment:       input.Comment,
		CreateAt:      time.Now().Unix(),
		UpdateAt:      0,
		DeleteAt:      0,
	}
	_, err = da.GetScheduleFeedbackDA().Insert(ctx, data)
	if err != nil {
		log.Error(ctx, "insert error", log.Err(err), log.Any("op", op), log.Any("input", input))
		return "", err
	}
	return data.ID, nil
}

func (s *scheduleFeedbackModel) verifyScheduleFeedback(ctx context.Context, op *entity.Operator, input *entity.ScheduleFeedbackAddInput) error {
	if strings.TrimSpace(input.ScheduleID) == "" ||
		strings.TrimSpace(input.AssignmentUrl) == "" {
		log.Info(ctx, "invalid args", log.Any("op", op), log.Any("input", input))
		return constant.ErrInvalidArgs
	}
	exist, err := GetScheduleModel().ExistScheduleByID(ctx, input.ScheduleID)
	if err != nil {
		log.Error(ctx, "GetScheduleModel.ExistScheduleByID error", log.Err(err), log.Any("input", input))
		return err
	}
	if !exist {
		log.Info(ctx, "schedule id not found", log.Any("op", op), log.Any("input", input))
		return constant.ErrRecordNotFound
	}
	return nil
}

var (
	_scheduleFeedbackOnce  sync.Once
	_scheduleFeedbackModel IScheduleFeedbackModel
)

func GetScheduleFeedbackModel() IScheduleFeedbackModel {
	_scheduleFeedbackOnce.Do(func() {
		_scheduleFeedbackModel = &scheduleFeedbackModel{}
	})
	return _scheduleFeedbackModel
}
