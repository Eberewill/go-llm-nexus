package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/willexm1/go-llm-nexus/internal/core/ports"
)

type PostgresRepository struct {
	conn *pgx.Conn
}

func NewPostgresRepository(connString string) (*PostgresRepository, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	ctx := context.Background()
	// Ensure pgcrypto extension exists before using gen_random_uuid
	if _, err = conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "pgcrypto"`); err != nil {
		return nil, fmt.Errorf("failed to ensure pgcrypto extension: %v", err)
	}

	// Create users table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create users table: %v", err)
	}

	// Create request_logs table with optional user reference
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS request_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NULL REFERENCES users(id),
			prompt TEXT,
			provider TEXT,
			response TEXT,
			duration_ms BIGINT,
			prompt_tokens INT,
			completion_tokens INT,
			total_tokens INT,
			cost_usd NUMERIC(18,6),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return &PostgresRepository{conn: conn}, nil
}

func (r *PostgresRepository) LogRequest(ctx context.Context, log ports.RequestLog) error {
	var userID sql.NullString
	if log.UserID != "" {
		userID = sql.NullString{String: log.UserID, Valid: true}
	}
	_, err := r.conn.Exec(ctx, `
		INSERT INTO request_logs (user_id, prompt, provider, response, duration_ms, prompt_tokens, completion_tokens, total_tokens, cost_usd, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, userID, log.Prompt, log.Provider, log.Response, log.DurationMs, log.PromptTokens, log.CompletionTokens, log.TotalTokens, log.CostUSD, log.CreatedAt)
	return err
}

func (r *PostgresRepository) CreateUser(ctx context.Context, name string) (*ports.User, error) {
	row := r.conn.QueryRow(ctx, `
		INSERT INTO users (name)
		VALUES ($1)
		RETURNING id, created_at
	`, name)
	var id string
	var created time.Time
	if err := row.Scan(&id, &created); err != nil {
		return nil, err
	}
	return &ports.User{ID: id, Name: name, CreatedAt: created}, nil
}

func (r *PostgresRepository) GetUser(ctx context.Context, id string) (*ports.User, error) {
	row := r.conn.QueryRow(ctx, `
		SELECT id, name, created_at FROM users WHERE id = $1
	`, id)
	var user ports.User
	if err := row.Scan(&user.ID, &user.Name, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}
