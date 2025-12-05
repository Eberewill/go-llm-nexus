package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/willexm1/go-llm-nexus/internal/core/ports"
	"google.golang.org/genai"
)

type GeminiProvider struct {
	apiKey          string
	model           string
	inputCostPer1K  float64
	outputCostPer1K float64
	mu              sync.Mutex
	client          *genai.Client
}

type GeminiConfig struct {
	APIKey          string
	Model           string
	InputCostPer1K  float64
	OutputCostPer1K float64
}

func NewGeminiProvider(cfg GeminiConfig) *GeminiProvider {
	model := cfg.Model
	if model == "" {
		model = "gemini-2.0-flash-exp"
	}
	return &GeminiProvider{
		apiKey:          cfg.APIKey,
		model:           model,
		inputCostPer1K:  cfg.InputCostPer1K,
		outputCostPer1K: cfg.OutputCostPer1K,
	}
}

func (p *GeminiProvider) Name() string {
	return "Gemini"
}

func (p *GeminiProvider) Generate(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	client, err := p.clientForRequests()
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	maxTokens := req.MaxTokens
	result, err := client.Models.GenerateContent(
		ctx,
		p.model,
		genai.Text(req.Prompt),
		&genai.GenerateContentConfig{
			Temperature:     &req.Temperature,
			MaxOutputTokens: maxTokens,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %w", err)
	}

	usage := &ports.UsageInfo{}
	if result.UsageMetadata != nil {
		usage.PromptTokens = result.UsageMetadata.PromptTokenCount
		usage.CompletionTokens = result.UsageMetadata.CandidatesTokenCount
		usage.TotalTokens = result.UsageMetadata.TotalTokenCount
	}
	usage.CostUSD = p.calculateCost(usage.PromptTokens, usage.CompletionTokens)

	return &ports.LLMResponse{
		Content: result.Text(),
		Usage:   usage,
	}, nil
}

func (p *GeminiProvider) clientForRequests() (*genai.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.client != nil {
		return p.client, nil
	}
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: p.apiKey,
	})
	if err != nil {
		return nil, err
	}
	p.client = client
	return client, nil
}

func (p *GeminiProvider) calculateCost(promptTokens, completionTokens int32) float64 {
	cost := (float64(promptTokens) / 1000.0 * p.inputCostPer1K) + (float64(completionTokens) / 1000.0 * p.outputCostPer1K)
	return cost
}
