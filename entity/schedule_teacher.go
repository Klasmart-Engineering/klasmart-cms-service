package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ScheduleTeacher struct {
	ID         string `gorm:"column:id;PRIMARY_KEY"`
	TeacherID  string `gorm:"column:teacher_id;type:varchar(100) dynamodbav:"teacher_id"`
	ScheduleID string `gorm:"column:schedule_id;type:varchar(100) dynamodbav:"schedule_id"`
	DeletedAt  int64  `gorm:"column:deleted_at;type:bigint"`
}

func (ScheduleTeacher) TableName() string {
	return constant.TableNameScheduleTeacher
}
