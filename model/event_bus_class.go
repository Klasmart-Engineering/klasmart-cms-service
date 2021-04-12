package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

const (
	BusTopicClassAddMembers    BusTopic = "ClassAddMembers"
	BusTopicClassDeleteMembers BusTopic = "ClassDeleteMembers"
)

type BusTopicClassAddMembersFunc func(ctx context.Context, op *entity.Operator, event *entity.ScheduleClassEvent) error

type IClassEventBus interface {
	SubAddMembers(handler BusTopicClassAddMembersFunc) error
	PubAddMembers(ctx context.Context, op *entity.Operator, event *entity.ScheduleClassEvent) error
}

type classEventBus struct {
	bus *AsyncEventBus
}

func (b *classEventBus) SubAddMembers(handler BusTopicClassAddMembersFunc) error {
	return b.bus.Sub(BusTopicClassAddMembers, handler)
}

func (b *classEventBus) PubAddMembers(ctx context.Context, op *entity.Operator, event *entity.ScheduleClassEvent) error {
	return b.bus.Pub(BusTopicClassAddMembers, ctx, op, event)
}

var (
	_busEventOnce sync.Once
	_busModel     IClassEventBus
)

func GetClassEventBusModel() IClassEventBus {
	_busEventOnce.Do(func() {
		bus := &classEventBus{
			bus: NewAsyncEventBus(),
		}
		bus.SubAddMembers(GetScheduleEventModel().AddUserEvent)
		_busModel = bus
	})
	return _busModel
}
