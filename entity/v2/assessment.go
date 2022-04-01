package v2

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type AssessmentQueryReq struct {
	QueryKey       string              `form:"query_key"`
	QueryType      AssessmentQueryType `form:"query_type"`
	AssessmentType AssessmentType      `form:"assessment_type"`
	OrderBy        string              `form:"order_by"`
	Status         string              `form:"status"`
	PageIndex      int                 `form:"page"`
	PageSize       int                 `form:"page_size"`
}
type AssessmentStatusForApiCompliant string

const (
	AssessmentStatusCompliantNotCompleted AssessmentStatusForApiCompliant = "in_progress"
	AssessmentStatusCompliantCompleted    AssessmentStatusForApiCompliant = "complete"
)

func (a AssessmentStatusForApiCompliant) String() string {
	return string(a)
}

type AssessmentPageReply struct {
	Total       int                     `json:"total"`
	Assessments []*AssessmentQueryReply `json:"assessments"`
}

type ListAssessmentsResultForHomePage struct {
	Total int                          `json:"total"`
	Items []*AssessmentItemForHomePage `json:"items"`
}
type AssessmentItemForHomePage struct {
	ID       string                          `json:"id"`
	Title    string                          `json:"title"`
	Teachers []*entity.IDName                `json:"teachers"`
	Status   AssessmentStatusForApiCompliant `json:"status"`
}

type AssessmentQueryReply struct {
	// all type
	ID     string           `json:"id"`
	Title  string           `json:"title"`
	Status AssessmentStatus `json:"status"`

	// onlineClass,offlineClass,OnlineStudy
	LessonPlan *entity.IDName `json:"lesson_plan"`

	// onlineClass,offlineClass,OnlineStudy,ReviewStudy
	Teachers []*entity.IDName `json:"teachers"`

	// onlineClass,offlineClass,OnlineStudy
	CompleteAt int64 `json:"complete_at"`

	// onlineClass,offlineClass
	Program    *entity.IDName   `json:"program"`
	Subjects   []*entity.IDName `json:"subjects"`
	ClassEndAt int64            `json:"class_end_at"`

	// OnlineStudy,ReviewStudy
	ClassInfo    *entity.IDName `json:"class_info"`
	DueAt        int64          `json:"due_at"`
	CompleteRate float64        `json:"complete_rate"`

	// OnlineStudy
	RemainingTime int64 `json:"remaining_time"`
}

type AssessmentAddWhenCreateSchedulesReq struct {
	RepeatScheduleIDs []string
	Users             []*AssessmentUserReq
	AssessmentType    AssessmentType
	//LessPlanID           string
	ClassRosterClassName string
	ScheduleTitle        string
}
type AssessmentUserReq struct {
	UserID   string
	UserType AssessmentUserType
}

func (req *AssessmentAddWhenCreateSchedulesReq) Valid(ctx context.Context) bool {
	if len(req.RepeatScheduleIDs) <= 0 || !req.AssessmentType.Valid(ctx) {
		return false
	}
	return true
}

type AssessmentAttendancesReq struct {
	AttendanceID   string
	AttendanceType AssessmentUserType
}

type ScheduleEndClassCallBackReq struct {
	ScheduleID    string   `json:"schedule_id"`
	AttendanceIDs []string `json:"attendance_ids"`
	ClassLength   int      `json:"class_length"`
	ClassEndAt    int64    `json:"class_end_time"`
}

// Valid implement jwt Claims interface
func (a *ScheduleEndClassCallBackReq) Valid() error {
	return nil
}

type ScheduleEndClassCallBackResp struct {
	ScheduleID string `json:"schedule_id"`
}

type AssessmentDetailReply struct {
	ID                string           `json:"id"`
	Title             string           `json:"title"`
	Status            AssessmentStatus `json:"status"`
	RoomID            string           `json:"room_id"`
	Class             *entity.IDName   `json:"class"`
	Teachers          []*entity.IDName `json:"teachers"`
	Program           *entity.IDName   `json:"program"`
	Subjects          []*entity.IDName `json:"subjects"`
	ClassEndAt        int64            `json:"class_end_at"`
	ClassLength       int              `json:"class_length"`
	RemainingTime     int64            `json:"remaining_time"`
	CompleteAt        int64            `json:"complete_at"`
	ScheduleTitle     string           `json:"schedule_title"`
	ScheduleDueAt     int64            `json:"schedule_due_at"`
	CompleteRate      float64          `json:"complete_rate"`
	IsAnyOneAttempted bool             `json:"is_anyone_attempted"`

	Outcomes []*AssessmentOutcomeReply `json:"outcomes"`
	Contents []*AssessmentContentReply `json:"contents"`
	Students []*AssessmentStudentReply `json:"students"`

	DiffContentStudents []*AssessmentDiffContentStudentsReply `json:"diff_content_students,omitempty"`
}

