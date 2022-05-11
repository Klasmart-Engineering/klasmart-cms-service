package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/tracecontext"
	newrelic "github.com/newrelic/go-agent"
)

type GraphQLError []struct {
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
	Data   map[string]T `json:"data,omitempty"`
	Errors GraphQLError `json:"errors,omitempty"`
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

type GraphGLClient struct {
	endpoint    string
	httpClient  *http.Client
	httpTimeout time.Duration
}

type OptionClient func(c *GraphGLClient)

func WithTimeout(duration time.Duration) OptionClient {
	return func(c *GraphGLClient) {
		c.httpTimeout = duration
	}
}

type debugTransport struct{}

func (d debugTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	ctx := req.Context()
	txnExist := newrelic.FromContext(ctx) != nil
	log.Debug(ctx, "chlorine costume round trip",
		log.Any("headers", req.Header),
		log.Bool("txn exist", txnExist))
	return http.DefaultTransport.RoundTrip(req)
}

const defaultHttpTimeout = time.Minute

func NewClient(endpoint string, options ...OptionClient) *GraphGLClient {
	c := &GraphGLClient{
		endpoint: endpoint,
		// New Relic will look for txn in the request context if txn is nil,
		// and using default transport if original transport is nil. So args: (nil, nil) is ok
		httpClient:  &http.Client{Transport: newrelic.NewRoundTripper(nil, debugTransport{})},
		httpTimeout: defaultHttpTimeout,
	}
	for i := range options {
		options[i](c)
	}
	return c
}

func (c *GraphGLClient) Run(ctx context.Context, req *GraphQLRequest, resp interface{}) (int, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.httpTimeout)
	defer cancel()

	reqBody := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     req.query,
		Variables: req.vars,
	}
	reqBuffer, err := json.Marshal(&reqBody)
	if err != nil {
		log.Warn(ctxWithTimeout, "Run: marshalFilter failed", log.Err(err), log.Any("reqBody", reqBody))
		return 0, err
	}
	request, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodPost, c.endpoint, bytes.NewBuffer(reqBuffer))
	if err != nil {
		log.Warn(ctxWithTimeout, "Run: New httpRequest failed", log.Err(err), log.Any("reqBody", reqBody))
		return 0, err
	}
	if traceContext, ok := tracecontext.GetTraceContext(ctxWithTimeout); ok {
		traceContext.SetHeader(request.Header)
	}
	request.Header = req.Header
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Accept", "application; charset=utf-8")
	var result *http.Response
	var resultErr error
	start := time.Now()
	result, resultErr = c.httpClient.Do(request)

	duration := time.Since(start)
	if resultErr != nil {
		log.Error(ctxWithTimeout, "Run: do http failed",
			log.Duration("duration", duration),
			log.Err(resultErr),
			log.String("endpoint", c.endpoint),
			log.Any("reqBody", reqBody))
		return 0, resultErr
	}

	defer result.Body.Close()
	response, err := ioutil.ReadAll(result.Body)
	if err != nil {
		log.Error(ctxWithTimeout, "Run: read response failed",
			log.Duration("duration", duration),
			log.Err(err), log.String("endpoint", c.endpoint),
			log.Any("reqBody", reqBody), log.String("response", string(response)))
		return result.StatusCode, err
	}

	err = json.Unmarshal(response, resp)
	if err != nil {
		log.Error(ctxWithTimeout, "Run: unmarshal response failed",
			log.Duration("duration", duration),
			log.Err(err), log.String("endpoint", c.endpoint),
			log.Any("reqBody", reqBody), log.String("response", string(response)))
		return result.StatusCode, err
	}
	log.Debug(ctxWithTimeout, "Run: Success",
		log.Duration("duration", duration),
		log.Any("reqBody", reqBody),
		log.String("response", string(response)))
	return result.StatusCode, nil
}
