package utils

import (
	"sync"
	"time"
)

type Stopwatch struct {
	mutex         sync.Mutex
	remains       int64
	start         time.Time
	totalDuration time.Duration
}

func (s *Stopwatch) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.remains++
	if s.remains > 1 {
		return
	}

	s.start = time.Now()
}

func (s *Stopwatch) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.remains--
	if s.remains > 0 {
		return
	}

	if s.remains < 0 {
		s.remains = 0
	}

	s.totalDuration += time.Since(s.start)
}

func (s *Stopwatch) Duration() time.Duration {
	return s.totalDuration
}
