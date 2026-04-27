package litellm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

type ReviewRequest struct {
	Diff        string
	RepoSummary string
	RequestID   string
}

type ReviewResponse struct {
	Result    string
	Model     string
	LatencyMS int64
}

type ErrorKind string

const (
	KindTimeout     ErrorKind = "timeout"
	KindDownstream  ErrorKind = "downstream"
	KindBadResponse ErrorKind = "bad_response"
)

type Error struct {
	Kind       ErrorKind
	StatusCode int
	Err        error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return string(e.Kind)
	}
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewClient(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("litellm base url is required")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("parse litellm base url: %w", err)
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "code-cheap"
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  strings.TrimSpace(cfg.APIKey),
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) ReviewDiff(ctx context.Context, req ReviewRequest) (ReviewResponse, error) {
	start := time.Now()

	payload := chatCompletionRequest{
		Model: c.model,
		Messages: []message{
			{
				Role:    "system",
				Content: "You are a concise code review assistant. Review the diff and return actionable findings.",
			},
			{
				Role:    "user",
				Content: buildReviewPrompt(req.Diff, req.RepoSummary),
			},
		},
		Stream: false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("marshal litellm request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("build litellm request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if req.RequestID != "" {
		httpReq.Header.Set("X-Request-ID", req.RequestID)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ReviewResponse{}, &Error{Kind: KindTimeout, Err: err}
		}
		return ReviewResponse{}, &Error{Kind: KindDownstream, Err: err}
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return ReviewResponse{}, &Error{Kind: KindBadResponse, StatusCode: httpResp.StatusCode, Err: err}
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode > 299 {
		return ReviewResponse{}, &Error{
			Kind:       KindDownstream,
			StatusCode: httpResp.StatusCode,
			Err:        fmt.Errorf("litellm status %d: %s", httpResp.StatusCode, strings.TrimSpace(string(respBody))),
		}
	}

	var decoded chatCompletionResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return ReviewResponse{}, &Error{Kind: KindBadResponse, StatusCode: httpResp.StatusCode, Err: err}
	}

	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return ReviewResponse{}, &Error{
			Kind:       KindBadResponse,
			StatusCode: httpResp.StatusCode,
			Err:        fmt.Errorf("litellm response did not include assistant content"),
		}
	}

	return ReviewResponse{
		Result:    decoded.Choices[0].Message.Content,
		Model:     decoded.Model,
		LatencyMS: time.Since(start).Milliseconds(),
	}, nil
}

func buildReviewPrompt(diff, repoSummary string) string {
	var b strings.Builder
	if strings.TrimSpace(repoSummary) != "" {
		b.WriteString("Repository summary:\n")
		b.WriteString(strings.TrimSpace(repoSummary))
		b.WriteString("\n\n")
	}
	b.WriteString("Review this diff:\n")
	b.WriteString(strings.TrimSpace(diff))
	return b.String()
}

type chatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}
