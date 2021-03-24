package external

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

// APCondition academic profile query condition
type APCondition struct {
	Status NullAPStatus
	System entity.NullBool
}

func NewCondition(options ...APOption) *APCondition {
	condition := new(APCondition)
	for _, option := range options {
		option(condition)
	}

	return condition
}

// APOption academic profile query option
type APOption func(*APCondition)

func WithStatus(status APStatus) APOption {
	return func(c *APCondition) {
		c.Status = NullAPStatus{Status: status, Valid: true}
	}
}

func WithSystem(system bool) APOption {
	return func(c *APCondition) {
		c.System = entity.NullBool{Bool: system, Valid: true}
	}
}
