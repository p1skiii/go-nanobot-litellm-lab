package usage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Record struct {
	TaskID           string    `json:"task_id"`
	RequestID        string    `json:"request_id,omitempty"`
	ModelAlias       string    `json:"model_alias,omitempty"`
	ReturnedModel    string    `json:"returned_model,omitempty"`
	RouteReason      string    `json:"route_reason,omitempty"`
	LatencyMS        int64     `json:"latency_ms"`
	PromptTokens     int       `json:"prompt_tokens,omitempty"`
	CompletionTokens int       `json:"completion_tokens,omitempty"`
	TotalTokens      int       `json:"total_tokens,omitempty"`
	Status           string    `json:"status"`
	Error            string    `json:"error,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type JSONLLogger struct {
	path string
	mu   sync.Mutex
}

func NewJSONLLogger(path string) (*JSONLLogger, error) {
	if path == "" {
		return nil, fmt.Errorf("usage log path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create usage log dir: %w", err)
	}
	return &JSONLLogger{path: path}, nil
}

func (l *JSONLLogger) Log(record Record) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}

	raw, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal usage record: %w", err)
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open usage log: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(raw, '\n')); err != nil {
		return fmt.Errorf("write usage log: %w", err)
	}
	return nil
}

func (l *JSONLLogger) Recent(limit int) ([]Record, error) {
	if limit <= 0 {
		return []Record{}, nil
	}

	records, err := readRecords(l.path)
	if err != nil {
		return nil, err
	}
	if len(records) <= limit {
		return records, nil
	}
	return records[len(records)-limit:], nil
}

func (l *JSONLLogger) ForTask(taskID string) ([]Record, error) {
	records, err := readRecords(l.path)
	if err != nil {
		return nil, err
	}

	matches := make([]Record, 0)
	for _, record := range records {
		if record.TaskID == taskID {
			matches = append(matches, record)
		}
	}
	return matches, nil
}

func readRecords(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Record{}, nil
		}
		return nil, fmt.Errorf("open usage log: %w", err)
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
			return nil, fmt.Errorf("decode usage log line %d: %w", lineNo, err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan usage log: %w", err)
	}
	return records, nil
}

type NoopLogger struct{}

func (NoopLogger) Log(Record) error {
	return nil
}

func (NoopLogger) Recent(int) ([]Record, error) {
	return []Record{}, nil
}

func (NoopLogger) ForTask(string) ([]Record, error) {
	return []Record{}, nil
}
