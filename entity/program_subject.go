package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ProgramSubject struct {
	ID        string `gorm:"column:id;PRIMARY_KEY"`
	ProgramID string `gorm:"column:program_id;type:varchar(256);not null"`
	SubjectID string `gorm:"column:subject_id;type:varchar(256);not null"`
}

func (e ProgramSubject) TableName() string {
	return constant.TableNameProgramSubject
}

func (e ProgramSubject) GetID() interface{} {
	return e.ID
}
