package zhipu

import (
	"context"
	"net/http"
)

type ChatCompletionStreamChoiceDelta struct {
	Content string `json:"content"`
	Role    string `json:"role,omitempty"`
}

type ChatCompletionStreamChoice struct {
	Delta ChatCompletionStreamChoiceDelta `json:"delta"`
}

type GlmChatCompletionStreamResponse struct {
	ID      string                       `json:"id"`
	Event   string                       `json:"event"`
	Choices []ChatCompletionStreamChoice `json:"choices"`
	Meta    GlmMeta                      `json:"meta"`
}

type GlmChatCompletionStream struct {
	*streamReader[GlmChatCompletionStreamResponse]
}
type GlmMeta struct {
	TaskStatus string `json:"task_status"`
	Usage      struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	TaskID    string `json:"task_id"`
	RequestID string `json:"request_id"`
}

func (c *Client) CreateChatCompletionStream(
	ctx context.Context,
	request ChatCompletionRequest,
) (*GlmChatCompletionStream, error) {
	urlSuffix := chatStreamCompletionsSuffix

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix, request.Model), withBody(request))
	if err != nil {
		return nil, err
	}

	resp, err := sendRequestStream[GlmChatCompletionStreamResponse](c, req)
	if err != nil {
		return nil, err
	}
	return &GlmChatCompletionStream{
		streamReader: resp,
	}, nil
}
