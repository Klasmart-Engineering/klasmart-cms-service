package entity

import "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

type ScheduleReview struct {
	ScheduleID     string                  `json:"schedule_id" gorm:"column:schedule_id;type:varchar(100)"`
	StudentID      string                  `json:"student_id" gorm:"column:student_id;type:varchar(100)"`
	ReviewStatus   ScheduleReviewStatus    `json:"review_status" gorm:"column:review_status;type:varchar(100)"`
	LiveLessonPlan *ScheduleLiveLessonPlan `gorm:"column:live_lesson_plan;type:json" json:"live_lesson_plan"`
}

func (s ScheduleReview) TableName() string {
	return constant.TableNameScheduleReview
}
