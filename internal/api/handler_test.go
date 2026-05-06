package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-nanobot-litellm-lab/internal/invocation"
	"go-nanobot-litellm-lab/internal/litellm"
	"go-nanobot-litellm-lab/internal/router"
	"go-nanobot-litellm-lab/internal/tasks"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}

	var body healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("status = %q, want ok", body.Status)
	}

	if body.Service != serviceName {
		t.Fatalf("service = %q, want %q", body.Service, serviceName)
	}
}

func TestHealthRejectsUnsupportedMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}

	if got := rec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("allow = %q, want %q", got, http.MethodGet)
	}
}

func TestReviewDiffSuccessStoresTask(t *testing.T) {
	store := tasks.NewStore()
	reviewer := &fakeReviewer{
		resp: litellm.ReviewResponse{
			Result:    "review result",
			Model:     "code-cheap",
			LatencyMS: 25,
			Usage:     litellm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		},
	}
	ledger := &fakeInvocationLedger{}
	body := stringsReader(`{"diff":"diff --git a/a.go b/a.go","repo_summary":"Go service","stream":false}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks/review-diff", body)
	req.Header.Set("X-Request-ID", "req-test")
	req.Header.Set("X-Run-ID", "run-test")
	req.Header.Set("X-Attempt-ID", "attempt-test")
	req.Header.Set("X-Scenario", "test.success")
	rec := httptest.NewRecorder()

	NewHandler(Options{Store: store, Reviewer: reviewer, InvocationLedger: ledger, RequestTimeout: time.Second}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var resp taskResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TaskID == "" {
		t.Fatalf("task id is empty")
	}
	if resp.RunID != "run-test" {
		t.Fatalf("run id = %q, want run-test", resp.RunID)
	}
	if resp.AttemptID != "attempt-test" {
		t.Fatalf("attempt id = %q, want attempt-test", resp.AttemptID)
	}
	if resp.RequestID != "req-test" {
		t.Fatalf("request id = %q, want req-test", resp.RequestID)
	}
	if resp.Status != string(tasks.StatusSuccess) {
		t.Fatalf("status = %q, want success", resp.Status)
	}
	if resp.Result != "review result" {
		t.Fatalf("result = %q, want review result", resp.Result)
	}
	if strings.TrimSpace(resp.RouteReason) == "" {
		t.Fatalf("route reason is empty")
	}
	if resp.ContextReport == nil {
		t.Fatalf("context report is nil")
	}
	if !contains(resp.ContextReport.KeptBlocks, "current_diff") {
		t.Fatalf("kept blocks missing current_diff: %+v", resp.ContextReport.KeptBlocks)
	}

	got, ok := store.Get(resp.TaskID)
	if !ok {
		t.Fatalf("stored task not found")
	}
	if got.Result != "review result" {
		t.Fatalf("stored result = %q, want review result", got.Result)
	}
	if reviewer.req.RequestID != "req-test" {
		t.Fatalf("reviewer request id = %q, want req-test", reviewer.req.RequestID)
	}
	if reviewer.req.ModelAlias != "code-cheap" {
		t.Fatalf("reviewer model alias = %q, want code-cheap", reviewer.req.ModelAlias)
	}
	if strings.TrimSpace(reviewer.req.FinalContext) == "" {
		t.Fatalf("reviewer final context is empty")
	}
	if len(ledger.records) != 1 {
		t.Fatalf("invocation records = %d, want 1", len(ledger.records))
	}
	record := ledger.records[0]
	if record.RunID != "run-test" {
		t.Fatalf("invocation run id = %q, want run-test", record.RunID)
	}
	if record.AttemptID != "attempt-test" {
		t.Fatalf("invocation attempt id = %q, want attempt-test", record.AttemptID)
	}
	if record.Scenario != "test.success" {
		t.Fatalf("invocation scenario = %q, want test.success", record.Scenario)
	}
	if record.TaskID != resp.TaskID {
		t.Fatalf("invocation task id = %q, want %q", record.TaskID, resp.TaskID)
	}
	if record.TaskStatus != invocation.StatusSuccess {
		t.Fatalf("invocation status = %q, want success", record.TaskStatus)
	}
	if record.HTTPStatus != http.StatusOK {
		t.Fatalf("invocation http status = %d, want 200", record.HTTPStatus)
	}
	if record.ErrorKind != invocation.ErrorNone {
		t.Fatalf("invocation error kind = %q, want none", record.ErrorKind)
	}
	if record.Usage == nil {
		t.Fatalf("invocation usage is nil")
	}
	if record.Usage.ModelAlias != "code-cheap" {
		t.Fatalf("usage model alias = %q, want code-cheap", record.Usage.ModelAlias)
	}
	if record.Usage.ReturnedModel != "code-cheap" {
		t.Fatalf("usage returned model = %q, want code-cheap", record.Usage.ReturnedModel)
	}
	if record.Usage.TotalTokens != 15 {
		t.Fatalf("usage total tokens = %d, want 15", record.Usage.TotalTokens)
	}
}

func TestGetTask(t *testing.T) {
	store := tasks.NewStore()
	store.Save(tasks.Task{
		ID:          "task_test",
		RequestID:   "req_test",
		Status:      tasks.StatusSuccess,
		Result:      "ok",
		Model:       "code-cheap",
		RouteReason: "task=review_diff selected=code-cheap",
		LatencyMS:   10,
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks/task_test", nil)
	rec := httptest.NewRecorder()

	NewHandler(Options{Store: store}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var resp taskResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TaskID != "task_test" {
		t.Fatalf("task id = %q, want task_test", resp.TaskID)
	}
}

func TestGetRecentUsage(t *testing.T) {
	ledger := &fakeInvocationLedger{
		recent: []invocation.Record{
			{TaskID: "task_1", TaskStatus: invocation.StatusSuccess},
			{TaskID: "task_2", TaskStatus: invocation.StatusFailed},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/usage/recent?limit=2", nil)
	rec := httptest.NewRecorder()

	NewHandler(Options{InvocationLedger: ledger}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if ledger.limit != 2 {
		t.Fatalf("limit = %d, want 2", ledger.limit)
	}

	var resp usageListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Count != 2 {
		t.Fatalf("count = %d, want 2", resp.Count)
	}
}

func TestGetRecentUsageRejectsInvalidLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/usage/recent?limit=bad", nil)
	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetUsageByTask(t *testing.T) {
	ledger := &fakeInvocationLedger{
		byTask: []invocation.Record{
			{TaskID: "task_test", TaskStatus: invocation.StatusSuccess},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/usage/tasks/task_test", nil)
	rec := httptest.NewRecorder()

	NewHandler(Options{InvocationLedger: ledger}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if ledger.taskID != "task_test" {
		t.Fatalf("task id = %q, want task_test", ledger.taskID)
	}

	var resp usageListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TaskID != "task_test" {
		t.Fatalf("response task id = %q, want task_test", resp.TaskID)
	}
	if resp.Count != 1 {
		t.Fatalf("count = %d, want 1", resp.Count)
	}
}

func TestGetRecentInvocations(t *testing.T) {
	ledger := &fakeInvocationLedger{
		recent: []invocation.Record{
			{RunID: "run_1", AttemptID: "attempt_1", TaskID: "task_1", TaskStatus: invocation.StatusSuccess},
			{RunID: "run_2", AttemptID: "attempt_2", TaskID: "task_2", TaskStatus: invocation.StatusFailed},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/invocations/recent?limit=2", nil)
	rec := httptest.NewRecorder()

	NewHandler(Options{InvocationLedger: ledger}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if ledger.limit != 2 {
		t.Fatalf("limit = %d, want 2", ledger.limit)
	}

	var resp invocationListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Count != 2 {
		t.Fatalf("count = %d, want 2", resp.Count)
	}
	if resp.Records[0].RunID != "run_1" {
		t.Fatalf("run id = %q, want run_1", resp.Records[0].RunID)
	}
}

func TestGetInvocationsByTask(t *testing.T) {
	ledger := &fakeInvocationLedger{
		byTask: []invocation.Record{
			{RunID: "run_test", AttemptID: "attempt_test", TaskID: "task_test", TaskStatus: invocation.StatusSuccess},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/invocations/tasks/task_test", nil)
	rec := httptest.NewRecorder()

	NewHandler(Options{InvocationLedger: ledger}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if ledger.taskID != "task_test" {
		t.Fatalf("task id = %q, want task_test", ledger.taskID)
	}

	var resp invocationListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TaskID != "task_test" {
		t.Fatalf("response task id = %q, want task_test", resp.TaskID)
	}
	if resp.Count != 1 {
		t.Fatalf("count = %d, want 1", resp.Count)
	}
}

func TestGetInvocationsByRun(t *testing.T) {
	ledger := &fakeInvocationLedger{
		byRun: []invocation.Record{
			{RunID: "run_test", AttemptID: "attempt_test", TaskID: "task_test", TaskStatus: invocation.StatusSuccess},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/invocations/runs/run_test", nil)
	rec := httptest.NewRecorder()

	NewHandler(Options{InvocationLedger: ledger}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if ledger.runID != "run_test" {
		t.Fatalf("run id = %q, want run_test", ledger.runID)
	}

	var resp invocationListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.RunID != "run_test" {
		t.Fatalf("response run id = %q, want run_test", resp.RunID)
	}
	if resp.Count != 1 {
		t.Fatalf("count = %d, want 1", resp.Count)
	}
}

func TestReviewDiffRejectsEmptyDiff(t *testing.T) {
	ledger := &fakeInvocationLedger{}
	req := httptest.NewRequest(http.MethodPost, "/tasks/review-diff", stringsReader(`{"diff":"  "}`))
	req.Header.Set("X-Run-ID", "run-reject")
	req.Header.Set("X-Attempt-ID", "attempt-reject")
	rec := httptest.NewRecorder()

	NewHandler(Options{Reviewer: &fakeReviewer{}, InvocationLedger: ledger}).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if len(ledger.records) != 1 {
		t.Fatalf("invocation records = %d, want 1", len(ledger.records))
	}
	record := ledger.records[0]
	if record.TaskID != "" {
		t.Fatalf("rejected record task id = %q, want empty", record.TaskID)
	}
	if record.RunID != "run-reject" || record.AttemptID != "attempt-reject" {
		t.Fatalf("rejected ids = %q/%q, want run-reject/attempt-reject", record.RunID, record.AttemptID)
	}
	if record.TaskStatus != invocation.StatusRejected {
		t.Fatalf("rejected status = %q, want rejected", record.TaskStatus)
	}
	if record.ErrorKind != invocation.ErrorValidation {
		t.Fatalf("error kind = %q, want validation", record.ErrorKind)
	}

	var body errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.RunID != "run-reject" || body.AttemptID != "attempt-reject" {
		t.Fatalf("error response ids = %q/%q, want run-reject/attempt-reject", body.RunID, body.AttemptID)
	}
}

func TestReviewDiffRejectsStreaming(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/tasks/review-diff", stringsReader(`{"diff":"diff","stream":true}`))
	rec := httptest.NewRecorder()

	NewHandler(Options{Reviewer: &fakeReviewer{}}).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestReviewDiffMapsTimeoutTo504(t *testing.T) {
	ledger := &fakeInvocationLedger{}
	req := httptest.NewRequest(http.MethodPost, "/tasks/review-diff", stringsReader(`{"diff":"diff"}`))
	rec := httptest.NewRecorder()

	NewHandler(Options{Reviewer: &fakeReviewer{err: &litellm.Error{Kind: litellm.KindTimeout, Err: context.DeadlineExceeded}}, InvocationLedger: ledger}).ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want 504; body=%s", rec.Code, rec.Body.String())
	}
	if len(ledger.records) != 1 {
		t.Fatalf("invocation records = %d, want 1", len(ledger.records))
	}
	if ledger.records[0].ErrorKind != invocation.ErrorTimeout {
		t.Fatalf("error kind = %q, want timeout", ledger.records[0].ErrorKind)
	}
	if ledger.records[0].TaskStatus != invocation.StatusFailed {
		t.Fatalf("task status = %q, want failed", ledger.records[0].TaskStatus)
	}
}

func TestReviewDiffMapsDownstreamErrorTo502(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/tasks/review-diff", stringsReader(`{"diff":"diff"}`))
	rec := httptest.NewRecorder()
	ledger := &fakeInvocationLedger{}

	NewHandler(Options{Reviewer: &fakeReviewer{err: &litellm.Error{Kind: litellm.KindDownstream, Err: errors.New("downstream failed")}}, InvocationLedger: ledger}).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502; body=%s", rec.Code, rec.Body.String())
	}
	if len(ledger.records) != 1 {
		t.Fatalf("invocation records = %d, want 1", len(ledger.records))
	}
	if ledger.records[0].TaskStatus != invocation.StatusFailed {
		t.Fatalf("invocation status = %q, want failed", ledger.records[0].TaskStatus)
	}
	if ledger.records[0].ErrorKind != invocation.ErrorDownstream {
		t.Fatalf("invocation error kind = %q, want downstream", ledger.records[0].ErrorKind)
	}
	if ledger.records[0].Usage == nil {
		t.Fatalf("invocation usage is nil")
	}
	if ledger.records[0].Usage.Status != string(tasks.StatusFailed) {
		t.Fatalf("usage status = %q, want failed", ledger.records[0].Usage.Status)
	}
	if ledger.records[0].Usage.Error == "" {
		t.Fatalf("usage error is empty")
	}
}

func TestReviewDiffInvocationWriteFailureDoesNotFailTask(t *testing.T) {
	reviewer := &fakeReviewer{
		resp: litellm.ReviewResponse{
			Result:    "review result",
			Model:     "code-cheap",
			LatencyMS: 25,
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/tasks/review-diff", stringsReader(`{"diff":"diff"}`))
	rec := httptest.NewRecorder()

	NewHandler(Options{Reviewer: reviewer, InvocationLedger: &fakeInvocationLedger{err: errors.New("write failed")}}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
}

func TestReviewDiffHighQualityBudgetRoutesToSmartModel(t *testing.T) {
	store := tasks.NewStore()
	reviewer := &fakeReviewer{
		resp: litellm.ReviewResponse{
			Result:    "routed review result",
			Model:     "code-smart",
			LatencyMS: 20,
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/tasks/review-diff", stringsReader(`{"diff":"diff","budget_hint":"high_quality"}`))
	rec := httptest.NewRecorder()

	NewHandler(Options{Store: store, Reviewer: reviewer, RequestTimeout: time.Second}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if reviewer.req.ModelAlias != "code-smart" {
		t.Fatalf("reviewer model alias = %q, want code-smart", reviewer.req.ModelAlias)
	}
}

func TestReviewDiffReturns400WhenNoRoutableModel(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/tasks/review-diff", stringsReader(`{"diff":"diff"}`))
	rec := httptest.NewRecorder()

	NewHandler(Options{
		Reviewer:     &fakeReviewer{},
		PolicyRouter: &fakePolicyRouter{err: errors.New("no route")},
	}).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestReviewDiffEndToEndWithMockLiteLLM(t *testing.T) {
	mockLiteLLM := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q, want /v1/chat/completions", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model": "code-cheap",
			"choices": [
				{"message": {"role": "assistant", "content": "mock review from litellm"}}
			]
		}`))
	}))
	defer mockLiteLLM.Close()

	reviewer, err := litellm.NewClient(litellm.Config{
		BaseURL: mockLiteLLM.URL,
		Model:   "code-cheap",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("new litellm client: %v", err)
	}

	app := httptest.NewServer(NewHandler(Options{Store: tasks.NewStore(), Reviewer: reviewer, RequestTimeout: time.Second}))
	defer app.Close()

	resp, err := http.Post(app.URL+"/tasks/review-diff", "application/json", stringsReader(`{"diff":"diff --git a/a.go b/a.go"}`))
	if err != nil {
		t.Fatalf("post review diff: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", resp.StatusCode, body)
	}

	var decoded taskResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded.Result != "mock review from litellm" {
		t.Fatalf("result = %q, want mock review from litellm", decoded.Result)
	}
	if decoded.Model != "code-cheap" {
		t.Fatalf("model = %q, want code-cheap", decoded.Model)
	}
}

