package external

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"

type APCondition struct {
	OrganizationID entity.NullString
	Status         NullAPStatus
	System         entity.NullBool
}

func NewCondition(options ...APOption) *APCondition {
	condition := new(APCondition)
	for _, option := range options {
		option(condition)
	}

	return condition
}

type APOption func(*APCondition)

func WithOrganization(organizationID string) APOption {
	return func(c *APCondition) {
		c.OrganizationID = entity.NullString{String: organizationID, Valid: true}
	}
}

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
