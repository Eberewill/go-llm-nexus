package ports

import (
	"context"
	"time"
)

type RequestLog struct {
	ID           string
	Prompt       string
	Provider     string
	Response     string
	DurationMs   int64
	CreatedAt    time.Time
}

type Repository interface {
	LogRequest(ctx context.Context, log RequestLog) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
}
