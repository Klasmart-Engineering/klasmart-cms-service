package utils

import (
	"gopkg.in/go-playground/validator.v9"
	"sync"
)

var (
	_validatorOnce sync.Once
	_validator     *validator.Validate
)

func GetValidator() *validator.Validate {
	_validatorOnce.Do(func() {
		_validator = validator.New()
	})
	return _validator
}
