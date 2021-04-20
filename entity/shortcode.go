package entity

type ShortcodeKind string

const (
	KindMileStone ShortcodeKind = "milestones"
	KindOutcome   ShortcodeKind = "outcomes"
)

type ShortcodeElement struct {
	Shortcode string `gorm:"column:shortcode"`
}
