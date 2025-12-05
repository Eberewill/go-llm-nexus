package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/willexm1/go-llm-nexus/internal/config"
	"github.com/willexm1/go-llm-nexus/internal/core/ports"
)

// Mocks
type mockProvider struct {
	name string
}

func (m *mockProvider) Generate(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	return &ports.LLMResponse{
		Content: "mock response from " + m.name,
		Usage: &ports.UsageInfo{
			PromptTokens:     5,
			CompletionTokens: 10,
			TotalTokens:      15,
			CostUSD:          0.001,
		},
	}, nil
}
func (m *mockProvider) Name() string { return m.name }

type mockRepo struct {
	users map[string]*ports.User
}

func (m *mockRepo) LogRequest(ctx context.Context, log ports.RequestLog) error { return nil }
func (m *mockRepo) CreateUser(ctx context.Context, name string) (*ports.User, error) {
	if m.users == nil {
		m.users = make(map[string]*ports.User)
	}
	user := &ports.User{ID: "user-" + name, Name: name, CreatedAt: time.Now()}
	m.users[user.ID] = user
	return user, nil
}
func (m *mockRepo) GetUser(ctx context.Context, id string) (*ports.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

type mockCache struct {
	data map[string]string
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", nil
}
func (m *mockCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func TestLLMService_ProcessRequest(t *testing.T) {
	repo := &mockRepo{users: map[string]*ports.User{
		"user-123": {ID: "user-123", Name: "Test", CreatedAt: time.Now()},
	}}
	cache := &mockCache{data: make(map[string]string)}
	cfg := &config.Config{}
	svc := NewLLMService(cfg, repo, cache)
	svc.providers = map[string]ports.LLMProvider{
		"mock": &mockProvider{name: "mock"},
	}

	ctx := context.Background()
	req := ports.LLMRequest{UserID: "user-123", Prompt: "Hello"}

	resp, provider, err := svc.ProcessRequest(ctx, req, "mock")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "mock" {
		t.Fatalf("expected provider 'mock', got %s", provider)
	}
	if resp.Content == "" {
		t.Fatalf("expected content in response")
	}
	if resp.Usage == nil || resp.Usage.TotalTokens == 0 {
		t.Fatalf("expected usage data in response")
	}
}
