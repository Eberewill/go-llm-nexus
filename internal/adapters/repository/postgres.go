package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/willexm1/llm-backend-showcase/internal/core/ports"
)

type PostgresRepository struct {
	conn *pgx.Conn
}

func NewPostgresRepository(connString string) (*PostgresRepository, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	
	// Create table if not exists (Simple migration for demo)
	_, err = conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS request_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			prompt TEXT,
			provider TEXT,
			response TEXT,
			duration_ms BIGINT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return &PostgresRepository{conn: conn}, nil
}

func (r *PostgresRepository) LogRequest(ctx context.Context, log ports.RequestLog) error {
	_, err := r.conn.Exec(ctx, `
		INSERT INTO request_logs (prompt, provider, response, duration_ms, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, log.Prompt, log.Provider, log.Response, log.DurationMs, log.CreatedAt)
	return err
}
