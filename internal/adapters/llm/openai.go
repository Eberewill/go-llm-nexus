package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/willexm1/go-llm-nexus/internal/core/ports"
)

type OpenAIProvider struct {
	apiKey string
	client *http.Client
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey: apiKey,
		client: &http.Client{},
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
}

func (p *OpenAIProvider) Generate(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	requestBody := openAIRequest{
		Model:       "gpt-3.5-turbo", // Or configurable
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
		return nil, fmt.Errorf("openai api error: %s", resp.Status)
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from openai")
	}

	return &ports.LLMResponse{
		Content: openAIResp.Choices[0].Message.Content,
	}, nil
}
