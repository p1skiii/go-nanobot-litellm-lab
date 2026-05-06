package invocation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go-nanobot-litellm-lab/internal/usage"
)

const (
	OperationReviewDiff = "review_diff"

	StatusSuccess  = "success"
	StatusFailed   = "failed"
	StatusRejected = "rejected"

	ErrorNone        = "none"
	ErrorValidation  = "validation"
	ErrorTimeout     = "timeout"
	ErrorDownstream  = "downstream"
	ErrorBadResponse = "bad_response"
	ErrorRouting     = "routing"
)

type ContextReport struct {
	KeptBlocks       []string `json:"kept_blocks,omitempty"`
	CompressedBlocks []string `json:"compressed_blocks,omitempty"`
	DroppedBlocks    []string `json:"dropped_blocks,omitempty"`
}

type Record struct {
	RunID         string         `json:"run_id"`
	AttemptID     string         `json:"attempt_id"`
	Scenario      string         `json:"scenario"`
	Operation     string         `json:"operation"`
	TaskID        string         `json:"task_id,omitempty"`
	RequestID     string         `json:"request_id,omitempty"`
	HTTPStatus    int            `json:"http_status"`
	TaskStatus    string         `json:"task_status"`
	ErrorKind     string         `json:"error_kind"`
	Error         string         `json:"error,omitempty"`
	LatencyMS     int64          `json:"latency_ms"`
	ContextChars  int            `json:"context_chars,omitempty"`
	ContextReport *ContextReport `json:"context_report,omitempty"`
	ModelAlias    string         `json:"model_alias,omitempty"`
	ReturnedModel string         `json:"returned_model,omitempty"`
	RouteReason   string         `json:"route_reason,omitempty"`
	Usage         *usage.Record  `json:"usage,omitempty"`
	StartedAt     time.Time      `json:"started_at"`
	FinishedAt    time.Time      `json:"finished_at"`
}

type Ledger interface {
	Log(record Record) error
	Recent(limit int) ([]Record, error)
	ForTask(taskID string) ([]Record, error)
	ForRun(runID string) ([]Record, error)
}

type JSONLLedger struct {
	path string
	mu   sync.Mutex
}

func NewJSONLLedger(path string) (*JSONLLedger, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("invocation log path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create invocation log dir: %w", err)
	}
	return &JSONLLedger{path: path}, nil
}

func (l *JSONLLedger) Log(record Record) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if record.StartedAt.IsZero() {
		record.StartedAt = time.Now().UTC()
	}
	if record.FinishedAt.IsZero() {
		record.FinishedAt = time.Now().UTC()
	}
	if record.ErrorKind == "" {
		record.ErrorKind = ErrorNone
	}
	if record.Operation == "" {
		record.Operation = OperationReviewDiff
	}
	if record.Usage != nil {
		fillUsageFromInvocation(&record)
	}

	raw, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal invocation record: %w", err)
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open invocation log: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(raw, '\n')); err != nil {
		return fmt.Errorf("write invocation log: %w", err)
	}
	return nil
}

func (l *JSONLLedger) Recent(limit int) ([]Record, error) {
	if limit <= 0 {
		return []Record{}, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	records, err := readRecords(l.path)
	if err != nil {
		return nil, err
	}
	if len(records) <= limit {
		return records, nil
	}
	return records[len(records)-limit:], nil
}

func (l *JSONLLedger) ForTask(taskID string) ([]Record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	records, err := readRecords(l.path)
	if err != nil {
		return nil, err
	}
	return filter(records, func(record Record) bool {
		return record.TaskID == taskID
	}), nil
}

func (l *JSONLLedger) ForRun(runID string) ([]Record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	records, err := readRecords(l.path)
	if err != nil {
		return nil, err
	}
	return filter(records, func(record Record) bool {
		return record.RunID == runID
	}), nil
}

func ProjectUsage(records []Record) []usage.Record {
	projected := make([]usage.Record, 0, len(records))
	for _, record := range records {
		if record.TaskID == "" {
			continue
		}
		projected = append(projected, record.UsageProjection())
	}
	return projected
}

func (r Record) UsageProjection() usage.Record {
	if r.Usage != nil {
		projected := *r.Usage
		fillMissingUsageFields(&projected, r)
		return projected
	}

	createdAt := r.FinishedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	return usage.Record{
		TaskID:        r.TaskID,
		RequestID:     r.RequestID,
		ModelAlias:    r.ModelAlias,
		ReturnedModel: r.ReturnedModel,
		RouteReason:   r.RouteReason,
		LatencyMS:     r.LatencyMS,
		Status:        r.TaskStatus,
		Error:         r.Error,
		CreatedAt:     createdAt,
	}
}

type NoopLedger struct{}

func (NoopLedger) Log(Record) error {
	return nil
}

func (NoopLedger) Recent(int) ([]Record, error) {
	return []Record{}, nil
}

func (NoopLedger) ForTask(string) ([]Record, error) {
	return []Record{}, nil
}

func (NoopLedger) ForRun(string) ([]Record, error) {
	return []Record{}, nil
}

func readRecords(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Record{}, nil
		}
		return nil, fmt.Errorf("open invocation log: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	records := make([]Record, 0)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record Record
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("decode invocation log line %d: %w", lineNo, err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan invocation log: %w", err)
	}
	return records, nil
}

func filter(records []Record, keep func(Record) bool) []Record {
	matches := make([]Record, 0)
	for _, record := range records {
		if keep(record) {
			matches = append(matches, record)
		}
	}
	return matches
}

func fillUsageFromInvocation(record *Record) {
	fillMissingUsageFields(record.Usage, *record)
	if record.Usage.CreatedAt.IsZero() {
		record.Usage.CreatedAt = record.FinishedAt
	}
}

func fillMissingUsageFields(projected *usage.Record, record Record) {
	if projected.TaskID == "" {
		projected.TaskID = record.TaskID
	}
	if projected.RequestID == "" {
		projected.RequestID = record.RequestID
	}
	if projected.ModelAlias == "" {
		projected.ModelAlias = record.ModelAlias
	}
	if projected.ReturnedModel == "" {
		projected.ReturnedModel = record.ReturnedModel
	}
	if projected.RouteReason == "" {
		projected.RouteReason = record.RouteReason
	}
	if projected.LatencyMS == 0 {
		projected.LatencyMS = record.LatencyMS
	}
	if projected.Status == "" {
		projected.Status = record.TaskStatus
	}
	if projected.Error == "" {
		projected.Error = record.Error
	}
	if projected.CreatedAt.IsZero() {
		projected.CreatedAt = record.FinishedAt
	}
}
