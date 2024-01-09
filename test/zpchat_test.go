// @Author xiaozhaofu 2023/5/26 19:13:00
package test_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	zhipu "github.com/gtkit/llm-zhipu"
)

func TestZpChat(t *testing.T) {
	key := "111222333.55566633" // 示例:密钥格式

	token, err := zhipu.GenerateToken(key, time.Hour*24)

	if err != nil {
		t.Log("---GenerateToken err:", err)
		return
	}

	prompt := []zhipu.ChatCompletionMessage{
		{Role: zhipu.ChatMessageRoleUser, Content: "你用的什么模型"},
	}

	openConfig := zhipu.DefaultConfig(token)

	openConfig.HTTPClient = &http.Client{
		Timeout: 180 * time.Second,
	}

	// 实例化一个客户端
	c := zhipu.NewClientWithConfig(openConfig)

	req := zhipu.ChatCompletionRequest{
		Model:       zhipu.Turbo,
		Messages:    prompt,
		Temperature: 0.7,
		Incremental: true,
	}
	jr, _ := json.Marshal(req)
	t.Log("---req:", string(jr))

	stream, err := c.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		t.Logf("ChatCompletionStream error: %v\n", err)

		return
	}
	defer stream.Close()

	for {
		response, reserr := stream.Recv()
		// t.Logf("****response: %+v\n", response)
		if errors.Is(reserr, io.EOF) {
			t.Log("\nStream finished")
			return
		}

		if reserr != nil {
			t.Logf("\nStream error: %v\n", reserr)
			return
		}

		var s string
		for _, choice := range response.Choices {
			s += choice.Delta.Content
		}
		t.Log("---- response.Choices Content: ", s)
	}
}
