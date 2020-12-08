package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ProgramAge struct {
	ID        string `gorm:"column:id;PRIMARY_KEY"`
	ProgramID string `gorm:"column:program_id;type:varchar(256);not null"`
	AgeID     string `gorm:"column:age_id;type:varchar(256);not null"`
}

func (e ProgramAge) TableName() string {
	return constant.TableNameProgramAge
}

func (e ProgramAge) GetID() interface{} {
	return e.ID
}
