package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"go-nanobot-litellm-lab/internal/contextmgr"
	"go-nanobot-litellm-lab/internal/litellm"
	"go-nanobot-litellm-lab/internal/router"
	"go-nanobot-litellm-lab/internal/tasks"
)

const serviceName = "go-nanobot-litellm-lab"

type Reviewer interface {
	ReviewDiff(ctx context.Context, req litellm.ReviewRequest) (litellm.ReviewResponse, error)
}

type ContextManager interface {
	Build(input contextmgr.Input) (contextmgr.Output, error)
}

type PolicyRouter interface {
	Route(input router.Input) (router.Decision, error)
}

type Options struct {
	Store          *tasks.Store
	Reviewer       Reviewer
	ContextManager ContextManager
	PolicyRouter   PolicyRouter
	RequestTimeout time.Duration
	Logger         *log.Logger
}

type Handler struct {
	store          *tasks.Store
	reviewer       Reviewer
	contextManager ContextManager
	policyRouter   PolicyRouter
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
	PriorPlan   string `json:"prior_plan"`
	Logs        string `json:"logs"`
	Notes       string `json:"notes"`
	BudgetHint  string `json:"budget_hint"`
	Stream      bool   `json:"stream"`
}

type contextReport struct {
	KeptBlocks       []string `json:"kept_blocks,omitempty"`
	CompressedBlocks []string `json:"compressed_blocks,omitempty"`
	DroppedBlocks    []string `json:"dropped_blocks,omitempty"`
}

type taskResponse struct {
	TaskID        string         `json:"task_id"`
	RequestID     string         `json:"request_id,omitempty"`
	Status        string         `json:"status"`
	Result        string         `json:"result,omitempty"`
	Model         string         `json:"model,omitempty"`
	RouteReason   string         `json:"route_reason,omitempty"`
	LatencyMS     int64          `json:"latency_ms"`
	Error         string         `json:"error,omitempty"`
	ContextReport *contextReport `json:"context_report,omitempty"`
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
	if options.ContextManager == nil {
		options.ContextManager = contextmgr.NewManager(contextmgr.Config{})
	}
	if options.PolicyRouter == nil {
		options.PolicyRouter = router.NewDefault()
	}
	if options.Logger == nil {
		options.Logger = log.Default()
	}

	h := &Handler{
		store:          options.Store,
		reviewer:       options.Reviewer,
		contextManager: options.ContextManager,
		policyRouter:   options.PolicyRouter,
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
	contextOut, err := h.contextManager.Build(contextmgr.Input{
		CurrentDiff:     req.Diff,
		RepoSummary:     req.RepoSummary,
		PriorPlan:       req.PriorPlan,
		OldLogs:         req.Logs,
		IrrelevantNotes: req.Notes,
	})
	if err != nil {
		if errors.Is(err, contextmgr.ErrNoUsableContext) {
			writeError(w, http.StatusBadRequest, "no usable context")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid context input")
		return
	}
	report := contextReport{
		KeptBlocks:       contextOut.KeptBlocks,
		CompressedBlocks: contextOut.CompressedBlocks,
		DroppedBlocks:    contextOut.DroppedBlocks,
	}
	decision, err := h.policyRouter.Route(router.Input{
		TaskType:        "review_diff",
		ContextChars:    len(contextOut.FinalContext),
		StreamRequested: req.Stream,
		BudgetHint:      req.BudgetHint,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "no routable model")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	resp, err := h.reviewer.ReviewDiff(ctx, litellm.ReviewRequest{
		Diff:         req.Diff,
		RepoSummary:  req.RepoSummary,
		FinalContext: contextOut.FinalContext,
		ModelAlias:   decision.ModelAlias,
		RequestID:    requestID,
	})
	if err != nil {
		statusCode := mapLiteLLMError(err)
		task := h.store.Save(tasks.Task{
			ID:          taskID,
			RequestID:   requestID,
			Status:      tasks.StatusFailed,
			Model:       decision.ModelAlias,
			RouteReason: decision.RouteReason,
			Error:       err.Error(),
		})
		h.logger.Printf("review_diff failed task_id=%s request_id=%s status=%d model=%s route_reason=%q error=%v", taskID, requestID, statusCode, decision.ModelAlias, decision.RouteReason, err)
		resp := responseFromTask(task)
		resp.ContextReport = &report
		writeJSON(w, statusCode, resp)
		return
	}

	task := h.store.Save(tasks.Task{
		ID:          taskID,
		RequestID:   requestID,
		Status:      tasks.StatusSuccess,
		Result:      resp.Result,
		Model:       decision.ModelAlias,
		RouteReason: decision.RouteReason,
		LatencyMS:   resp.LatencyMS,
	})

	h.logger.Printf("review_diff success task_id=%s request_id=%s model=%s route_reason=%q latency_ms=%d", taskID, requestID, decision.ModelAlias, decision.RouteReason, resp.LatencyMS)
	response := responseFromTask(task)
	response.ContextReport = &report
	writeJSON(w, http.StatusOK, response)
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
		TaskID:      task.ID,
		RequestID:   task.RequestID,
		Status:      string(task.Status),
		Result:      task.Result,
		Model:       task.Model,
		RouteReason: task.RouteReason,
		LatencyMS:   task.LatencyMS,
		Error:       task.Error,
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
