package api

type ResponseLabel string

func (l ResponseLabel) String() string {
	return string(l)
}

const (
	GeneralUnknown             ResponseLabel = "general_error_unknown"
	GeneralUnAuthorized        ResponseLabel = "general_error_unauthorized"
	GeneralUnAuthorizedNoOrgID ResponseLabel = "general_error_no_organization"

	// Assessment
	AssessMsgOneStudent   ResponseLabel = "assess_msg_one_student"
	AssessMsgNoPermission ResponseLabel = "assess_msg_no_permission"

	// Report
	ReportMsgNoPermission ResponseLabel = "report_error_no_permissions"

	//Library
	LibraryMsgContentLocked      ResponseLabel = "library_error_content_locked"
	LibraryMsgContentDataInvalid ResponseLabel = "library_error_content_data_invalid"
	LibraryContentLockedByMe     ResponseLabel = "library_error_content_locked_by_me"

	//Folder
	FolderDeleteNoEmptyFolder ResponseLabel = "library_error_delete_folder"

	// schedule
	ScheduleMessageEditOverlap           ResponseLabel = "schedule_msg_edit_overlap"
	ScheduleMessageDeleteOverlap         ResponseLabel = "schedule_msg_delete_overlap"
	ScheduleMessageOverlap               ResponseLabel = "schedule_msg_overlap"
	ScheduleMessageNoPermission          ResponseLabel = "schedule_msg_no_permission"
	ScheduleMessageDeleteMissTime        ResponseLabel = "schedule_msg_delete_minutes"
	ScheduleMessageEditMissTime          ResponseLabel = "schedule_msg_edit_minutes"
	ScheduleMessageGoLiveTimeNotUp       ResponseLabel = "schedule_msg_start_minutes"
	ScheduleMessageTimeExpired           ResponseLabel = "schedule_msg_time_expired"
	ScheduleMessageLessonPlanInvalid     ResponseLabel = "schedule_msg_recall_lesson_plan"
	ScheduleMessageDueDateEarlierEndDate ResponseLabel = "schedule_msg_due_date_earlier"
	ScheduleMessageDueDateEarlierToDay   ResponseLabel = "schedule_msg_earlier_today"
	ScheduleMessageUsersConflict         ResponseLabel = "schedule_msg_users_conflict"
	ScheduleMsgDeleteMissDueDate         ResponseLabel = "schedule_msg_delete_due_date"
	ScheduleMsgEditMissDueDate           ResponseLabel = "schedule_msg_edit_due_date"
	ScheduleMsgHidden                    ResponseLabel = "schedule_msg_hidden"
	scheduleMsgHide                      ResponseLabel = "schedule_msg_hide"
	ScheduleMsgAssignmentNew             ResponseLabel = "schedule_msg_assignment_new"
	ScheduleFeedbackCompleted            ResponseLabel = "schedule_msg_cannot_submit"
)

// L create response object with label
func L(label ResponseLabel) interface{} {
	return Response{Label: label.String()}
}

// D create response object with data
func D(data interface{}) interface{} {
	return Response{Data: data}
}

// LD create response object with label and data
func LD(label ResponseLabel, data interface{}) interface{} {
	return Response{Label: label.String(), Data: data}
}

type Response struct {
	Label string      `json:"label,omitempty" example:"unknown"`
	Data  interface{} `json:"data"`
}
type SuccessRequestResponse Response

type BadRequestResponse Response

type ForbiddenResponse Response

type NotFoundResponse Response

type InternalServerErrorResponse Response

type ConflictResponse Response

type UnAuthorizedResponse Response

type IDResponse struct {
	ID string `json:"id"`
}
