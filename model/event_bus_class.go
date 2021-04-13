package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

const (
	BusTopicClassAddMembers    utils.BusTopic = "ClassAddMembers"
	BusTopicClassDeleteMembers utils.BusTopic = "ClassDeleteMembers"
)

type BusTopicClassAddMembersFunc func(ctx context.Context, op *entity.Operator, event *entity.ClassUpdateMembersEvent) error
type BusTopicClassDeleteMembersFunc func(ctx context.Context, op *entity.Operator, event *entity.ClassUpdateMembersEvent) error

type IClassEventBus interface {
	SubAddMembers(handler BusTopicClassAddMembersFunc) error
	PubAddMembers(ctx context.Context, op *entity.Operator, event *entity.ClassUpdateMembersEvent) error

	SubDeleteMembers(handler BusTopicClassDeleteMembersFunc) error
	PubDeleteMembers(ctx context.Context, op *entity.Operator, event *entity.ClassUpdateMembersEvent) error
}

type classEventBus struct {
	bus *utils.AsyncEventBus
}

func (b *classEventBus) SubDeleteMembers(handler BusTopicClassDeleteMembersFunc) error {
	return b.bus.Sub(BusTopicClassDeleteMembers, handler)
}

func (b *classEventBus) PubDeleteMembers(ctx context.Context, op *entity.Operator, event *entity.ClassUpdateMembersEvent) error {
	return b.bus.Pub(BusTopicClassDeleteMembers, ctx, op, event)
}

func (b *classEventBus) SubAddMembers(handler BusTopicClassAddMembersFunc) error {
	return b.bus.Sub(BusTopicClassAddMembers, handler)
}

func (b *classEventBus) PubAddMembers(ctx context.Context, op *entity.Operator, event *entity.ClassUpdateMembersEvent) error {
	return b.bus.Pub(BusTopicClassAddMembers, ctx, op, event)
}

var (
	_busEventOnce sync.Once
	_busModel     IClassEventBus
)

func GetClassEventBusModel() IClassEventBus {
	_busEventOnce.Do(func() {
		bus := &classEventBus{
			bus: utils.NewAsyncEventBus(),
		}
		bus.SubAddMembers(GetScheduleEventModel().AddMembersEvent)
		bus.SubDeleteMembers(GetScheduleEventModel().DeleteMembersEvent)
		_busModel = bus
	})
	return _busModel
}
