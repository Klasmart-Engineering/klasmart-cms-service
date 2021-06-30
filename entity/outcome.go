package entity

const (
	OutcomeStatusDraft     = "draft"
	OutcomeStatusPending   = "pending"
	OutcomeStatusPublished = "published"
	OutcomeStatusRejected  = "rejected"
	OutcomeStatusHidden    = "hidden"
)

const (
	JoinComma = ","
)

type OutcomeStatus string
type Outcome struct {
	ID         string `gorm:"type:varchar(50);column:id" dynamodbav:"outcome_id" json:"outcome_id" dynamoupdate:"-"`
	Name       string `gorm:"type:varchar(255);NOT NULL;column:name" dynamodbav:"outcome_name" json:"outcome_name" dynamoupdate:":n"`
	Shortcode  string `gorm:"type:char(8);DEFAULT NULL;column:shortcode" dynamodbav:"shortcode" json:"shortcode" dynamoupdate:":code"`
	AncestorID string `gorm:"type:varchar(50);column:ancestor_id" dynamodbav:"ancestor_id" json:"ancestor_id" dynamoupdate:"-"`
	Program    string `gorm:"type:varchar(1024);NOT NULL;column:program" dynamodbav:"program" json:"program" dynamoupdate:":p"`
	Subject    string `gorm:"type:varchar(1024);NOT NULL;column:subject" dynamodbav:"subject" json:"subject" dynamoupdate:":su"`
	// Category
	//Developmental string `gorm:"type:varchar(1024);NOT NULL;column:developmental" dynamodbav:"developmental" json:"developmental" dynamoupdate:":dv"`
	// SubCategory
	//Skills      string `gorm:"type:varchar(1024);NOT NULL;column:skills" dynamodbav:"skills" json:"skills" dynamoupdate:":sk"`
	Age         string `gorm:"type:varchar(1024);NOT NULL;column:age" dynamodbav:"age" json:"age" dynamoupdate:":a"`
	Grade       string `gorm:"type:varchar(1024);NOT NULL;column:grade" dynamodbav:"grade" json:"grade" dynamoupdate:":grd"`
	Keywords    string `gorm:"type:text;NOT NULL;column:keywords" dynamodbav:"keywords" json:"keywords" dynamoupdate:":ky"`
	Description string `gorm:"type:text;NOT NULL;column:description" dynamodbav:"description" json:"description" dynamoupdate:":de"`

	EstimatedTime  int    `gorm:"type:int;NOT NULL;column:estimated_time" dynamodbav:"estimated_time" json:"extra" dynamoupdate:":est"`
	AuthorID       string `gorm:"type:varchar(50);NOT NULL;column:author_id" dynamodbav:"author_id" json:"author" dynamoupdate:":au"`
	AuthorName     string `gorm:"type:varchar(128);NOT NULL;column:author_name" dynamodbav:"author_name" json:"author_name" dynamoupdate:":aun"`
	OrganizationID string `gorm:"type:varchar(50);NOT NULL;column:organization_id" dynamodbav:"org_id" json:"organization_id" dynamoupdate:":og"`

	PublishScope  string        `gorm:"type:varchar(50);NOT NULL;column:publish_scope;index" dynamodbav:"publish_scope" json:"publish_scope" dynamoupdate:":ps"`
	PublishStatus OutcomeStatus `gorm:"type:varchar(16);NOT NULL;column:publish_status;index" dynamodbav:"publish_status" json:"publish_status" dynamoupdate:":pst"`

	RejectReason string `gorm:"type:varchar(255);NOT NULL;column:reject_reason" dynamodbav:"reject_reason" json:"reject_reason" dynamoupdate:":rr"`
	Version      int    `gorm:"type:int;NOT NULL;column:version" dynamodbav:"version" json:"version" dynamoupdate:":ve"`
	LockedBy     string `gorm:"type:varchar(50);NOT NULL;column:locked_by" dynamodbav:"locked_by" json:"locked_by" dynamoupdate:":lb"`
	SourceID     string `gorm:"type:varchar(255);NOT NULL;column:source_id" dynamodbav:"source_id" json:"source_id" dynamoupdate:":si"`
	LatestID     string `gorm:"type:varchar(255);NOT NULL;column:latest_id" dynamodbav:"latest_id" json:"latest_id" dynamoupdate:":lsi"`
	Assumed      bool   `gorm:"type:tinyint(255);NOT NULL;column:assumed" dynamodbav:"assumed" json:"assumed" dynamoupdate:":asum"`

	CreateAt      int64        `gorm:"type:bigint;NOT NULL;column:create_at" dynamodbav:"created_at" json:"created_at" dynamoupdate:":ca"`
	UpdateAt      int64        `gorm:"type:bigint;NOT NULL;column:update_at" dynamodbav:"updated_at" json:"updated_at" dynamoupdate:":ua"`
	DeleteAt      int64        `gorm:"type:bigint;column:delete_at" dynamodbav:"deleted_at" json:"deleted_at" dynamoupdate:":da"`
	Sets          []*Set       `gorm:"-" json:"sets"`
	Programs      []string     `gorm:"-"`
	Subjects      []string     `gorm:"-"`
	Categories    []string     `gorm:"-"`
	Subcategories []string     `gorm:"-"`
	Grades        []string     `gorm:"-"`
	Ages          []string     `gorm:"-"`
	Milestones    []*Milestone `gorm:"-" json:"milestones"`
}

func (Outcome) TableName() string {
	return "learning_outcomes"
}

func (oc Outcome) GetID() interface{} {
	return oc.ID
}

type OutcomeCondition struct {
	IDs            []string `json:"ids" form:"ids"`
	OutcomeName    string   `json:"outcome_name" form:"outcome_name"`
	Description    string   `json:"description" form:"description"`
	Keywords       string   `json:"keywords" form:"keywords"`
	Shortcode      string   `json:"shortcode" form:"shortcode"`
	AuthorID       string   `json:"-" form:"-"`
	AuthorName     string   `json:"author_name" form:"author_name"`
	Page           int      `json:"page" form:"page"`
	PageSize       int      `json:"page_size" form:"page_size"`
	OrderBy        string   `json:"order_by" form:"order_by"`
	PublishStatus  string   `json:"publish_status" form:"publish_status"`
	FuzzyKey       string   `json:"search_key" form:"search_key"`
	AuthorIDs      []string `json:"-" form:"-"`
	Assumed        int      `json:"assumed" form:"assumed"`
	PublishScope   string   `json:"publish_scope" form:"publish_scope"`
	OrganizationID string   `json:"organization_id" form:"organization_id"`
	SetName        string   `json:"set_name" form:"set_name"`
}
