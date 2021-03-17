package utils

import (
	"context"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type Stopwatch struct {
	mutex         sync.RWMutex
	remains       int64
	start         time.Time
	totalDuration time.Duration
}

func NewStopwatch() *Stopwatch {
	return &Stopwatch{}
}

func (s *Stopwatch) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.remains == 0 {
		s.start = time.Now()
	}

	s.remains++
}

func (s *Stopwatch) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.remains--
	if s.remains > 0 {
		return
	}

	s.totalDuration += time.Since(s.start)
	s.remains = 0
}

func (s *Stopwatch) Duration() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.totalDuration
}

func SetupStopwatch(ctx context.Context) context.Context {
	stopwatches := map[string]*Stopwatch{
		string(constant.ExternalStopwatch): NewStopwatch(),
	}

	return context.WithValue(ctx, constant.ContextStopwatchKey, stopwatches)
}

func GetStopwatches(ctx context.Context) (map[string]*Stopwatch, bool) {
	stopwatches := ctx.Value(constant.ContextStopwatchKey)
	if stopwatches == nil {
		log.Debug(ctx, "context stopwatches not found")
		return nil, false
	}

	stopwatchMap, ok := stopwatches.(map[string]*Stopwatch)

	return stopwatchMap, ok
}

func GetStopwatch(ctx context.Context, _type constant.ContextStopwatchType) (*Stopwatch, bool) {
	stopwatchMap, found := GetStopwatches(ctx)
	if !found {
		return nil, false
	}

	stopwatch, found := stopwatchMap[string(_type)]
	return stopwatch, found
}
