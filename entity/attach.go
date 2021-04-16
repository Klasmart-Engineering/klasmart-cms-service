package entity

type Attach struct {
	ID         string `gorm:"column:id,primary_key"`
	MasterID   string `gorm:"column:master_id"`
	MasterType string `gorm:"column:master_type"`
	AttachID   string `gorm:"column:attach_id"`
	AttachType string `gorm:"column:attach_type"`
}

const (
	ProgramType     = "program"
	SubjectType     = "subject"
	CategoryType    = "category"
	SubcategoryType = "subcategory"
	GradeType       = "grade"
	AgeType         = "age"
	MilestoneType   = "milestone"
	OutcomeType     = "outcome"
)

const (
	AttachMilestoneTable = "milestones_attaches"
	AttachOutcomeTable   = "outcomes_attaches"
)
