package ports

import (
	"context"
)

type LLMRequest struct {
	Prompt      string
	Temperature float32
	MaxTokens   int32
}

type LLMResponse struct {
	Content string
}

type LLMProvider interface {
	Generate(ctx context.Context, req LLMRequest) (*LLMResponse, error)
	Name() string
}