type AssessmentDiffContentStudentsReply struct {
	StudentID       string                           `json:"student_id"`
	StudentName     string                           `json:"student_name"`
	Status          AssessmentUserStatus             `json:"status" enums:"Participate,NotParticipate"`
	ReviewerComment string                           `json:"reviewer_comment"`
	Results         []*DiffContentStudentResultReply `json:"results,omitempty"`
}

type DiffContentStudentResultReply struct {
	Answer    string                     `json:"answer"`
	Score     float64                    `json:"score"`
	Attempted bool                       `json:"attempted"`
	Content   AssessmentDiffContentReply `json:"content"`
}

type AssessmentDiffContentReply struct {
	Number               string                 `json:"number"`
	ParentID             string                 `json:"parent_id"`
	ContentID            string                 `json:"content_id"`
	H5PID                string                 `json:"h5p_id"`
	ContentName          string                 `json:"content_name"`
	ContentType          AssessmentContentType  `json:"content_type" enums:"LessonPlan,LessonMaterial,Unknown"`
	ContentSubtype       string                 `json:"content_subtype"`
	FileType             AssessmentFileArchived `json:"file_type"  enums:"Unknown,HasChildContainer,NotChildContainer,SupportScoreStandAlone,NotSupportScoreStandAlone"`
	MaxScore             float64                `json:"max_score"`
	H5PSubID             string                 `json:"h5p_sub_id"`
	RoomProvideContentID string                 `json:"-"`
}

type AssessmentStudentReply struct {
	StudentID          string                          `json:"student_id"`
	StudentName        string                          `json:"student_name"`
	Status             AssessmentUserStatus            `json:"status" enums:"Participate,NotParticipate"`
	ReviewerComment    string                          `json:"reviewer_comment"`
	Results            []*AssessmentStudentResultReply `json:"results,omitempty"`
	OfflineStudyResult *StudentOfflineStudyResult      `json:"offline_study_result,omitempty"`
}

type StudentOfflineStudyResult struct {
	AssessScore      AssessmentUserAssess                   `json:"assess_score" enums:"1,2,3,4,5"`
	Outcomes         []*AssessmentStudentResultOutcomeReply `json:"outcomes"`
	StudentFeedBacks []*entity.ScheduleFeedbackView         `json:"student_feed_backs"`
}

type AssessmentStudentResultReply struct {
	Answer    string                                 `json:"answer"`
	Score     float64                                `json:"score"`
	Attempted bool                                   `json:"attempted"`
	ContentID string                                 `json:"content_id"`
	Outcomes  []*AssessmentStudentResultOutcomeReply `json:"outcomes"`
}

type AssessmentStudentResultOutcomeReply struct {
	OutcomeID string                      `json:"outcome_id"`
	Status    AssessmentUserOutcomeStatus `json:"status"  enums:"Unknown,NotCovered,NotAchieved,Achieved"`
}

type AssessmentContentReply struct {
	Number               string                  `json:"number"`
	ParentID             string                  `json:"parent_id"`
	ContentID            string                  `json:"content_id"`
	H5PID                string                  `json:"h5p_id"`
	ContentName          string                  `json:"content_name"`
	ReviewerComment      string                  `json:"reviewer_comment"`
	Status               AssessmentContentStatus `json:"status"   enums:"Covered,NotCovered"`
	OutcomeIDs           []string                `json:"outcome_ids"`
	ContentType          AssessmentContentType   `json:"content_type" enums:"LessonPlan,LessonMaterial,Unknown"`
	ContentSubtype       string                  `json:"content_subtype"`
	FileType             AssessmentFileArchived  `json:"file_type"  enums:"Unknown,HasChildContainer,NotChildContainer,SupportScoreStandAlone,NotSupportScoreStandAlone"`
	MaxScore             float64                 `json:"max_score"`
	H5PSubID             string                  `json:"h5p_sub_id"`
	RoomProvideContentID string                  `json:"-"`
	IgnoreCalculateScore bool                    `json:"-"`
	//LatestID string `json:"latest_id"`
}

