package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-nanobot-litellm-lab/internal/contextmgr"
	"go-nanobot-litellm-lab/internal/invocation"
	"go-nanobot-litellm-lab/internal/litellm"
	"go-nanobot-litellm-lab/internal/router"
	"go-nanobot-litellm-lab/internal/tasks"
	"go-nanobot-litellm-lab/internal/usage"
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
	Store            *tasks.Store
	Reviewer         Reviewer
	ContextManager   ContextManager
	PolicyRouter     PolicyRouter
	InvocationLedger invocation.Ledger
	RequestTimeout   time.Duration
	Logger           *log.Logger
}

type Handler struct {
	store            *tasks.Store
	reviewer         Reviewer
	contextManager   ContextManager
	policyRouter     PolicyRouter
	invocationLedger invocation.Ledger
	requestTimeout   time.Duration
	logger           *log.Logger
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
	RunID         string         `json:"run_id,omitempty"`
	AttemptID     string         `json:"attempt_id,omitempty"`
	RequestID     string         `json:"request_id,omitempty"`
	Status        string         `json:"status"`
	Result        string         `json:"result,omitempty"`
	Model         string         `json:"model,omitempty"`
	RouteReason   string         `json:"route_reason,omitempty"`
	LatencyMS     int64          `json:"latency_ms"`
	Error         string         `json:"error,omitempty"`
	ContextReport *contextReport `json:"context_report,omitempty"`
}

type usageListResponse struct {
	TaskID  string         `json:"task_id,omitempty"`
	Count   int            `json:"count"`
	Records []usage.Record `json:"records"`
}

type invocationListResponse struct {
	TaskID  string              `json:"task_id,omitempty"`
	RunID   string              `json:"run_id,omitempty"`
	Count   int                 `json:"count"`
	Records []invocation.Record `json:"records"`
}

