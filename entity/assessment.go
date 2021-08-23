package entity

const (
	AssessmentTypeClass        AssessmentType = "class"
	AssessmentTypeLive         AssessmentType = "live"
	AssessmentTypeStudy        AssessmentType = "study"
	AssessmentTypeHomeFunStudy AssessmentType = "home_fun_study"
)

type AssessmentScheduleType struct {
	ClassType ScheduleClassType
	IsHomeFun bool
}

type AssessmentType string

func (a AssessmentType) ToScheduleClassType() AssessmentScheduleType {
	switch a {
	case AssessmentTypeClass:
		return AssessmentScheduleType{ClassType: ScheduleClassTypeOfflineClass, IsHomeFun: false}
	case AssessmentTypeLive:
		return AssessmentScheduleType{ClassType: ScheduleClassTypeOnlineClass, IsHomeFun: false}
	case AssessmentTypeStudy:
		return AssessmentScheduleType{ClassType: ScheduleClassTypeHomework, IsHomeFun: false}
	case AssessmentTypeHomeFunStudy:
		return AssessmentScheduleType{ClassType: ScheduleClassTypeHomework, IsHomeFun: true}
	}
	return AssessmentScheduleType{ClassType: ScheduleClassTypeOnlineClass, IsHomeFun: false}
}

func (a AssessmentType) Valid() bool {
	switch a {
	case AssessmentTypeClass:
		return true
	case AssessmentTypeLive:
		return true
	case AssessmentTypeStudy:
		return true
	case AssessmentTypeHomeFunStudy:
		return true
	}
	return false
}

func (a AssessmentType) String() string {
	return string(a)
}

type NullAssessmentTypes struct {
	Value []AssessmentType
	Valid bool
}

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

type AssessmentStudentViewH5PLessonMaterial struct {
	ParentID             string                  `json:"parent_id"`
	H5PID                string                  `json:"h5p_id"`     // add: 2021.06.24
	SubH5PID             string                  `json:"sub_h5p_id"` // add: 2021.06.24
	Number               string                  `json:"number"`
	LessonMaterialID     string                  `json:"lesson_material_id"`
	LessonMaterialName   string                  `json:"lesson_material_name"`
	LessonMaterialType   string                  `json:"lesson_material_type"`
	Answer               string                  `json:"answer"`
	MaxScore             float64                 `json:"max_score"`
	AchievedScore        float64                 `json:"achieved_score"`
	Attempted            bool                    `json:"attempted"`
	IsH5P                bool                    `json:"is_h5p"`
	Outcomes             []*AssessmentIDNamePair `json:"outcomes"`
	NotApplicableScoring bool                    `json:"not_applicable_scoring"`
	HasSubItems          bool                    `json:"has_sub_items"`
	// internal
	LessonMaterialOrderedNumber int                                       `json:"lesson_material_ordered_number"`
	OrderedID                   int                                       `json:"ordered_id"`
	Children                    []*AssessmentStudentViewH5PLessonMaterial `json:"children"`
}

type UpdateAssessmentH5PStudent struct {
	StudentID       string                               `json:"student_id"`
	Comment         string                               `json:"comment"`
	LessonMaterials []*UpdateAssessmentH5PLessonMaterial `json:"lesson_materials"`
}

type UpdateAssessmentH5PLessonMaterial struct {
	LessonMaterialID string  `json:"lesson_material_id"`
	AchievedScore    float64 `json:"achieved_score"`
	H5PID            string  `json:"h5p_id"`     // add: 2021.06.24
	SubH5PID         string  `json:"sub_h5p_id"` // add: 2021.06.24
}

type AddAssessmentArgs struct {
	Title         string              `json:"title"`
	ScheduleID    string              `json:"schedule_id"`
	ScheduleTitle string              `json:"schedule_title"`
	LessonPlanID  string              `json:"lesson_plan_id"`
	ClassID       string              `json:"class_id"`
	ClassLength   int                 `json:"class_length"`
	ClassEndTime  int64               `json:"class_end_time"`
	Attendances   []*ScheduleRelation `json:"attendances"`
}