type AssessmentOutcomeReply struct {
	OutcomeID          string                        `json:"outcome_id"`
	OutcomeName        string                        `json:"outcome_name"`
	AssignedTo         []AssessmentOutcomeAssignType `json:"assigned_to" enums:"LessonPlan,LessonMaterial"`
	Assumed            bool                          `json:"assumed"`
	AssignedToLessPlan bool                          `json:"-"`
	AssignedToMaterial bool                          `json:"-"`
	ScoreThreshold     float32                       `json:"score_threshold"`
	//StudentIDs  []string                      `json:"student_ids"`
}

type AssessmentUpdateReq struct {
	ID       string                        `json:"id"`
	Action   AssessmentAction              `json:"action"  enums:"Draft,Complete"`
	Students []*AssessmentStudentUpdateReq `json:"students"`
	Contents []*AssessmentUpdateContentReq `json:"contents"`
}

type AssessmentStudentUpdateReq struct {
	StudentID             string                        `json:"student_id"`
	Status                AssessmentUserStatus          `json:"status"  enums:"Participate,NotParticipate"`
	ReviewerComment       string                        `json:"reviewer_comment"`
	Results               []*AssessmentStudentResultReq `json:"results"`
	OfflineStudyResultReq *OfflineStudyResultReq        `json:"offline_study_result"`
}

type OfflineStudyResultReq struct {
	Outcomes         []*AssessmentStudentResultOutcomeReq
	AssessFeedbackID string                    `json:"assess_feedback_id"`
	AssessScore      AssessmentUserAssess      `json:"assess_score" enums:"1,2,3,4,5"`
	Assignments      []*FeedbackAssignmentsReq `json:"assignments"`
}

type FeedbackAssignmentsReq struct {
	ID                   string `json:"id"`
	ReviewerAssignmentID string `json:"reviewer_assignment_id"`
}

type AssessmentStudentResultReq struct {
	ParentID  string                               `json:"parent_id"`
	ContentID string                               `json:"content_id"`
	Outcomes  []*AssessmentStudentResultOutcomeReq `json:"outcomes"`
	Score     float64                              `json:"score"`
}

type AssessmentStudentResultOutcomeReq struct {
	OutcomeID string                      `json:"outcome_id"`
	Status    AssessmentUserOutcomeStatus `json:"status"  enums:"Unknown,NotCovered,NotAchieved,Achieved"`
}

type AssessmentUpdateContentReq struct {
	ParentID        string                  `json:"parent_id"`
	ContentID       string                  `json:"content_id"`
	ReviewerComment string                  `json:"reviewer_comment"`
	Status          AssessmentContentStatus `json:"status"   enums:"Covered,NotCovered"`
}
type AssessmentOutComeAssigned struct {
	AssignedToLessPlan bool
	AssignedToMaterial bool
}

type OfflineStudyUserPageReply struct {
	Total int64                       `json:"total"`
	Item  []*OfflineStudyUserPageItem `json:"item"`
}
type OfflineStudyUserPageItem struct {
	ID          string                  `json:"id"`
	Title       string                  `json:"title"`
	Teachers    []*entity.IDName        `json:"teachers"`
	Student     *entity.IDName          `json:"student"`
	CompleteAt  int64                   `json:"complete_at"`
	Status      UserResultProcessStatus `json:"status" enums:"Started,Draft,Complete"`
	DueAt       int64                   `json:"due_at"`
	SubmitAt    int64                   `json:"submit_at"`
	AssessScore AssessmentUserAssess    `json:"assess_score"`
}

type OfflineStudyUserResultAddReq struct {
	ScheduleID string
	UserID     string
	FeedbackID string
}

