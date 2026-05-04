package usage

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONLLoggerWritesRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	logger, err := NewJSONLLogger(path)
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	err = logger.Log(Record{
		TaskID:           "task_1",
		RequestID:        "req_1",
		ModelAlias:       "code-smart",
		ReturnedModel:    "mimo-v2-flash",
		RouteReason:      "selected=code-smart",
		LatencyMS:        123,
		PromptTokens:     10,
		CompletionTokens: 5,
		TotalTokens:      15,
		Status:           "success",
	})
	if err != nil {
		t.Fatalf("log: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open usage log: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatalf("missing usage record")
	}

	var got Record
	if err := json.Unmarshal(scanner.Bytes(), &got); err != nil {
		t.Fatalf("decode record: %v", err)
	}
	if got.TaskID != "task_1" {
		t.Fatalf("task id = %q, want task_1", got.TaskID)
	}
	if got.ModelAlias != "code-smart" {
		t.Fatalf("model alias = %q, want code-smart", got.ModelAlias)
	}
	if got.TotalTokens != 15 {
		t.Fatalf("total tokens = %d, want 15", got.TotalTokens)
	}
	if got.CreatedAt.IsZero() {
		t.Fatalf("created_at is zero")
	}
}

func TestJSONLLoggerRejectsEmptyPath(t *testing.T) {
	_, err := NewJSONLLogger("")
	if err == nil {
		t.Fatalf("expected error")
	}
}
