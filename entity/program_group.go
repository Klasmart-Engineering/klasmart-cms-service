package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ProgramGroup struct {
	ProgramID string           `gorm:"column:program_id;type:varchar(100);PRIMARY_KEY"`
	GroupName ProgramGroupName `gorm:"column:group_name;type:varchar(100);not null"`
}

func (e ProgramGroup) TableName() string {
	return constant.TableNameProgramGroup
}

type ProgramGroupName string

const (
	ProgramGroupBadaESL                 ProgramGroupName = "BadaESL"
	ProgramGroupBadaSTEAM               ProgramGroupName = "BadaSTEAM"
	ProgramGroupBadaMore                ProgramGroupName = "More"
	ProgramGroupBadaMoreFeaturedContent ProgramGroupName = "More Featured Content"
)

func (t ProgramGroupName) Valid() bool {
	switch t {
	case ProgramGroupBadaESL,
		ProgramGroupBadaSTEAM,
		ProgramGroupBadaMore,
		ProgramGroupBadaMoreFeaturedContent:
		return true
	default:
		return false
	}
}