type errorResponse struct {
	Error     string `json:"error"`
	RunID     string `json:"run_id,omitempty"`
	AttemptID string `json:"attempt_id,omitempty"`
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
	if options.InvocationLedger == nil {
		options.InvocationLedger = invocation.NoopLedger{}
	}
	if options.Logger == nil {
		options.Logger = log.Default()
	}

	h := &Handler{
		store:            options.Store,
		reviewer:         options.Reviewer,
		contextManager:   options.ContextManager,
		policyRouter:     options.PolicyRouter,
		invocationLedger: options.InvocationLedger,
		requestTimeout:   options.RequestTimeout,
		logger:           options.Logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", method(http.MethodGet, health))
	mux.HandleFunc("/tasks/review-diff", method(http.MethodPost, h.reviewDiff))
	mux.HandleFunc("/tasks/", method(http.MethodGet, h.getTask))
	mux.HandleFunc("/invocations/recent", method(http.MethodGet, h.getRecentInvocations))
	mux.HandleFunc("/invocations/tasks/", method(http.MethodGet, h.getInvocationsByTask))
	mux.HandleFunc("/invocations/runs/", method(http.MethodGet, h.getInvocationsByRun))
	mux.HandleFunc("/usage/recent", method(http.MethodGet, h.getRecentUsage))
	mux.HandleFunc("/usage/tasks/", method(http.MethodGet, h.getUsageByTask))
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
	startedAt := time.Now().UTC()
	requestID := requestIDFrom(r)
	runID := runIDFrom(r)
	attemptID := attemptIDFrom(r)
	scenario := scenarioFrom(r)
	baseRecord := invocation.Record{
		RunID:      runID,
		AttemptID:  attemptID,
		Scenario:   scenario,
		Operation:  invocation.OperationReviewDiff,
		RequestID:  requestID,
		StartedAt:  startedAt,
		FinishedAt: startedAt,
	}

	var req reviewDiffRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); err != nil {
		h.writeRejectedReview(w, baseRecord, http.StatusBadRequest, invocation.ErrorValidation, "invalid json")
		return
	}

	if strings.TrimSpace(req.Diff) == "" {
		h.writeRejectedReview(w, baseRecord, http.StatusBadRequest, invocation.ErrorValidation, "diff is required")
		return
	}

	if req.Stream {
		h.writeRejectedReview(w, baseRecord, http.StatusBadRequest, invocation.ErrorValidation, "streaming is not supported")
		return
	}

	if h.reviewer == nil {
		h.writeRejectedReview(w, baseRecord, http.StatusBadGateway, invocation.ErrorDownstream, "litellm reviewer is not configured")
		return
	}

	taskID := tasks.NewID("task")
	contextOut, err := h.contextManager.Build(contextmgr.Input{
		CurrentDiff:     req.Diff,
		RepoSummary:     req.RepoSummary,
		PriorPlan:       req.PriorPlan,
		OldLogs:         req.Logs,
		IrrelevantNotes: req.Notes,
	})
	if err != nil {
		if errors.Is(err, contextmgr.ErrNoUsableContext) {
			h.writeRejectedReview(w, baseRecord, http.StatusBadRequest, invocation.ErrorValidation, "no usable context")
			return
		}
		h.writeRejectedReview(w, baseRecord, http.StatusBadRequest, invocation.ErrorValidation, "invalid context input")
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
		h.writeRejectedReview(w, baseRecord, http.StatusBadRequest, invocation.ErrorRouting, "no routable model")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	start := time.Now()
	resp, err := h.reviewer.ReviewDiff(ctx, litellm.ReviewRequest{
		Diff:         req.Diff,
		RepoSummary:  req.RepoSummary,
		FinalContext: contextOut.FinalContext,
		ModelAlias:   decision.ModelAlias,
		RequestID:    requestID,
	})
	if err != nil {
		statusCode := mapLiteLLMError(err)
		latencyMS := time.Since(start).Milliseconds()
		task := h.store.Save(tasks.Task{
			ID:          taskID,
			RunID:       runID,
			AttemptID:   attemptID,
			RequestID:   requestID,
			Status:      tasks.StatusFailed,
			Model:       decision.ModelAlias,
			RouteReason: decision.RouteReason,
			LatencyMS:   latencyMS,
			Error:       err.Error(),
		})
		h.logInvocation(invocation.Record{
			RunID:         runID,
			AttemptID:     attemptID,
			Scenario:      scenario,
			Operation:     invocation.OperationReviewDiff,
			TaskID:        task.ID,
			RequestID:     task.RequestID,
			HTTPStatus:    statusCode,
			TaskStatus:    invocation.StatusFailed,
			ErrorKind:     mapInvocationErrorKind(err),
			Error:         task.Error,
			LatencyMS:     task.LatencyMS,
			ContextChars:  len(contextOut.FinalContext),
			ContextReport: invocationReportFrom(report),
			ModelAlias:    decision.ModelAlias,
			RouteReason:   decision.RouteReason,
			Usage: &usage.Record{
				TaskID:      task.ID,
				RequestID:   task.RequestID,
				ModelAlias:  decision.ModelAlias,
				RouteReason: decision.RouteReason,
				LatencyMS:   task.LatencyMS,
				Status:      string(task.Status),
				Error:       task.Error,
			},
			StartedAt:  startedAt,
			FinishedAt: time.Now().UTC(),
		})
		h.logger.Printf("review_diff failed task_id=%s run_id=%s attempt_id=%s request_id=%s status=%d model=%s route_reason=%q error=%v", taskID, runID, attemptID, requestID, statusCode, decision.ModelAlias, decision.RouteReason, err)
		resp := responseFromTask(task)
		resp.ContextReport = &report
		writeJSON(w, statusCode, resp)
		return
	}

	task := h.store.Save(tasks.Task{
		ID:          taskID,
		RunID:       runID,
		AttemptID:   attemptID,
		RequestID:   requestID,
		Status:      tasks.StatusSuccess,
		Result:      resp.Result,
		Model:       decision.ModelAlias,
		RouteReason: decision.RouteReason,
		LatencyMS:   resp.LatencyMS,
	})
	h.logInvocation(invocation.Record{
		RunID:         runID,
		AttemptID:     attemptID,
		Scenario:      scenario,
		Operation:     invocation.OperationReviewDiff,
		TaskID:        task.ID,
		RequestID:     task.RequestID,
		HTTPStatus:    http.StatusOK,
		TaskStatus:    invocation.StatusSuccess,
		ErrorKind:     invocation.ErrorNone,
		LatencyMS:     resp.LatencyMS,
		ContextChars:  len(contextOut.FinalContext),
		ContextReport: invocationReportFrom(report),
		ModelAlias:    decision.ModelAlias,
		ReturnedModel: resp.Model,
		RouteReason:   decision.RouteReason,
		Usage: &usage.Record{
			TaskID:           task.ID,
			RequestID:        task.RequestID,
			ModelAlias:       decision.ModelAlias,
			ReturnedModel:    resp.Model,
			RouteReason:      decision.RouteReason,
			LatencyMS:        resp.LatencyMS,
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
			Status:           string(task.Status),
		},
		StartedAt:  startedAt,
		FinishedAt: time.Now().UTC(),
	})

	h.logger.Printf("review_diff success task_id=%s run_id=%s attempt_id=%s request_id=%s model=%s route_reason=%q latency_ms=%d", taskID, runID, attemptID, requestID, decision.ModelAlias, decision.RouteReason, resp.LatencyMS)
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

func (h *Handler) getRecentUsage(w http.ResponseWriter, r *http.Request) {
	limit, ok := parseListLimit(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid limit")
		return
	}

	records, err := h.invocationLedger.Recent(limit)
	if err != nil {
		h.logger.Printf("usage read failed endpoint=recent error=%v", err)
		writeError(w, http.StatusInternalServerError, "usage read failed")
		return
	}
	usageRecords := invocation.ProjectUsage(records)

	writeJSON(w, http.StatusOK, usageListResponse{
		Count:   len(usageRecords),
		Records: usageRecords,
	})
}

func (h *Handler) getUsageByTask(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/usage/tasks/")
	if taskID == "" || strings.Contains(taskID, "/") {
		writeError(w, http.StatusNotFound, "usage not found")
		return
	}

	records, err := h.invocationLedger.ForTask(taskID)
	if err != nil {
		h.logger.Printf("usage read failed endpoint=task task_id=%s error=%v", taskID, err)
		writeError(w, http.StatusInternalServerError, "usage read failed")
		return
	}
	usageRecords := invocation.ProjectUsage(records)

	writeJSON(w, http.StatusOK, usageListResponse{
		TaskID:  taskID,
		Count:   len(usageRecords),
		Records: usageRecords,
	})
}

func (h *Handler) getRecentInvocations(w http.ResponseWriter, r *http.Request) {
	limit, ok := parseListLimit(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid limit")
		return
	}

	records, err := h.invocationLedger.Recent(limit)
	if err != nil {
		h.logger.Printf("invocation read failed endpoint=recent error=%v", err)
		writeError(w, http.StatusInternalServerError, "invocation read failed")
		return
	}

	writeJSON(w, http.StatusOK, invocationListResponse{
		Count:   len(records),
		Records: records,
	})
}

func (h *Handler) getInvocationsByTask(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/invocations/tasks/")
	if taskID == "" || strings.Contains(taskID, "/") {
		writeError(w, http.StatusNotFound, "invocations not found")
		return
	}

	records, err := h.invocationLedger.ForTask(taskID)
	if err != nil {
		h.logger.Printf("invocation read failed endpoint=task task_id=%s error=%v", taskID, err)
		writeError(w, http.StatusInternalServerError, "invocation read failed")
		return
	}

	writeJSON(w, http.StatusOK, invocationListResponse{
		TaskID:  taskID,
		Count:   len(records),
		Records: records,
	})
}

func (h *Handler) getInvocationsByRun(w http.ResponseWriter, r *http.Request) {
	runID := strings.TrimPrefix(r.URL.Path, "/invocations/runs/")
	if runID == "" || strings.Contains(runID, "/") {
		writeError(w, http.StatusNotFound, "invocations not found")
		return
	}

	records, err := h.invocationLedger.ForRun(runID)
	if err != nil {
		h.logger.Printf("invocation read failed endpoint=run run_id=%s error=%v", runID, err)
		writeError(w, http.StatusInternalServerError, "invocation read failed")
		return
	}

	writeJSON(w, http.StatusOK, invocationListResponse{
		RunID:   runID,
		Count:   len(records),
		Records: records,
	})
}

func parseListLimit(r *http.Request) (int, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return 20, true
	}

	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 {
		return 0, false
	}
	if limit > 100 {
		return 100, true
	}
	return limit, true
}

func requestIDFrom(r *http.Request) string {
	if requestID := strings.TrimSpace(r.Header.Get("X-Request-ID")); requestID != "" {
		return requestID
	}
	return tasks.NewID("req")
}

func runIDFrom(r *http.Request) string {
	if runID := strings.TrimSpace(r.Header.Get("X-Run-ID")); runID != "" {
		return runID
	}
	return tasks.NewID("run")
}

func attemptIDFrom(r *http.Request) string {
	if attemptID := strings.TrimSpace(r.Header.Get("X-Attempt-ID")); attemptID != "" {
		return attemptID
	}
	return tasks.NewID("attempt")
}

func scenarioFrom(r *http.Request) string {
	if scenario := strings.TrimSpace(r.Header.Get("X-Scenario")); scenario != "" {
		return scenario
	}
	return "api.review_diff"
}

func mapLiteLLMError(err error) int {
	var llmErr *litellm.Error
	if errors.As(err, &llmErr) && llmErr.Kind == litellm.KindTimeout {
		return http.StatusGatewayTimeout
	}
	return http.StatusBadGateway
}

func mapInvocationErrorKind(err error) string {
	var llmErr *litellm.Error
	if !errors.As(err, &llmErr) {
		return invocation.ErrorDownstream
	}
	switch llmErr.Kind {
	case litellm.KindTimeout:
		return invocation.ErrorTimeout
	case litellm.KindBadResponse:
		return invocation.ErrorBadResponse
	default:
		return invocation.ErrorDownstream
	}
}

func responseFromTask(task tasks.Task) taskResponse {
	return taskResponse{
		TaskID:      task.ID,
		RunID:       task.RunID,
		AttemptID:   task.AttemptID,
		RequestID:   task.RequestID,
		Status:      string(task.Status),
		Result:      task.Result,
		Model:       task.Model,
		RouteReason: task.RouteReason,
		LatencyMS:   task.LatencyMS,
		Error:       task.Error,
	}
}

func (h *Handler) writeRejectedReview(w http.ResponseWriter, record invocation.Record, statusCode int, errorKind string, message string) {
	now := time.Now().UTC()
	record.HTTPStatus = statusCode
	record.TaskStatus = invocation.StatusRejected
	record.ErrorKind = errorKind
	record.Error = message
	record.FinishedAt = now
	record.LatencyMS = now.Sub(record.StartedAt).Milliseconds()
	h.logInvocation(record)
	writeJSON(w, statusCode, errorResponse{Error: message, RunID: record.RunID, AttemptID: record.AttemptID})
}

func (h *Handler) logInvocation(record invocation.Record) {
	if err := h.invocationLedger.Log(record); err != nil {
		h.logger.Printf("invocation log failed task_id=%s run_id=%s attempt_id=%s error=%v", record.TaskID, record.RunID, record.AttemptID, err)
	}
}

func invocationReportFrom(report contextReport) *invocation.ContextReport {
	return &invocation.ContextReport{
		KeptBlocks:       report.KeptBlocks,
		CompressedBlocks: report.CompressedBlocks,
		DroppedBlocks:    report.DroppedBlocks,
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
