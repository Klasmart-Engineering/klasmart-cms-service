package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

var (
	ScheduleNotFound = errors.New("schedule not found")
	NoPlanInSchedule = errors.New("no schedule in plan")
)

type IH5PEventModel interface {
	CreateH5PEvent(ctx context.Context, tx *dbo.DBContext, req *entity.CreateH5pEventRequest, operator *entity.Operator) error
	Query(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, c da.H5PEventCondition) ([]*entity.H5PEvent, error)
}

type H5PEventModel struct{}

func (h *H5PEventModel) Query(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, c da.H5PEventCondition) ([]*entity.H5PEvent, error) {
	var r []*entity.H5PEvent
	if err := da.GetH5PEventDA().QueryTx(ctx, tx, c, &r); err != nil {
		log.Error(ctx, "Query: call GetH5PEventDA().Query failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("condition", c),
		)
		return nil, err
	}
	return r, nil
}

func (h *H5PEventModel) CreateH5PEvent(ctx context.Context, tx *dbo.DBContext, req *entity.CreateH5pEventRequest, operator *entity.Operator) error {
	//Get plan id by schedule
	schedule, err := GetScheduleModel().GetPlainByID(ctx, req.ScheduleID)
	if err == constant.ErrRecordNotFound {
		log.Error(ctx, "ScheduleID not found",
			log.Err(err),
			log.Any("op", operator),
			log.String("scheduleID", req.ScheduleID))
		return ScheduleNotFound
	}
	if err != nil {
		log.Error(ctx, "GetPlainByID error",
			log.Err(err),
			log.Any("op", operator),
			log.String("scheduleID", req.ScheduleID))
		return err
	}
	//if plan in schedule is nil, error
	if schedule.LessonPlanID == "" {
		log.Error(ctx, "lesson plan is is nil",
			log.Err(err),
			log.Any("op", operator),
			log.Any("schedule", schedule),
			log.String("scheduleID", req.ScheduleID))
		return NoPlanInSchedule
	}

	//parse schedule request to db object
	event, err := req.ToEventObject(ctx, schedule.LessonPlanID, operator)
	if err != nil {
		log.Error(ctx, "ToEventObject error",
			log.Err(err),
			log.Any("op", operator),
			log.Any("req", req),
			log.String("plan", schedule.LessonPlanID))
		return err
	}

	//do insert
	event.ID = utils.NewID()
	err = da.GetH5PEventDA().Create(ctx, tx, event)
	if err != nil {
		log.Error(ctx, "Create event error",
			log.Err(err),
			log.Any("op", operator),
			log.Any("req", req),
			log.Any("event", event))
		return err
	}
	return nil
}

var (
	_h5pEventModel     IH5PEventModel
	_h5pEventModelOnce sync.Once
)

func GetH5PEventModel() IH5PEventModel {
	_h5pEventModelOnce.Do(func() {
		_h5pEventModel = new(H5PEventModel)
	})
	return _h5pEventModel
}
