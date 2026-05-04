package litellm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReviewDiffCallsChatCompletion(t *testing.T) {
	var captured chatCompletionRequest
	var requestID string
	var auth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}

		requestID = r.Header.Get("X-Request-ID")
		auth = r.Header.Get("Authorization")

		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model": "code-cheap",
			"choices": [
				{"message": {"role": "assistant", "content": "looks good"}}
			]
		}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "sk-test",
		Model:   "code-cheap",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	resp, err := client.ReviewDiff(context.Background(), ReviewRequest{
		Diff:        "diff --git a/a.go b/a.go",
		RepoSummary: "Go service",
		RequestID:   "req-test",
	})
	if err != nil {
		t.Fatalf("review diff: %v", err)
	}

	if captured.Model != "code-cheap" {
		t.Fatalf("model = %q, want code-cheap", captured.Model)
	}
	if captured.Stream {
		t.Fatalf("stream = true, want false")
	}
	if len(captured.Messages) != 2 {
		t.Fatalf("messages length = %d, want 2", len(captured.Messages))
	}
	if requestID != "req-test" {
		t.Fatalf("request id = %q, want req-test", requestID)
	}
	if auth != "Bearer sk-test" {
		t.Fatalf("authorization = %q, want bearer token", auth)
	}
	if resp.Result != "looks good" {
		t.Fatalf("result = %q, want looks good", resp.Result)
	}
	if resp.Model != "code-cheap" {
		t.Fatalf("response model = %q, want code-cheap", resp.Model)
	}
}

func TestReviewDiffMapsDownstreamStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, Model: "code-cheap", Timeout: time.Second})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.ReviewDiff(context.Background(), ReviewRequest{Diff: "diff"})
	var llmErr *Error
	if !errors.As(err, &llmErr) {
		t.Fatalf("err = %T, want *Error", err)
	}
	if llmErr.Kind != KindDownstream {
		t.Fatalf("kind = %q, want downstream", llmErr.Kind)
	}
	if llmErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("status code = %d, want 502", llmErr.StatusCode)
	}
}

func TestReviewDiffMapsMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"model":"code-cheap","choices":[]}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, Model: "code-cheap", Timeout: time.Second})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.ReviewDiff(context.Background(), ReviewRequest{Diff: "diff"})
	var llmErr *Error
	if !errors.As(err, &llmErr) {
		t.Fatalf("err = %T, want *Error", err)
	}
	if llmErr.Kind != KindBadResponse {
		t.Fatalf("kind = %q, want bad_response", llmErr.Kind)
	}
}

func TestReviewDiffMapsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, Model: "code-cheap", Timeout: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.ReviewDiff(context.Background(), ReviewRequest{Diff: "diff"})
	var llmErr *Error
	if !errors.As(err, &llmErr) {
		t.Fatalf("err = %T, want *Error", err)
	}
	if llmErr.Kind != KindTimeout {
		t.Fatalf("kind = %q, want timeout", llmErr.Kind)
	}
}

func TestReviewDiffUsesFinalContextWhenProvided(t *testing.T) {
	var captured chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"code-cheap","choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, Model: "code-cheap", Timeout: time.Second})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.ReviewDiff(context.Background(), ReviewRequest{
		Diff:         "diff --git",
		RepoSummary:  "summary",
		FinalContext: "FINAL_CONTEXT_PAYLOAD",
	})
	if err != nil {
		t.Fatalf("review diff: %v", err)
	}

	if len(captured.Messages) != 2 {
		t.Fatalf("messages len = %d, want 2", len(captured.Messages))
	}
	if captured.Messages[1].Content != "FINAL_CONTEXT_PAYLOAD" {
		t.Fatalf("user content = %q, want FINAL_CONTEXT_PAYLOAD", captured.Messages[1].Content)
	}
}

func TestReviewDiffUsesRequestModelAliasWhenProvided(t *testing.T) {
	var captured chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"code-smart","choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, Model: "code-cheap", Timeout: time.Second})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.ReviewDiff(context.Background(), ReviewRequest{
		Diff:       "diff --git",
		ModelAlias: "code-smart",
	})
	if err != nil {
		t.Fatalf("review diff: %v", err)
	}

	if captured.Model != "code-smart" {
		t.Fatalf("model = %q, want code-smart", captured.Model)
	}
}
