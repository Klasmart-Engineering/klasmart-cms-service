package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type TeacherSchedule struct {
	TeacherID  string `dynamodbav:"teacher_id"`
	ScheduleID string `dynamodbav:"schedule_id"`
	StartAt    int64  `dynamodbav:"start_at"`
}

func (TeacherSchedule) TableName() string {
	return constant.TableNameTeacherSchedule
}

type DeleteTeacherScheduleViewData struct {
}
