package grpc

import (
	"context"
	"time"

	pb "github.com/willexm1/go-llm-nexus/api/proto/v1"
	"github.com/willexm1/go-llm-nexus/internal/core/ports"
	"github.com/willexm1/go-llm-nexus/internal/core/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedLLMServiceServer
	service *services.LLMService
}

func NewServer(service *services.LLMService) *Server {
	return &Server{
		service: service,
	}
}

func (s *Server) ProcessRequest(ctx context.Context, req *pb.LLMRequest) (*pb.LLMResponse, error) {
	start := time.Now()

	// Map proto provider enum to string
	var providerName string
	switch req.Provider {
	case pb.Provider_PROVIDER_OPENAI:
		providerName = "openai"
	case pb.Provider_PROVIDER_GEMINI:
		providerName = "gemini"
	default:
		providerName = "" // Let service decide
	}

	coreReq := ports.LLMRequest{
		Prompt:      req.Prompt,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	resp, providerUsed, err := s.service.ProcessRequest(ctx, coreReq, providerName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process request: %v", err)
	}

	return &pb.LLMResponse{
		Content:          resp.Content,
		ProviderUsed:     providerUsed,
		ProcessingTimeMs: time.Since(start).Milliseconds(),
	}, nil
}
