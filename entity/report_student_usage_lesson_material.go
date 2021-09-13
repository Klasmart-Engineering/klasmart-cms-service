package entity

import (
	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type StudentUsageRecord struct {
	ClassType         string     `json:"class_type"  gorm:"column:class_type"  enums:"live,class,study,home fun,task" `
	RoomID            string     `json:"room_id" gorm:"column:room_id" `
	LessonMaterialUrl string     `json:"lesson_material_url"  gorm:"column:lesson_material_url"  `
	ContentType       string     `json:"content_type"  gorm:"column:content_type"   enums:"h5p, audio, video, image, document"`
	ActionType        string     `json:"action_type"  gorm:"column:action_type"   enums:"view"`
	Timestamp         int64      `json:"timestamp"  gorm:"column:timestamp"  `
	Students          []*Student `json:"students"  gorm:"-"  `

	// below fields not from api
	ID               string `json:"id" gorm:"column:id" `
	ScheduleStartAt  int64  `json:"schedule_start_at" gorm:"column:schedule_start_at" `
	LessonPlanID     string `json:"lesson_plan_id" gorm:"column:lesson_plan_id" `
	StudentUserID    string `json:"student_user_id" gorm:"column:student_user_id" `
	StudentEmail     string `json:"student_email" gorm:"column:student_email" `
	StudentName      string `json:"student_name" gorm:"column:student_name" `
	LessonMaterialID string `json:"lesson_material_id" gorm:"column:lesson_material_id" `
	ClassID          string `json:"class_id" gorm:"column:class_id" `
}

func (StudentUsageRecord) TableName() string {
	return constant.TableNameStudentUsageRecord
}
func (r *StudentUsageRecord) GetBatchInsertColsAndValues() (cols []string, values []interface{}) {
	cols = append(cols, "id")
	values = append(values, r.ID)

	cols = append(cols, "class_type")
	values = append(values, r.ClassType)

	cols = append(cols, "room_id")
	values = append(values, r.RoomID)

	cols = append(cols, "lesson_material_url")
	values = append(values, r.LessonMaterialUrl)

	cols = append(cols, "content_type")
	values = append(values, r.ContentType)

	cols = append(cols, "action_type")
	values = append(values, r.ActionType)

	cols = append(cols, "timestamp")
	values = append(values, r.Timestamp)

	cols = append(cols, "student_user_id")
	values = append(values, r.StudentUserID)

	cols = append(cols, "student_email")
	values = append(values, r.StudentEmail)

	cols = append(cols, "student_name")
	values = append(values, r.StudentName)

	cols = append(cols, "lesson_material_id")
	values = append(values, r.LessonMaterialID)

	cols = append(cols, "lesson_plan_id")
	values = append(values, r.LessonPlanID)

	cols = append(cols, "schedule_start_at")
	values = append(values, r.ScheduleStartAt)

	cols = append(cols, "class_id")
	values = append(values, r.ClassID)

	return
}

type Student struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}
type StudentUsageRecordInJwt struct {
	StudentUsageRecord
	*jwt.StandardClaims
}
type JwtToken struct {
	Token string `json:"token"`
}

type StudentUsageMaterialReportRequest struct {
	TimeRangeList   []TimeRange `json:"time_range_list" form:"time_range_list"`
	ClassIDList     []string    `json:"class_id_list" form:"class_id_list"`
	ContentTypeList []string    `json:"content_type_list" form:"content_type_list"`
}

type MaterialUsage struct {
	TimeRange   TimeRange `json:"time_range" gorm:"column:time_range" `
	ClassID     string    `json:"class_id" gorm:"column:class_id" `
	ContentType string    `json:"content_type" gorm:"column:content_type" `
	UsedCount   int64     `json:"used_count" gorm:"column:used_count" `
}

type StudentUsageMaterialReportResponse struct {
	Request        *StudentUsageMaterialReportRequest `json:"request"`
	ClassUsageList []*ClassUsage                      `json:"class_usage_list"`
}
type ClassUsage struct {
	ID               string          `json:"id"`
	ContentUsageList []*ContentUsage `json:"content_usage_list"`
}

type ContentUsage struct {
	TimeRange TimeRange `json:"time_range"`
	Type      string    `json:"type"`
	Count     int64     `json:"count"`
}

type StudentUsageMaterialViewCountReportRequest struct {
	TimeRangeList   []TimeRange `json:"time_range_list" form:"time_range_list"`
	ClassIDList     []string    `json:"class_id_list" form:"class_id_list"`
	ContentTypeList []string    `json:"content_type_list" form:"content_type_list"`
}

type StudentUsageMaterialViewCountReportResponse struct {
	Request          *StudentUsageMaterialViewCountReportRequest `json:"request"`
	ContentUsageList []*ContentUsage                             `json:"content_usage_list"`
}
