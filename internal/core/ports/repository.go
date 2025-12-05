package ports

import (
	"context"
	"time"
)

type RequestLog struct {
	ID               string
	Prompt           string
	Provider         string
	Response         string
	DurationMs       int64
	UserID           string
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
	CostUSD          float64
	CreatedAt        time.Time
}

type User struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type Repository interface {
	LogRequest(ctx context.Context, log RequestLog) error
	CreateUser(ctx context.Context, name string) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
}
