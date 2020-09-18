package api

type ResponseLabel string

const (
	Unknown ResponseLabel = "unknown"
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
