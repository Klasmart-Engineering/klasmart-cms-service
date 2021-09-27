package entity

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

type TeacherLoadAssignmentRequest struct {
	TeacherIDList []string `json:"teacher_id_list" binding:"gt=0"`
	ClassIDList   []string `json:"class_id_list"  binding:"gt=0"`

	// one of study, home_fun
	ClassTypeList ClassTypeList `json:"class_type_list"  binding:"gt=0"`
	Duration      TimeRange     `json:"duration"`
}
type ClassTypeList []constant.ReportClassType

func (ctl ClassTypeList) Validate(ctx context.Context) (err error) {
	if len(ctl) < 1 {
		err = constant.ErrInvalidArgs
		log.Error(ctx, "class_type_list is required", log.Err(err), log.Any("class_type_list", ctl))
		return
	}
	for _, s := range ctl {
		if s != constant.ReportClassTypeStudy && s != constant.ReportClassTypeHomeFun {
			err = constant.ErrInvalidArgs
			log.Error(ctx, "invalid class_type, class_type should be one of study,home_fun", log.Err(err), log.Any("class_type_list", ctl))
		}
	}
	return
}
func (ctl ClassTypeList) Contains(t constant.ReportClassType) bool {
	for _, classType := range ctl {
		if classType == t {
			return true
		}
	}
	return false
}

type TeacherLoadAssignmentResponseItem struct {
	TeacherID string `json:"teacher_id" gorm:"column:teacher_id" `
	// TeacherName just used by font-end: generate swagger json --> generate typescript class
	TeacherName                string  `json:"teacher_name" gorm:"column:teacher_name" `
	CountOfClasses             int64   `json:"count_of_classes" gorm:"column:count_of_classes" `
	CountOfStudents            int64   `json:"count_of_students" gorm:"column:count_of_students" `
	CountOfScheduledAssignment int64   `json:"count_of_scheduled_assignment" gorm:"column:count_of_scheduled_assignment" `
	CountOfCompletedAssignment int64   `json:"count_of_completed_assignment" gorm:"column:count_of_completed_assignment" `
	CountOfCommentedAssignment int64   `json:"-" gorm:"column:count_of_commented_assignment" `
	FeedbackPercentage         float64 `json:"feedback_percentage" gorm:"column:feedback_percentage" `
	CountOfPendingAssignment   int64   `json:"count_of_pending_assignment" gorm:"column:count_of_pending_assignment" `
	AvgDaysOfPendingAssignment float64 `json:"avg_days_of_pending_assignment" gorm:"column:avg_days_of_pending_assignment" `
}
type TeacherLoadAssignmentResponseItemSlice []*TeacherLoadAssignmentResponseItem

func (s TeacherLoadAssignmentResponseItemSlice) MapTeacherID() (m map[string]*TeacherLoadAssignmentResponseItem) {
	m = map[string]*TeacherLoadAssignmentResponseItem{}
	for _, item := range s {
		m[item.TeacherID] = item
	}
	return
}
