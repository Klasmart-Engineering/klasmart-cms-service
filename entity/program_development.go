package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ProgramDevelopment struct {
	ID            string `gorm:"column:id;PRIMARY_KEY"`
	ProgramID     string `gorm:"column:program_id;type:varchar(256);not null"`
	DevelopmentID string `gorm:"column:development_id;type:varchar(256);not null"`
}

func (e ProgramDevelopment) TableName() string {
	return constant.TableNameProgramDevelopment
}

func (e ProgramDevelopment) GetID() interface{} {
	return e.ID
}
