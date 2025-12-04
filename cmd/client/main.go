package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "github.com/willexm1/go-llm-nexus/api/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	serverAddr := flag.String("addr", "localhost:50051", "The server address in the format of host:port")
	prompt := flag.String("prompt", "Hello, LLM!", "The prompt to send")
	provider := flag.String("provider", "openai", "The provider to use (openai, gemini)")
	flag.Parse()

	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewLLMServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var p pb.Provider
	switch *provider {
	case "openai":
		p = pb.Provider_PROVIDER_OPENAI
	case "gemini":
		p = pb.Provider_PROVIDER_GEMINI
	default:
		p = pb.Provider_PROVIDER_UNSPECIFIED
	}

	r, err := c.ProcessRequest(ctx, &pb.LLMRequest{
		Prompt:      *prompt,
		Provider:    p,
		Temperature: 0.7,
		MaxTokens:   100,
	})
	if err != nil {
		log.Fatalf("could not process request: %v", err)
	}

	log.Printf("Response: %s", r.GetContent())
	log.Printf("Provider Used: %s", r.GetProviderUsed())
	log.Printf("Processing Time: %dms", r.GetProcessingTimeMs())
}
