package entity

const (
	KindMileStone = "milestones"
	KindOutcome   = "outcomes"
)

type ShortcodeElement struct {
	Shortcode string `gorm:"column:shortcode"`
}
