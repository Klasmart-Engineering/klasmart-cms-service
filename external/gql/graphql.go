package gql

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GraphQLError struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Extensions struct {
		Code      string `json:"code"`
		Exception struct {
			Stacktrace []string `json:"stacktrace"`
		} `json:"exception"`
	}
}

func (ge GraphQLError) Error() string {
	str, _ := json.Marshal(ge)
	return string(str)
}

type GraphQLResponse[T ConnectionResponse] struct {
	Data   map[string]T    `json:"data,omitempty"`
	Errors []*GraphQLError `json:"errors,omitempty"`
}
type GraphQLRequest struct {
	query  string
	vars   map[string]interface{}
	Header http.Header
}

type OptionFunc func(*GraphQLRequest)

func RequestToken(token string) OptionFunc {
	return func(req *GraphQLRequest) {
		req.Header.Add(cookieKey, fmt.Sprintf("access=%s", token))
	}
}

func NewRequest(q string, opt ...OptionFunc) *GraphQLRequest {
	req := &GraphQLRequest{
		query:  q,
		Header: make(map[string][]string),
	}

	for i := range opt {
		opt[i](req)
	}
	return req
}

func (req *GraphQLRequest) Var(key string, value interface{}) {
	if req.vars == nil {
		req.vars = make(map[string]interface{})
	}
	req.vars[key] = value
}

func (req *GraphQLRequest) SetHeader(key string, value string) {
	req.Header[key] = []string{value}
}

func (req *GraphQLRequest) SetHeaders(key string, values []string) {
	req.Header[key] = values
}

const cookieKey = "Cookie"
