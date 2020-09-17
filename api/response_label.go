package api

type ResponseLabel string

const (
	Unknown ResponseLabel = "unknown"
)

type ErrorResponse struct {
	Label ResponseLabel `json:"label"`
}

// L create response object with label
func L(label ResponseLabel) interface{} {
	return map[string]interface{}{
		"label": string(label),
	}
}
