package zhipu

import (
	"net/http"
)

const (
	glmaiAPIURLv1 = "https://open.bigmodel.cn/api/paas/v3/model-api/"
)

// ClientConfig is a configuration of a client.
type ClientConfig struct {
	authToken  string
	BaseURL    string
	HTTPClient *http.Client
}

func DefaultConfig(authToken string) ClientConfig {
	return ClientConfig{
		authToken:  authToken,
		BaseURL:    glmaiAPIURLv1,
		HTTPClient: &http.Client{},
	}
}

func (ClientConfig) String() string {
	return "<GlmAI API ClientConfig>"
}