type fakeReviewer struct {
	req  litellm.ReviewRequest
	resp litellm.ReviewResponse
	err  error
}

func (f *fakeReviewer) ReviewDiff(_ context.Context, req litellm.ReviewRequest) (litellm.ReviewResponse, error) {
	f.req = req
	return f.resp, f.err
}

type fakePolicyRouter struct {
	decision router.Decision
	err      error
}

func (f *fakePolicyRouter) Route(_ router.Input) (router.Decision, error) {
	if f.err != nil {
		return router.Decision{}, f.err
	}
	if f.decision.ModelAlias == "" {
		f.decision.ModelAlias = "code-cheap"
		f.decision.RouteReason = "task=review_diff context_chars=10 budget=balanced selected=code-cheap score=0.780"
	}
	return f.decision, nil
}

type fakeInvocationLedger struct {
	records []invocation.Record
	recent  []invocation.Record
	byTask  []invocation.Record
	byRun   []invocation.Record
	limit   int
	taskID  string
	runID   string
	err     error
}

func (f *fakeInvocationLedger) Log(record invocation.Record) error {
	f.records = append(f.records, record)
	return f.err
}

func (f *fakeInvocationLedger) Recent(limit int) ([]invocation.Record, error) {
	f.limit = limit
	return f.recent, f.err
}

func (f *fakeInvocationLedger) ForTask(taskID string) ([]invocation.Record, error) {
	f.taskID = taskID
	return f.byTask, f.err
}

func (f *fakeInvocationLedger) ForRun(runID string) ([]invocation.Record, error) {
	f.runID = runID
	return f.byRun, f.err
}

func stringsReader(s string) *strings.Reader {
	return strings.NewReader(s)
}

func contains(list []string, needle string) bool {
	for _, v := range list {
		if v == needle {
			return true
		}
	}
	return false
}
