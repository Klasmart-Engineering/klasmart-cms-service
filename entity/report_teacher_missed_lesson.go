package entity

type TeacherLoadMissedLessonsRequest struct {
	TeacherId string    `json:"teacher_id" form:"teacher_id" binding:"required"`
	ClassIDs  []string  `json:"class_ids" form:"class_ids" binding:"required"`
	Duration  TimeRange `json:"duration" form:"duration" binding:"required"`
	Page      int       `json:"page" form:"page" binding:"required"`
	PageSize  int       `json:"page_size" form:"page_size" binding:"required"`
}
type TeacherLoadMissedLessonsResponse struct {
	List  []*TeacherLoadMissedLesson `json:"list"`
	Total int                        `json:"total"`
}

type TeacherLoadMissedLesson struct {
	LessonType  string `gorm:"column:class_type" json:"class_type"`
	LessonName  string `gorm:"column:title" json:"title"`
	ClassId     string `gorm:"column:class_id" json:"class_id"`
	NoOfStudent int    `gorm:"column:no_of_student" json:"no_of_student"`
	StartDate   int64  `gorm:"column:start_date" json:"start_date"`
	EndDate     int64  `gorm:"column:end_date" json:"end_date"`
}
