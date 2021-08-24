package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

const (
	KFPSAttachment KFPSClassify = "attachment"
)

type KFPSClassify string

func (k KFPSClassify) Classify() string {
	return constant.KFPSMessageQueuePrefix + string(k)
}
