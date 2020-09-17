package api

type ErrorResponse struct {
	Label string `json:"label" example:"unknown"`
}
type ResponseLabel string

const (
	Unknown ResponseLabel = "unknown"
)

// L create response object with label
func L(label ResponseLabel) interface{} {
	return ErrorResponse{Label: string(label)}
}
