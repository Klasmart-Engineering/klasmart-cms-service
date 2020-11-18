package api

type ResponseLabel string

const (
	GeneralUnknown             ResponseLabel = "general_error_unknown"
	GeneralUnAuthorized        ResponseLabel = "general_error_unauthorized"
	GeneralUnAuthorizedNoOrgID ResponseLabel = "general_error_no_organization"
)

const (
	AssessMsgOneStudent ResponseLabel = "assess_msg_one_student"
)

// schedule msg
const (
	ScheduleMsgEditOverlap   ResponseLabel = "schedule_schedule_msg_edit_all"
	ScheduleMsgDeleteOverlap ResponseLabel = "schedule_msg_delete_overlap"
	ScheduleMsgOverlap       ResponseLabel = "schedule_msg_overlap"
	ScheduleMsgNoPermission  ResponseLabel = "schedule_msg_no_permission"
)

const (
	AssessMsgNoPermission  ResponseLabel = "assess_error_no_permissions"
)

// L create response object with label
func L(label ResponseLabel) interface{} {
	return ErrorResponse{Label: string(label)}
}

type ErrorResponse struct {
	Label string `json:"label,omitempty" example:"unknown" enums:"unknown"`
}

type BadRequestResponse ErrorResponse

type ForbiddenResponse ErrorResponse

type NotFoundResponse ErrorResponse

type InternalServerErrorResponse ErrorResponse

type ConflictResponse ErrorResponse
