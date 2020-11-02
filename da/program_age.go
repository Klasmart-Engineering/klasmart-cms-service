package da

import (
	"sync"
)

type IProgramAgeDA interface {
}

type programAgeDA struct {
}

var (
	_programAgeOnce sync.Once
	_programAgeDA   IProgramAgeDA
)

func GetProgramAgeDA() IProgramAgeDA {
	_programAgeOnce.Do(func() {
		_programAgeDA = &programAgeDA{}
	})
	return _programAgeDA
}
