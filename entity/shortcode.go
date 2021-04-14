package entity

const (
	KindMileStone = "milestone"
	KindOutcome   = "learning_outcomes"
)

type ShortcodeElement struct {
	Shortcode string `gorm:"column:shortcode"`
}
