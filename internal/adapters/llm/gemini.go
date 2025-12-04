package llm

import (
	"context"
	"fmt"

	"github.com/willexm1/go-llm-nexus/internal/core/ports"
	"google.golang.org/genai"
)

type GeminiProvider struct {
	apiKey string
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
	}
}

func (p *GeminiProvider) Name() string {
	return "Gemini"
}

func (p *GeminiProvider) Generate(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: p.apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	maxTokens := req.MaxTokens
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash-exp",
		genai.Text(req.Prompt),
		&genai.GenerateContentConfig{
			Temperature:     &req.Temperature,
			MaxOutputTokens: maxTokens,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %w", err)
	}

	return &ports.LLMResponse{
		Content: result.Text(),
	}, nil
}
