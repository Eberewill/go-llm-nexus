package main

import (
	"fmt"
	"log"
	"net/http"

	myHttp "github.com/willexm1/go-llm-nexus/internal/adapters/handler/http"
	"github.com/willexm1/go-llm-nexus/internal/adapters/repository"
	"github.com/willexm1/go-llm-nexus/internal/config"
	"github.com/willexm1/go-llm-nexus/internal/core/ports"
	"github.com/willexm1/go-llm-nexus/internal/core/services"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Infrastructure
	// Database (optional)
	var repo ports.Repository
	if cfg.Database.Host != "" {
		dbConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
		dbRepo, err := repository.NewPostgresRepository(dbConnStr)
		if err != nil {
			log.Fatalf("Failed to connect to database (required for user registration): %v", err)
		}
		repo = dbRepo
	} else {
		log.Fatalf("Database configuration is required to store users")
	}

	// Redis (optional)
	var cache ports.Cache
	if cfg.Redis.Addr != "" {
		cache = repository.NewRedisCache(cfg.Redis.Addr, cfg.Redis.Password)
	}

	// 3. Initialize Services
	llmService := services.NewLLMService(cfg, repo, cache)

	// 4. HTTP Server only
	httpHandler := myHttp.NewHandler(llmService)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/generate", httpHandler.Generate)
	mux.HandleFunc("/api/health", httpHandler.Health)
	mux.HandleFunc("/api/users", httpHandler.RegisterUser)

	httpAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("HTTP Server listening on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatalf("failed to serve http: %v", err)
	}
}
