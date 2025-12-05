package services

import (
	"context"
	"fmt"
	"time"

	"github.com/willexm1/go-llm-nexus/internal/adapters/llm"
	"github.com/willexm1/go-llm-nexus/internal/config"
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
		providers["openai"] = llm.NewOpenAIProvider(llm.OpenAIConfig{
			APIKey:          cfg.LLM.OpenAIKey,
			Model:           cfg.LLM.OpenAIModel,
			InputCostPer1K:  cfg.LLM.OpenAIInputCostPer1K,
			OutputCostPer1K: cfg.LLM.OpenAIOutputCostPer1K,
		})
	}
	if cfg.LLM.GeminiKey != "" {
		providers["gemini"] = llm.NewGeminiProvider(llm.GeminiConfig{
			APIKey:          cfg.LLM.GeminiKey,
			Model:           cfg.LLM.GeminiModel,
			InputCostPer1K:  cfg.LLM.GeminiInputCostPer1K,
			OutputCostPer1K: cfg.LLM.GeminiOutputCostPer1K,
		})
	}

	return &LLMService{
		providers: providers,
		repo:      repo,
		cache:     cache,
	}
}

func (s *LLMService) ProcessRequest(ctx context.Context, req ports.LLMRequest, providerName string) (*ports.LLMResponse, string, error) {
	if err := s.ensureUser(ctx, req.UserID); err != nil {
		return nil, "", err
	}

	// 1. Check Cache (if configured). Incorporate user to avoid cross-user leakage.
	cacheKey := fmt.Sprintf("%s:%s:%s", providerName, req.UserID, req.Prompt)
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			return &ports.LLMResponse{Content: cached}, "cache", nil
		}
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
	if s.cache != nil {
		go func() {
			_ = s.cache.Set(context.Background(), cacheKey, resp.Content, 1*time.Hour)
		}()
	}

	// 5. Log Request (Async)
	if s.repo != nil {
		var promptTokens, completionTokens, totalTokens int32
		var cost float64
		if resp.Usage != nil {
			promptTokens = resp.Usage.PromptTokens
			completionTokens = resp.Usage.CompletionTokens
			totalTokens = resp.Usage.TotalTokens
			cost = resp.Usage.CostUSD
		}
		go func() {
			_ = s.repo.LogRequest(context.Background(), ports.RequestLog{
				Prompt:           req.Prompt,
				Provider:         provider.Name(),
				Response:         resp.Content,
				DurationMs:       duration,
				UserID:           req.UserID,
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				TotalTokens:      totalTokens,
				CostUSD:          cost,
				CreatedAt:        time.Now(),
			})
		}()
	}

	return resp, provider.Name(), nil
}

func (s *LLMService) RegisterUser(ctx context.Context, name string) (*ports.User, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("user storage not configured")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return s.repo.CreateUser(ctx, name)
}

func (s *LLMService) ensureUser(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	if s.repo == nil {
		return fmt.Errorf("user storage not configured")
	}
	_, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}
	return nil
}
