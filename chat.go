package zhipu

import (
	"context"
	"net/http"
	"time"
)

// Chat message role defined by the OpenAI API.
const (
	ChatMessageRoleSystem    = "system"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
)

const chatCompletionsSuffix = "/invoke"
const chatStreamCompletionsSuffix = "/sse-invoke"

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest  请求模型参数.
type ChatCompletionRequest struct {
	Model       string                  `json:"model"`  // 模型
	Messages    []ChatCompletionMessage `json:"prompt"` // prompt
	Temperature float32                 `json:"temperature,omitempty"`
	TopP        float32                 `json:"top_p,omitempty"`
	// 智谱 SSE接口调用时，用于控制每次返回内容方式是增量还是全量，不提供此参数时默认为增量返回 - true 为增量返回 - false 为全量返回
	Incremental bool `json:"incremental"`
}

// ChatglmCompletionResponse Api文本返回.
type ChatglmCompletionResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		RequestID  string                  `json:"request_id"`
		TaskID     string                  `json:"task_id"`
		TaskStatus string                  `json:"task_status"`
		Choices    []ChatCompletionMessage `json:"choices"`
		Usage      struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	} `json:"data"`
	Success bool `json:"success"`
}

type ChatCompletionChoice struct {
	Message ChatCompletionMessage `json:"message"`
}

// ChatCompletionResponse represents a response structure for chat completion API.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object,omitempty"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage,omitempty"`
}

// CreateChatCompletion — API call to Create a completion for the chat message.
func (c *Client) CreateChatCompletion(
	ctx context.Context,
	request ChatCompletionRequest,
) (response ChatCompletionResponse, err error) {
	urlSuffix := chatCompletionsSuffix

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix, request.Model), withBody(request))
	if err != nil {
		return
	}
	var glm ChatglmCompletionResponse

	if err = c.sendRequest(req, &glm); err != nil {
		return
	}

	return ChatCompletionResponse{
		ID:      glm.Data.TaskID,
		Object:  glm.Msg,
		Created: time.Now().Unix(),
		Model:   request.Model,
		Choices: []ChatCompletionChoice{
			{
				Message: glm.Data.Choices[0],
			},
		},
		Usage: Usage{
			TotalTokens: glm.Data.Usage.TotalTokens,
		},
	}, nil
}
