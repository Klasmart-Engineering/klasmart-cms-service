package model

import (
	"sync"
)

const (
	BusTopicClassAddMembers    BusTopic = "ClassAddMembers"
	BusTopicClassDeleteMembers BusTopic = "ClassDeleteMembers"
)

var (
	_busEventOnce sync.Once
	_busModel     IBus
)

func GetClassEventBusModel() IBus {
	_busEventOnce.Do(func() {
		bus := NewAsyncEventBus()
		bus.Sub(BusTopicClassAddMembers, GetScheduleEventModel().AddUserEvent)
		bus.Sub(BusTopicClassDeleteMembers, GetScheduleEventModel().DeleteUserEvent)

		_busModel = bus
	})
	return _busModel
}
