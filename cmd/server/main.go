package main

import (
	"fmt"
	"log"
	"net"

	"net/http"

	"github.com/willexm1/go-llm-nexus/internal/config"
	"github.com/willexm1/go-llm-nexus/internal/core/services"
	grpcHandler "github.com/willexm1/go-llm-nexus/internal/adapters/handler/grpc"
	myHttp "github.com/willexm1/go-llm-nexus/internal/adapters/handler/http"
	"github.com/willexm1/go-llm-nexus/internal/adapters/repository"
	pb "github.com/willexm1/go-llm-nexus/api/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Infrastructure
	// Database
	dbConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", 
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	repo, err := repository.NewPostgresRepository(dbConnStr)
	if err != nil {
		log.Printf("Failed to connect to database: %v. Continuing without DB persistence.", err)
		// In production, we might want to fail hard, but for demo we can continue or use a mock
	}

	// Redis
	cache := repository.NewRedisCache(cfg.Redis.Addr, cfg.Redis.Password)

	// 3. Initialize Services
	llmService := services.NewLLMService(cfg, repo, cache)

	// 4. Initialize Servers
	// HTTP Server (in goroutine)
	go func() {
		httpHandler := myHttp.NewHandler(llmService)
		mux := http.NewServeMux()
		mux.HandleFunc("/api/generate", httpHandler.Generate)
		mux.HandleFunc("/api/health", httpHandler.Health)
		
		log.Printf("HTTP Server listening on port 8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatalf("failed to serve http: %v", err)
		}
	}()

	// gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	
	// Register Handler
	handler := grpcHandler.NewServer(llmService)
	pb.RegisterLLMServiceServer(s, handler)

	// Enable reflection for grpcurl
	reflection.Register(s)

	log.Printf("gRPC Server listening on port %s", cfg.Server.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
