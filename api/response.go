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

	// schedule
	ScheduleMsgEditOverlap           ResponseLabel = "schedule_msg_edit_overlap"
	ScheduleMsgDeleteOverlap         ResponseLabel = "schedule_msg_delete_overlap"
	ScheduleMsgOverlap               ResponseLabel = "schedule_msg_overlap"
	ScheduleMsgNoPermission          ResponseLabel = "schedule_msg_no_permission"
	ScheduleMsgDeleteMissTime        ResponseLabel = "schedule_msg_delete_minutes"
	ScheduleMsgEditMissTime          ResponseLabel = "schedule_msg_edit_minutes"
	ScheduleMsgGoLiveTimeNotUp       ResponseLabel = "schedule_msg_start_minutes"
	ScheduleMsgTimeExpired           ResponseLabel = "schedule_msg_time_expired"
	ScheduleMsgLessonPlanInvalid     ResponseLabel = "schedule_msg_recall_lesson_plan"
	ScheduleMsgDueDateEarlierEndDate ResponseLabel = "schedule_msg_due_date_earlier"
	ScheduleMsgDueDateEarlierToDay   ResponseLabel = "schedule_msg_earlier_today"
)

// L create response object with label
func L(label ResponseLabel) interface{} {
	return ErrorResponse{Label: label.String()}
}

// LD create response object with label and data
func LD(lable ResponseLabel, data interface{}) interface{} {
	return ErrorResponse{Label: lable.String(), Data: data}
}

type ErrorResponse struct {
	Label string      `json:"label,omitempty" example:"unknown"`
	Data  interface{} `json:"data"`
}

type BadRequestResponse ErrorResponse

type ForbiddenResponse ErrorResponse

type NotFoundResponse ErrorResponse

type InternalServerErrorResponse ErrorResponse

type ConflictResponse ErrorResponse

type UnAuthorizedResponse ErrorResponse
