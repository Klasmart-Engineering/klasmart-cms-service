package entity

import "strings"

type Assessment struct {
	ID         string `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	ScheduleID string `gorm:"column:schedule_id;type:varchar(64);not null" json:"schedule_id"`
	//Type         AssessmentType   `gorm:"column:type;type:varchar(1024);not null" json:"type"` // add: 2021-05-15,delete 2021-05-31
	Title        string           `gorm:"column:title;type:varchar(1024);not null" json:"title"`
	CompleteTime int64            `gorm:"column:complete_time;type:bigint;not null" json:"complete_time"`
	Status       AssessmentStatus `gorm:"column:status;type:varchar(128);not null" json:"status"`

	CreateAt int64 `gorm:"column:create_at;type:bigint;not null" json:"create_at"`
	UpdateAt int64 `gorm:"column:update_at;type:bigint;not null" json:"update_at"`
	DeleteAt int64 `gorm:"column:delete_at;type:bigint;not null" json:"delete_at"`

	// Union Fields
	ClassLength  int   `gorm:"column:class_length;type:int;not null" json:"class_length"`
	ClassEndTime int64 `gorm:"column:class_end_time;type:bigint;not null" json:"class_end_time"`
}

func (Assessment) TableName() string {
	return "assessments"
}

type AssessmentStudentViewH5PItem struct {
	StudentID       string                                    `json:"student_id"`
	StudentName     string                                    `json:"student_name"`
	Comment         string                                    `json:"comment"`
	LessonMaterials []*AssessmentStudentViewH5PLessonMaterial `json:"lesson_materials"`
}

type AssessmentStudentViewH5PItemsOrder []*AssessmentStudentViewH5PItem

func (h AssessmentStudentViewH5PItemsOrder) Len() int {
	return len(h)
}

func (h AssessmentStudentViewH5PItemsOrder) Less(i, j int) bool {
	return strings.ToLower(h[i].StudentName) < strings.ToLower(h[j].StudentName)
}

func (h AssessmentStudentViewH5PItemsOrder) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

type AssessmentStudentViewH5PLessonMaterial struct {
	LessonMaterialID   string   `json:"lesson_material_id"`
	LessonMaterialName string   `json:"lesson_material_name"`
	LessonMaterialType string   `json:"lesson_material_type"`
	Answer             string   `json:"answer"`
	MaxScore           float64  `json:"max_score"`
	AchievedScore      float64  `json:"achieved_score"`
	Attempted          bool     `json:"attempted"`
	IsH5P              bool     `json:"is_h5p"`
	OutcomeNames       []string `json:"outcome_names"`
}

type UpdateAssessmentH5PStudent struct {
	StudentID       string                               `json:"student_id"`
	Comment         string                               `json:"comment"`
	LessonMaterials []*UpdateAssessmentH5PLessonMaterial `json:"lesson_materials"`
}

type UpdateAssessmentH5PLessonMaterial struct {
	LessonMaterialID string  `json:"lesson_material_id"`
	AchievedScore    float64 `json:"achieved_score"`
}
