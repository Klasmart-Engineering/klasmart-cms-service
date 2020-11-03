package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ProgramGrade struct {
	ID        string `gorm:"column:id;PRIMARY_KEY"`
	ProgramID string `gorm:"column:program_id;type:varchar(256);not null"`
	GradeID   string `gorm:"column:grade_id;type:varchar(256);not null"`
}

func (e ProgramGrade) TableName() string {
	return constant.TableNameProgramGrade
}

func (e ProgramGrade) GetID() interface{} {
	return e.ID
}