type BatchAddAssessmentSuperArgs struct {
	Raw                       []*AddAssessmentArgs
	ScheduleIDs               []string
	Outcomes                  []*Outcome
	OutcomeMap                map[string]*Outcome
	LessonPlanMap             map[string]*AssessmentExternalLessonPlan
	ScheduleIDToOutcomeIDsMap map[string][]string
}

type StudentAssessmentTeacher struct {
	Teacher *StudentAssessmentTeacherInfo `json:"teacher"`
	Comment string                        `json:"comment"`
}
type StudentAssessmentTeacherInfo struct {
	ID         string `json:"id"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Avatar     string `json:"avatar"`
}

type StudentAssessmentSchedule struct {
	ID         string                       `json:"id"`
	Title      string                       `json:"title"`
	Type       string                       `json:"type"`
	Attachment *StudentAssessmentAttachment `json:"attachment,omitempty"`
}
type StudentAssessmentAttachment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type StudentAssessment struct {
	ID                  string                        `json:"id"`
	Title               string                        `json:"title"`
	Score               int                           `json:"score"`
	Status              string                        `json:"status"`
	CreateAt            int64                         `json:"create_at"`
	UpdateAt            int64                         `json:"update_at"`
	CompleteAt          int64                         `json:"complete_at"`
	TeacherComments     []*StudentAssessmentTeacher   `json:"teacher_comments"`
	Schedule            *StudentAssessmentSchedule    `json:"schedule"`
	FeedbackAttachments []StudentAssessmentAttachment `json:"student_attachments"`

	CompleteBy string   `json:"-"`
	ScheduleID string   `json:"-"`
	FeedbackID string   `json:"-"`
	StudentID  string   `json:"-"`
	TeacherIDs []string `json:"-"`
	Comment    string   `json:"-"`

	IsHomeFun bool `json:"-"`
}

type StudentCollectRelatedIDs struct {
	ScheduleIDs      []string
	AllAssessmentIDs []string
	FeedbackIDs      []string
	AssessmentsIDs   []string
}

type StudentQueryAssessmentConditions struct {
	ID          string   `form:"assessment_id"`
	OrgID       string   `form:"org_id"`
	StudentID   string   `form:"student_id"`
	TeacherID   string   `form:"teacher_id"`
	ScheduleID  string   `form:"schedule_id"`
	ScheduleIDs []string `form:"schedule_ids"`
	Status      string   `form:"status"`

	CreatedStartAt int64 `form:"create_at_ge"`
	CreatedEndAt   int64 `form:"create_at_le"`

	UpdateStartAt   int64 `form:"update_at_ge"`
	UpdateEndAt     int64 `form:"update_at_le"`
	CompleteStartAt int64 `form:"complete_at_ge"`
	CompleteEndAt   int64 `form:"complete_at_le"`

	ClassType string `form:"type"`

	OrderBy  string `form:"order_by"`
	Page     string `form:"page"`
	PageSize string `form:"page_size"`
}

type SearchStudentAssessmentsResponse struct {
	List  []*StudentAssessment `json:"list"`
	Total int                  `json:"total"`
}

type NullTimeRange struct {
	StartAt int64
	EndAt   int64
	Valid   bool
}

type H5PRoomComment struct {
	Comment    string `json:"comment"`
	TeacherID  string `json:"teacher_id"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}

type UnifiedAssessment struct {
	ID           string           `json:"id"`
	ScheduleID   string           `json:"schedule_id"`
	Title        string           `json:"title"`
	CompleteTime int64            `json:"complete_time"`
	Status       AssessmentStatus `json:"status"`
	CreateAt     int64            `json:"create_at"`
	UpdateAt     int64            `json:"update_at"`
	DeleteAt     int64            `json:"delete_at"`
}

type QueryUnifiedAssessmentArgs struct {
	Types           NullAssessmentTypes  `json:"types"`
	Status          NullAssessmentStatus `json:"status"`
	OrgID           NullString           `json:"org_id"`
	IDs             NullStrings          `json:"ids"`
	ScheduleIDs     NullStrings          `json:"schedule_ids"`
	CompleteBetween NullTimeRange        `json:"complete_between"`
}

type AssessmentContentType string

const (
	AssessmentContentTypeLessonPlan     AssessmentContentType = "lesson_plan"
	AssessmentContentTypeLessonMaterial AssessmentContentType = "lesson_material"
)

type AssessmentIDNamePair struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
