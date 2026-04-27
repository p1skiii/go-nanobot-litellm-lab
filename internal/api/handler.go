package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"go-nanobot-litellm-lab/internal/litellm"
	"go-nanobot-litellm-lab/internal/tasks"
)

const serviceName = "go-nanobot-litellm-lab"

type Reviewer interface {
	ReviewDiff(ctx context.Context, req litellm.ReviewRequest) (litellm.ReviewResponse, error)
}

type Options struct {
	Store          *tasks.Store
	Reviewer       Reviewer
	RequestTimeout time.Duration
	Logger         *log.Logger
}

type Handler struct {
	store          *tasks.Store
	reviewer       Reviewer
	requestTimeout time.Duration
	logger         *log.Logger
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type reviewDiffRequest struct {
	Diff        string `json:"diff"`
	RepoSummary string `json:"repo_summary"`
	Stream      bool   `json:"stream"`
}

type taskResponse struct {
	TaskID    string `json:"task_id"`
	RequestID string `json:"request_id,omitempty"`
	Status    string `json:"status"`
	Result    string `json:"result,omitempty"`
	Model     string `json:"model,omitempty"`
	LatencyMS int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

func NewHandler(opts ...Options) http.Handler {
	options := Options{}
	if len(opts) > 0 {
		options = opts[0]
	}
	if options.Store == nil {
		options.Store = tasks.NewStore()
	}
	if options.RequestTimeout <= 0 {
		options.RequestTimeout = 30 * time.Second
	}
	if options.Logger == nil {
		options.Logger = log.Default()
	}

	h := &Handler{
		store:          options.Store,
		reviewer:       options.Reviewer,
		requestTimeout: options.RequestTimeout,
		logger:         options.Logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", method(http.MethodGet, health))
	mux.HandleFunc("/tasks/review-diff", method(http.MethodPost, h.reviewDiff))
	mux.HandleFunc("/tasks/", method(http.MethodGet, h.getTask))
	return mux
}

func method(allowed string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowed {
			w.Header().Set("Allow", allowed)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		next(w, r)
	}
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:  "ok",
		Service: serviceName,
	})
}

func (h *Handler) reviewDiff(w http.ResponseWriter, r *http.Request) {
	var req reviewDiffRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if strings.TrimSpace(req.Diff) == "" {
		writeError(w, http.StatusBadRequest, "diff is required")
		return
	}

	if req.Stream {
		writeError(w, http.StatusBadRequest, "streaming is not supported in M1")
		return
	}

	if h.reviewer == nil {
		writeError(w, http.StatusBadGateway, "litellm reviewer is not configured")
		return
	}

	taskID := tasks.NewID("task")
	requestID := requestIDFrom(r)

	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	resp, err := h.reviewer.ReviewDiff(ctx, litellm.ReviewRequest{
		Diff:        req.Diff,
		RepoSummary: req.RepoSummary,
		RequestID:   requestID,
	})
	if err != nil {
		statusCode := mapLiteLLMError(err)
		task := h.store.Save(tasks.Task{
			ID:        taskID,
			RequestID: requestID,
			Status:    tasks.StatusFailed,
			Error:     err.Error(),
		})
		h.logger.Printf("review_diff failed task_id=%s request_id=%s status=%d error=%v", taskID, requestID, statusCode, err)
		writeJSON(w, statusCode, responseFromTask(task))
		return
	}

	task := h.store.Save(tasks.Task{
		ID:        taskID,
		RequestID: requestID,
		Status:    tasks.StatusSuccess,
		Result:    resp.Result,
		Model:     resp.Model,
		LatencyMS: resp.LatencyMS,
	})

	h.logger.Printf("review_diff success task_id=%s request_id=%s model=%s latency_ms=%d", taskID, requestID, resp.Model, resp.LatencyMS)
	writeJSON(w, http.StatusOK, responseFromTask(task))
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	task, ok := h.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	writeJSON(w, http.StatusOK, responseFromTask(task))
}

func requestIDFrom(r *http.Request) string {
	if requestID := strings.TrimSpace(r.Header.Get("X-Request-ID")); requestID != "" {
		return requestID
	}
	return tasks.NewID("req")
}

func mapLiteLLMError(err error) int {
	var llmErr *litellm.Error
	if errors.As(err, &llmErr) && llmErr.Kind == litellm.KindTimeout {
		return http.StatusGatewayTimeout
	}
	return http.StatusBadGateway
}

func responseFromTask(task tasks.Task) taskResponse {
	return taskResponse{
		TaskID:    task.ID,
		RequestID: task.RequestID,
		Status:    string(task.Status),
		Result:    task.Result,
		Model:     task.Model,
		LatencyMS: task.LatencyMS,
		Error:     task.Error,
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}
