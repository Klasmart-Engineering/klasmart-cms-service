package entity

type DevelopmentSkill struct {
	ID            string `gorm:"column:id;PRIMARY_KEY"`
	DevelopmentID string `gorm:"column:development_id;type:varchar(256);not null"`
	SkillID       string `gorm:"column:skill_id;type:varchar(256);not null"`
}

func (e DevelopmentSkill) TableName() string {
	return "developments_skills"
}

func (e DevelopmentSkill) GetID() interface{} {
	return e.ID
}
