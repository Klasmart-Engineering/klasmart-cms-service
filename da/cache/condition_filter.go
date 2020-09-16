package cache

import "gitlab.badanamu.com.cn/calmisland/dbo"

type ConditionEntity struct {
	ConditionKey string
	Conditions *dbo.Conditions
}
type ConditionFilterResult struct {
	ConditionKey string
	IsMapping bool
}

type CachedObject interface{
	ID() string
	Name() string
	IsConditionMapping(cdList []*ConditionEntity) []*ConditionFilterResult
}

type Decoder interface{
	ObjectName() string
	DecodeObject(data string) (CachedObject, error)
	BatchDecodeObject(data string) ([]CachedObject, error)
	DecodeCondition(data string) (*dbo.Conditions, error)
}
