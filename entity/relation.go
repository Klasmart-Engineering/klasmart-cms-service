package entity

type Relation struct {
	ID           string       `gorm:"column:id,primary_key"`
	MasterID     string       `gorm:"column:master_id"`
	MasterType   RelationType `gorm:"column:master_type"`
	RelationID   string       `gorm:"column:relation_id"`
	RelationType RelationType `gorm:"column:relation_type"`
}

type RelationType string

const (
	ProgramType     RelationType = "program"
	SubjectType     RelationType = "subject"
	CategoryType    RelationType = "category"
	SubcategoryType RelationType = "subcategory"
	GradeType       RelationType = "grade"
	AgeType         RelationType = "age"
	MilestoneType   RelationType = "milestone"
	OutcomeType     RelationType = "outcome"
)

const (
	MilestoneRelationTable = "milestones_relations"
	OutcomeRelationTable   = "outcomes_relations"
)

type MilestoneRelation struct {
	Relation
}

func (MilestoneRelation) TableName() string {
	return MilestoneRelationTable
}

type OutcomeRelation struct {
	Relation
}

func (OutcomeRelation) TableName() string {
	return OutcomeRelationTable
}