type GetOfflineStudyUserResultDetailReply struct {
	ID            string                          `json:"id"`
	ScheduleID    string                          `json:"schedule_id"`
	Title         string                          `json:"title"`
	Teachers      []*entity.IDName                `json:"teachers"`
	Student       *entity.IDName                  `json:"student"`
	Status        UserResultProcessStatus         `json:"status" enums:"InProgress,Draft,Complete"`
	DueAt         int64                           `json:"due_at"`
	CompleteAt    int64                           `json:"complete_at"`
	FeedbackID    string                          `json:"feedback_id"`
	AssessScore   AssessmentUserAssess            `json:"assess_score" enums:"1,2,3,4,5"`
	AssessComment string                          `json:"assess_comment"`
	Outcomes      []*OfflineStudyUserOutcomeReply `json:"outcomes"`
}

type OfflineStudyUserOutcomeReply struct {
	OutcomeID   string                      `json:"outcome_id"`
	OutcomeName string                      `json:"outcome_name"`
	Assumed     bool                        `json:"assumed"`
	Status      AssessmentUserOutcomeStatus `json:"status" enums:"Unknown,NotCovered,NotAchieved,Achieved"`
}

type OfflineStudyUserResultUpdateReq struct {
	ID               string                              `json:"id"`
	AssessFeedbackID string                              `json:"assess_feedback_id"`
	AssessScore      AssessmentUserAssess                `json:"assess_score" enums:"1,2,3,4,5"`
	AssessComment    string                              `json:"assess_comment"`
	Action           AssessmentAction                    `json:"action"  enums:"Draft,Complete"`
	Outcomes         []*OfflineStudyUserOutcomeUpdateReq `json:"outcomes"`
}
type OfflineStudyUserOutcomeUpdateReq struct {
	OutcomeID string                      `json:"outcome_id"`
	Status    AssessmentUserOutcomeStatus `json:"status" enums:"Unknown,NotCovered,NotAchieved,Achieved"`
}

type AssessmentsSummary struct {
	Complete   int `json:"complete"`
	InProgress int `json:"in_progress"`
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

	ClassType AssessmentTypeCompliant `form:"type"`

	OrderBy  string `form:"order_by"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type AssessmentTypeCompliant string

const (
	AssessmentTypeCompliantOfflineClass AssessmentTypeCompliant = "class"
	AssessmentTypeCompliantOnlineClass  AssessmentTypeCompliant = "live"
	AssessmentTypeCompliantOnlineStudy  AssessmentTypeCompliant = "study"
	AssessmentTypeCompliantOfflineStudy AssessmentTypeCompliant = "home_fun_study"
)

func (a AssessmentTypeCompliant) ToAssessmentType(ctx context.Context) (AssessmentType, error) {
	switch a {
	case AssessmentTypeCompliantOfflineClass:
		return AssessmentTypeOfflineClass, nil
	case AssessmentTypeCompliantOnlineClass:
		return AssessmentTypeOnlineClass, nil
	case AssessmentTypeCompliantOnlineStudy:
		return AssessmentTypeOnlineStudy, nil
	case AssessmentTypeCompliantOfflineStudy:
		return AssessmentTypeOfflineStudy, nil
	default:
		log.Warn(ctx, "not support assessment type", log.Any("AssessmentTypeCompliant", a))
		return "", constant.ErrInvalidArgs
	}
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

type SearchStudentAssessmentsResponse struct {
	List  []*StudentAssessment `json:"list"`
	Total int64                `json:"total"`
}

type AssessmentAnyoneAttemptedReply struct {
	IsAnyoneAttempted bool
	AssessmentStatus  AssessmentStatus
}

type AssessmentContentView struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	ContentType AssessmentContentType `json:"content_type"`
	OutcomeIDs  []string              `json:"outcome_ids"`
	LatestID    string                `json:"latest_id"`
	FileType    entity.FileType       `json:"file_type"`
}

type StatisticsCountReq struct {
	Status string `json:"status"`
}

type UpdateAssessmentUserOutput struct {
	AssessmentUserMap map[string]*AssessmentUser
	WaitUpdatedUsers  []*AssessmentUser
	AssessmentUserPKs []string
}

type PrepareAssessmentContentUpdateDataOutput struct {
	OldToLatestContentIDMap map[string]string
	WaitAddContentsMap      map[string]*AssessmentContent
	WaitAddContents         []*AssessmentContent
	ContentReqMap           map[string]*AssessmentUpdateContentReq
}
