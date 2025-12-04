package http

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/willexm1/go-llm-nexus/internal/core/ports"
	"github.com/willexm1/go-llm-nexus/internal/core/services"
)

type Handler struct {
	service *services.LLMService
}

func NewHandler(service *services.LLMService) *Handler {
	return &Handler{service: service}
}

type GenerateRequest struct {
	Prompt      string  `json:"prompt"`
	Provider    string  `json:"provider"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int32   `json:"max_tokens"`
}

type GenerateResponse struct {
	Content          string `json:"content"`
	ProviderUsed     string `json:"provider_used"`
	ProcessingTimeMs int64  `json:"processing_time_ms"`
}

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		log.Printf("[HTTP] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[HTTP] Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[HTTP] Received request - Provider: %s, Prompt: %.50s...", req.Provider, req.Prompt)

	start := time.Now()
	coreReq := ports.LLMRequest{
		Prompt:      req.Prompt,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	resp, providerUsed, err := h.service.ProcessRequest(r.Context(), coreReq, req.Provider)
	if err != nil {
		log.Printf("[HTTP] Error processing request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	duration := time.Since(start)
	log.Printf("[HTTP] Request completed - Provider: %s, Duration: %v", providerUsed, duration)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateResponse{
		Content:          resp.Content,
		ProviderUsed:     providerUsed,
		ProcessingTimeMs: duration.Milliseconds(),
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "go-llm-nexus",
	})
}
