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
	UserID      string  `json:"user_id"`
	Prompt      string  `json:"prompt"`
	Provider    string  `json:"provider"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int32   `json:"max_tokens"`
}

type GenerateResponse struct {
	Content          string        `json:"content"`
	ProviderUsed     string        `json:"provider_used"`
	ProcessingTimeMs int64         `json:"processing_time_ms"`
	Usage            *UsagePayload `json:"usage,omitempty"`
}

type UsagePayload struct {
	PromptTokens     int32   `json:"prompt_tokens"`
	CompletionTokens int32   `json:"completion_tokens"`
	TotalTokens      int32   `json:"total_tokens"`
	CostUSD          float64 `json:"cost_usd"`
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

	if req.UserID == "" {
		log.Printf("[HTTP] Missing user identifier")
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	log.Printf("[HTTP] Received request - Provider: %s, User: %s, Prompt: %.50s...", req.Provider, req.UserID, req.Prompt)

	start := time.Now()
	coreReq := ports.LLMRequest{
		UserID:      req.UserID,
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
		Usage:            convertUsage(resp.Usage),
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
		"status":  "healthy",
		"service": "go-llm-nexus",
	})
}

type registerUserRequest struct {
	Name string `json:"name"`
}

type registerUserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		log.Printf("[HTTP] Method not allowed for register user: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[HTTP] Failed to decode register user request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	user, err := h.service.RegisterUser(r.Context(), req.Name)
	if err != nil {
		log.Printf("[HTTP] Failed to register user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(registerUserResponse{
		ID:        user.ID,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	})
}

func convertUsage(u *ports.UsageInfo) *UsagePayload {
	if u == nil {
		return nil
	}
	return &UsagePayload{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
		CostUSD:          u.CostUSD,
	}
}
