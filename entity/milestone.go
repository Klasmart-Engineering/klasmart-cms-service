package entity

const (
	MilestoneStatusDraft     = "draft"
	MilestoneStatusPending   = "pending"
	MilestoneStatusPublished = "published"
	MilestoneStatusRejected  = "rejected"
	MilestoneStatusHidden    = "hidden"
)

type MilestoneStatus string

type Milestone struct {
	ID             string          `gorm:"column:id;primary_key" json:"milestone_id"`
	Name           string          `gorm:"column:name" json:"milestone_name"`
	Shortcode      string          `gorm:"column:shortcode" json:"shortcode"`
	OrganizationID string          `gorm:"column:organization_id" json:"organization_id"`
	AuthorID       string          `gorm:"column:author_id" json:"author_id"`
	Description    string          `gorm:"column:describe" json:"description"`
	LoCounts       int             `gorm:"-" json:"lo_counts"`
	Type           TypeOfMilestone `gorm:"column:type" json:"type"`

	Status       MilestoneStatus `gorm:"column:status" json:"status"`
	RejectReason string          `gorm:"column:reject_reason" json:"reject_reason"`

	LockedBy   string `gorm:"column:locked_by" json:"locked_by"`
	AncestorID string `gorm:"column:ancestor_id" json:"ancestor_id"`
	SourceID   string `gorm:"column:source_id" json:"source_id"`
	LatestID   string `gorm:"column:latest_id" json:"latest_id"`

	CreateAt         int64      `gorm:"column:create_at" json:"created_at"`
	UpdateAt         int64      `gorm:"column:update_at" json:"updated_at"`
	DeleteAt         int64      `gorm:"column:delete_at" json:"deleted_at"`
	Programs         []string   `gorm:"-"`
	Subjects         []string   `gorm:"-"`
	Categories       []string   `gorm:"-"`
	Subcategories    []string   `gorm:"-"`
	Grades           []string   `gorm:"-"`
	Ages             []string   `gorm:"-"`
	Outcomes         []*Outcome `gorm:"-" json:"outcomes"`
	EditingMilestone *Milestone `gorm:"-" json:"-"`
}

func (Milestone) TableName() string {
	return "milestones"
}

func (m Milestone) HasLocked() bool {
	return m.LockedBy != ""
}

func (m Milestone) IsLatest() bool {
	return m.LatestID == m.ID
}

func (m Milestone) IsAncestor() bool {
	return m.ID == m.AncestorID
}

func (m *Milestone) SetStatus(status MilestoneStatus) bool {
	switch status {
	case MilestoneStatusHidden:
		if m.Status == MilestoneStatusPublished {
			m.Status = MilestoneStatusHidden
			return true
		}
	case MilestoneStatusPending:
		if m.Status == MilestoneStatusDraft || m.Status == MilestoneStatusRejected {
			m.Status = MilestoneStatusPending
			return true
		}
	case MilestoneStatusPublished:
		if m.Status == MilestoneStatusPending {
			m.Status = MilestoneStatusPublished
			return true
		}
	case MilestoneStatusRejected:
		if m.Status == MilestoneStatusPending {
			m.Status = MilestoneStatusRejected
			return true
		}
	}
	return false
}

type TypeOfMilestone string

const (
	CustomMilestoneType  TypeOfMilestone = "normal"
	GeneralMilestoneType TypeOfMilestone = "general"
)

const (
	GeneralMilestoneName = "General Milestone"
)

type MilestoneOutcome struct {
	ID              int    `gorm:"column:id;primary_key"`
	MilestoneID     string `gorm:"column:milestone_id" json:"milestone_id"`
	OutcomeAncestor string `gorm:"column:outcome_ancestor" json:"outcome_ancestor"`
	CreateAt        int64  `gorm:"column:create_at" json:"created_at"`
	UpdateAt        int64  `gorm:"column:update_at" json:"updated_at"`
	DeleteAt        int64  `gorm:"column:delete_at" json:"deleted_at"`
}

func (MilestoneOutcome) TableName() string {
	return "milestones_outcomes"
}

type MilestoneCondition struct {
	ID             string   `json:"id" form:"id"`
	IDs            []string `json:"ids" form:"ids"`
	Name           string   `json:"name" form:"name"`
	Description    string   `json:"description" form:"description"`
	Shortcode      string   `json:"shortcode" form:"shortcode"`
	AuthorID       string   `json:"author_id" form:"author_id"`
	AuthorName     string   `json:"author_name" form:"author_name"`
	Page           string   `json:"page" form:"page"`
	PageSize       string   `json:"page_size" form:"page_size"`
	OrderBy        string   `json:"order_by" form:"order_by"`
	Status         string   `json:"status" form:"status"`
	SearchKey      string   `json:"search_key" form:"search_key"`
	AuthorIDs      []string `json:"-" form:"-"`
	OrganizationID string   `json:"-" form:"-"`
}
