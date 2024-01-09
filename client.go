package zhipu

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gtkit/llm-zhipu/utils"
)

type ChatCompletion[T ChatCompletionRequest] interface {
	CreateChatCompletionStream(ctx context.Context, request T) (stream *GlmChatCompletionStream, err error)
	CreateChatCompletion(ctx context.Context, request T) (response ChatCompletionResponse, err error)
	newRequest(ctx context.Context, method, url string, setters ...requestOption) (*http.Request, error)
	sendRequest(req *http.Request, v any) error
	setCommonHeaders(req *http.Request)
	fullURL(suffix string, args ...any) string
	handleErrorResp(resp *http.Response) error
}

type Client struct {
	config         ClientConfig
	requestBuilder utils.RequestBuilder
}

type requestOptions struct {
	body   any
	header http.Header
}

type requestOption func(*requestOptions)

var _ ChatCompletion[ChatCompletionRequest] = (*Client)(nil)

func NewClient(authToken string) ChatCompletion[ChatCompletionRequest] {
	config := DefaultConfig(authToken)
	return NewClientWithConfig(config)
}

func NewClientWithConfig(config ClientConfig) ChatCompletion[ChatCompletionRequest] {
	return &Client{
		config:         config,
		requestBuilder: utils.NewRequestBuilder(),
	}
}

func (c *Client) newRequest(ctx context.Context, method, url string, setters ...requestOption) (*http.Request, error) {
	args := &requestOptions{
		body:   nil,
		header: make(http.Header),
	}
	for _, setter := range setters {
		setter(args)
	}

	req, err := c.requestBuilder.Build(ctx, method, url, args.body, args.header)
	if err != nil {
		return nil, err
	}
	c.setCommonHeaders(req)
	return req, nil
}

func (c *Client) sendRequest(req *http.Request, v any) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")

	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	res, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if isFailureStatusCode(res) {
		return c.handleErrorResp(res)
	}

	return decodeResponse(res.Body, v)
}

func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("Authorization", c.config.authToken)
}

// fullURL returns full URL for request.
// args[0] is model name.
func (c *Client) fullURL(suffix string, args ...any) string {
	deploymentName := ""
	if len(args) > 0 {
		model, ok := args[0].(string)
		if ok {
			deploymentName = model
		}
	}

	return fmt.Sprintf("%s%s%s", c.config.BaseURL, deploymentName, suffix)
}

func (c *Client) handleErrorResp(resp *http.Response) error {
	var errRes ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&errRes)
	if err != nil || errRes.Error == nil {
		reqErr := &RequestError{
			HTTPStatusCode: resp.StatusCode,
			Err:            err,
		}
		if errRes.Error != nil {
			reqErr.Err = errRes.Error
		}
		return reqErr
	}

	errRes.Error.HTTPStatusCode = resp.StatusCode
	return errRes.Error
}

func sendRequestStream[T streamable](client *Client, req *http.Request) (*streamReader[T], error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.config.HTTPClient.Do(req) //nolint:bodyclose // body is closed in stream.Close()
	if err != nil {
		return new(streamReader[T]), err
	}
	if isFailureStatusCode(resp) {
		return new(streamReader[T]), client.handleErrorResp(resp)
	}

	return &streamReader[T]{
		reader:         bufio.NewReader(resp.Body),
		response:       resp,
		errAccumulator: utils.NewErrorAccumulator(),
		unmarshaler:    &utils.JSONUnmarshaler{},
	}, nil
}

func withBody(body any) requestOption {
	return func(args *requestOptions) {
		args.body = body
	}
}

func isFailureStatusCode(resp *http.Response) bool {
	return resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest
}

func decodeResponse(body io.Reader, v any) error {
	if v == nil {
		return nil
	}

	if result, ok := v.(*string); ok {
		return decodeString(body, result)
	}
	return json.NewDecoder(body).Decode(v)
}

func decodeString(body io.Reader, output *string) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	*output = string(b)
	return nil
}
