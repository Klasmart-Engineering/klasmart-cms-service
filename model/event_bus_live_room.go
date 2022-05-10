package model

import (
	"context"
	"sync"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

// End of class, callback event
const (
	BusTopicLiveRoomEndClass utils.BusTopic = "LiveRoomEndClass"
)

type BusTopicLiveRoomEndClassFunc func(ctx context.Context, op *entity.Operator, event *v2.ScheduleEndClassCallBackReq) error

type ILiveRoomEventBus interface {
	SubEndClass(handler BusTopicLiveRoomEndClassFunc) error
	PubEndClass(ctx context.Context, op *entity.Operator, event *v2.ScheduleEndClassCallBackReq) error
}

type liveRoomEventBus struct {
	bus *utils.AsyncEventBus
}

func (b *liveRoomEventBus) SubEndClass(handler BusTopicLiveRoomEndClassFunc) error {
	return b.bus.Sub(BusTopicLiveRoomEndClass, handler)
}

func (b *liveRoomEventBus) PubEndClass(ctx context.Context, op *entity.Operator, event *v2.ScheduleEndClassCallBackReq) error {
	return b.bus.Pub(BusTopicLiveRoomEndClass, ctx, op, event)
}

var (
	_liveRoomBusEventOnce sync.Once
	_liveRoomBusModel     ILiveRoomEventBus
)

func GetLiveRoomEventBusModel() ILiveRoomEventBus {
	_liveRoomBusEventOnce.Do(func() {
		bus := &liveRoomEventBus{
			bus: utils.NewAsyncEventBus(),
		}

		bus.SubEndClass(GetAssessmentInternalModel().ScheduleEndClassCallback)
		bus.SubEndClass(GetClassesAssignmentsModel().CreateRecord)

		_liveRoomBusModel = bus
	})
	return _liveRoomBusModel
}
