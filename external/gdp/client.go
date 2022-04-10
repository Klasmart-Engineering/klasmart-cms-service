package gdp

import (
	"bytes"
	"context"
	"encoding/json"
	newrelic "github.com/newrelic/go-agent"
	"gitlab.badanamu.com.cn/calmisland/common-cn/helper"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"io/ioutil"
	"net/http"
	"time"
)

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

func GraphQLRun[ResType ConnectionResponse](ctx context.Context, c *GraphGLClient, req *GraphQLRequest, resp *GraphQLResponse[ResType]) (int, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.httpTimeout)
	defer cancel()

	reqBody := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     req.q,
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
	if bada, ok := helper.GetBadaCtx(ctxWithTimeout); ok {
		bada.SetHeader(request.Header)
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
