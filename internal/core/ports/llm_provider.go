package ports

import (
	"context"
)

type LLMRequest struct {
	UserID      string
	Prompt      string
	Temperature float32
	MaxTokens   int32
}

type LLMResponse struct {
	Content string
	Usage   *UsageInfo
}

type UsageInfo struct {
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
	CostUSD          float64
}

type LLMProvider interface {
	Generate(ctx context.Context, req LLMRequest) (*LLMResponse, error)
	Name() string
}
