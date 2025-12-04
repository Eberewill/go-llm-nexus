package services

import (
	"context"
	"fmt"
	"time"

	"github.com/willexm1/go-llm-nexus/internal/config"
	"github.com/willexm1/go-llm-nexus/internal/adapters/llm"
	"github.com/willexm1/go-llm-nexus/internal/core/ports"
)

type LLMService struct {
	providers map[string]ports.LLMProvider
	repo      ports.Repository
	cache     ports.Cache
}

func NewLLMService(cfg *config.Config, repo ports.Repository, cache ports.Cache) *LLMService {
	providers := make(map[string]ports.LLMProvider)

	if cfg.LLM.OpenAIKey != "" {
		providers["openai"] = llm.NewOpenAIProvider(cfg.LLM.OpenAIKey)
	}
	if cfg.LLM.GeminiKey != "" {
		providers["gemini"] = llm.NewGeminiProvider(cfg.LLM.GeminiKey)
	}

	return &LLMService{
		providers: providers,
		repo:      repo,
		cache:     cache,
	}
}

func (s *LLMService) ProcessRequest(ctx context.Context, req ports.LLMRequest, providerName string) (*ports.LLMResponse, string, error) {
	// 1. Check Cache
	cacheKey := fmt.Sprintf("%s:%s", providerName, req.Prompt)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		return &ports.LLMResponse{Content: cached}, "cache", nil
	}

	// 2. Select Provider
	var provider ports.LLMProvider
	var ok bool

	if providerName != "" {
		provider, ok = s.providers[providerName]
		if !ok {
			return nil, "", fmt.Errorf("provider %s not configured", providerName)
		}
	} else {
		// Simple fallback strategy: try OpenAI, then Gemini, then HF
		if p, exists := s.providers["openai"]; exists {
			provider = p
		} else if p, exists := s.providers["gemini"]; exists {
			provider = p
		} else {
			return nil, "", fmt.Errorf("no llm providers configured")
		}
	}

	// 3. Call Provider
	start := time.Now()
	resp, err := provider.Generate(ctx, req)
	if err != nil {
		return nil, provider.Name(), err
	}
	duration := time.Since(start).Milliseconds()

	// 4. Cache Response (Async)
	go func() {
		_ = s.cache.Set(context.Background(), cacheKey, resp.Content, 1*time.Hour)
	}()

	// 5. Log Request (Async)
	go func() {
		_ = s.repo.LogRequest(context.Background(), ports.RequestLog{
			Prompt:     req.Prompt,
			Provider:   provider.Name(),
			Response:   resp.Content,
			DurationMs: duration,
			CreatedAt:  time.Now(),
		})
	}()

	return resp, provider.Name(), nil
}
