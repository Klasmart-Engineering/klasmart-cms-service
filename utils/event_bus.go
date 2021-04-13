package utils

import (
	"fmt"

	"reflect"
	"sync"
)

type BusTopic string

type IBus interface {
	Sub(topic BusTopic, handler interface{}) error
	Pub(topic BusTopic, args ...interface{}) error
}

type AsyncEventBus struct {
	handlers map[BusTopic][]reflect.Value
	lock     sync.Mutex
	//isWait   bool
}

func NewAsyncEventBus() *AsyncEventBus {
	return &AsyncEventBus{
		handlers: map[BusTopic][]reflect.Value{},
		lock:     sync.Mutex{},
	}
}

func (b *AsyncEventBus) Sub(topic BusTopic, f interface{}) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	v := reflect.ValueOf(f)
	if v.Type().Kind() != reflect.Func {
		return fmt.Errorf("handler is not a function,%v", v.Type().Kind())
	}

	handler, ok := b.handlers[topic]
	if !ok {
		handler = []reflect.Value{}
	}
	handler = append(handler, v)
	b.handlers[topic] = handler

	return nil
}

func (b *AsyncEventBus) Pub(topic BusTopic, args ...interface{}) error {
	handlers, ok := b.handlers[topic]
	if !ok {
		return fmt.Errorf("not found handler in topic,%v", topic)
	}

	params := make([]reflect.Value, len(args))
	for i, arg := range args {
		params[i] = reflect.ValueOf(arg)
	}
	var wg sync.WaitGroup
	for i := range handlers {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			handlers[i].Call(params)
		}(i)
	}
	wg.Wait()
	return nil
}
