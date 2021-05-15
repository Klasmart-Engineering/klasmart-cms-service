package entity

type Assessment struct {
	ID           string           `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
	ScheduleID   string           `gorm:"column:schedule_id;type:varchar(64);not null" json:"schedule_id"`
	Type         AssessmentType   `gorm:"column:type;type:varchar(1024);not null" json:"type"` // add: 2021-05-15
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

type AssessmentType string

const (
	AssessmentTypeOutcomeClassAndLive = "outcome_class_and_live"
	AssessmentTypeContentClassAndLive = "content_class_and_live"
	AssessmentTypeContentStudy        = "content_study"
)

func (t AssessmentType) Valid() bool {
	switch t {
	case AssessmentTypeOutcomeClassAndLive, AssessmentTypeContentClassAndLive, AssessmentTypeContentStudy:
		return true
	default:
		return false
	}
}

type NullAssessmentType struct {
	Value AssessmentType
	Valid bool
}

func (Assessment) TableName() string {
	return "assessments"
}

type UpdateAssessmentAction string

const (
	UpdateAssessmentActionSave     UpdateAssessmentAction = "save"
	UpdateAssessmentActionComplete UpdateAssessmentAction = "complete"
)

func (a UpdateAssessmentAction) Valid() bool {
	switch a {
	case UpdateAssessmentActionSave, UpdateAssessmentActionComplete:
		return true
	default:
		return false
	}
}
