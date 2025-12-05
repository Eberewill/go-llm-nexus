package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/willexm1/go-llm-nexus/internal/core/ports"
)

type OpenAIProvider struct {
	apiKey          string
	model           string
	inputCostPer1K  float64
	outputCostPer1K float64
	client          *http.Client
}

type OpenAIConfig struct {
	APIKey          string
	Model           string
	InputCostPer1K  float64
	OutputCostPer1K float64
}

func NewOpenAIProvider(cfg OpenAIConfig) *OpenAIProvider {
	model := cfg.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	return &OpenAIProvider{
		apiKey:          cfg.APIKey,
		model:           model,
		inputCostPer1K:  cfg.InputCostPer1K,
		outputCostPer1K: cfg.OutputCostPer1K,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

type openAIRequest struct {
	Model       string  `json:"model"`
	Messages    []msg   `json:"messages"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int32   `json:"max_tokens"`
}

type msg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage openAIUsage `json:"usage"`
}

type openAIUsage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
}

func (p *OpenAIProvider) Generate(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	requestBody := openAIRequest{
		Model:       p.model,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Messages: []msg{
			{Role: "user", Content: req.Prompt},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai api error: %s - %s", resp.Status, string(bodyBytes))
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from openai")
	}

	usage := &ports.UsageInfo{
		PromptTokens:     openAIResp.Usage.PromptTokens,
		CompletionTokens: openAIResp.Usage.CompletionTokens,
		TotalTokens:      openAIResp.Usage.TotalTokens,
	}
	usage.CostUSD = p.calculateCost(usage.PromptTokens, usage.CompletionTokens)

	return &ports.LLMResponse{
		Content: openAIResp.Choices[0].Message.Content,
		Usage:   usage,
	}, nil
}

func (p *OpenAIProvider) calculateCost(promptTokens, completionTokens int32) float64 {
	cost := (float64(promptTokens) / 1000.0 * p.inputCostPer1K) + (float64(completionTokens) / 1000.0 * p.outputCostPer1K)
	return cost
}
