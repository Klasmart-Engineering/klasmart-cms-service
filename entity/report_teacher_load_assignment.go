package entity

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

type TeacherLoadAssignmentRequest struct {
	TeacherIDList []string `json:"teacher_id_list"`
	ClassIDList   []string `json:"class_id_list"`

	// one of study, home_fun
	ClassTypeList ClassTypeList `json:"class_type_list" binding:"oneOf=study,home_fun"`
	Duration      TimeRange     `json:"duration"`
}
type ClassTypeList []string

func (ctl ClassTypeList) Validate(ctx context.Context) (err error) {
	if len(ctl) < 1 {
		err = constant.ErrInvalidArgs
		log.Error(ctx, "class_type_list is required", log.Err(err), log.Any("class_type_list", ctl))
		return
	}
	for _, s := range ctl {
		if s != "study" && s != "home_fun" {
			err = constant.ErrInvalidArgs
			log.Error(ctx, "invalid class_type, class_type should be one of study,home_fun", log.Err(err), log.Any("class_type_list", ctl))
		}
	}
	return
}

type TeacherLoadAssignmentResponse struct {
	TeacherID                  string  `json:"teacher_id"`
	TeacherName                string  `json:"teacher_name"`
	CountOfClasses             int64   `json:"count_of_classes"`
	CountOfStudents            int64   `json:"count_of_students"`
	CountOfScheduledAssignment int64   `json:"count_of_scheduled_assignment"`
	CountOfCompletedAssignment int64   `json:"count_of_completed_assignment"`
	FeedbackPercentage         float64 `json:"feedback_percentage"`
	CountOfPendingAssignment   int64   `json:"count_of_pending_assignment"`
	AvgDaysOfPendingAssignment int64   `json:"avg_days_of_pending_assignment"`
}
