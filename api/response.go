package api

type BadRequestResponse struct {
	Label string `json:"label" example:"unknown"`
}

type NotFoundResponse struct{}

type InternalServerErrorResponse struct{}

type ResponseLabel string

const (
	Unknown ResponseLabel = "unknown"
)

// L create response object with label
func L(label ResponseLabel) interface{} {
	return BadRequestResponse{Label: string(label)}
}
