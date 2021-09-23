package entity

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type TypedError interface {
	error
	ErrorType() string
}
type ExternalError struct {
	Err  error
	Type constant.InternalErrorType
}

func (ee *ExternalError) ErrorType() string {
	if ee == nil {
		return ""
	}
	return string(ee.Type)
}
func (ee *ExternalError) Error() string {
	if ee == nil || ee.Err == nil {
		return ""
	}
	return ee.Err.Error()
}
