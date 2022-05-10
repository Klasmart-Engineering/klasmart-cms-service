package entity

import "github.com/KL-Engineering/kidsloop-cms-service/constant"

const (
	KFPSAttachment KFPSClassify = "attachment"
)

type KFPSClassify string

func (k KFPSClassify) Classify() string {
	return constant.KFPSMessageQueuePrefix + string(k)
}
