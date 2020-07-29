package entity

type Response struct {
	StatusCode int `json:"status_code"`
	StatusMsg interface{} `json:"status_msg"`
}

type ErrMsg struct {
	ErrMsg string `json:"err_msg"`
}

func NewErrorResponse(statusCode int, msg string) *Response {
	return &Response{
		StatusCode: statusCode,
		StatusMsg:  ErrMsg{ErrMsg: msg},
	}
}

type IdMsg struct {
	Id string `json:"id"`
}
