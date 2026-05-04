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

func TestJSONLLoggerReadsRecentRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	logger, err := NewJSONLLogger(path)
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	for _, taskID := range []string{"task_1", "task_2", "task_3"} {
		if err := logger.Log(Record{TaskID: taskID, Status: "success"}); err != nil {
			t.Fatalf("log %s: %v", taskID, err)
		}
	}

	records, err := logger.Recent(2)
	if err != nil {
		t.Fatalf("recent: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2", len(records))
	}
	if records[0].TaskID != "task_2" || records[1].TaskID != "task_3" {
		t.Fatalf("records = %+v, want task_2/task_3", records)
	}
}

func TestJSONLLoggerFiltersByTask(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	logger, err := NewJSONLLogger(path)
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	records := []Record{
		{TaskID: "task_1", Status: "failed"},
		{TaskID: "task_2", Status: "success"},
		{TaskID: "task_1", Status: "success"},
	}
	for _, record := range records {
		if err := logger.Log(record); err != nil {
			t.Fatalf("log: %v", err)
		}
	}

	got, err := logger.ForTask("task_1")
	if err != nil {
		t.Fatalf("for task: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	for _, record := range got {
		if record.TaskID != "task_1" {
			t.Fatalf("task id = %q, want task_1", record.TaskID)
		}
	}
}

func TestJSONLLoggerMissingFileReturnsEmptyRecords(t *testing.T) {
	logger, err := NewJSONLLogger(filepath.Join(t.TempDir(), "missing", "usage.jsonl"))
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	records, err := logger.Recent(10)
	if err != nil {
		t.Fatalf("recent: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("records = %d, want 0", len(records))
	}
}

func TestJSONLLoggerMalformedLineReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	if err := os.WriteFile(path, []byte("{bad-json}\n"), 0o644); err != nil {
		t.Fatalf("write usage log: %v", err)
	}
	logger, err := NewJSONLLogger(path)
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	if _, err := logger.Recent(10); err == nil {
		t.Fatalf("expected malformed json error")
	}
}

func TestJSONLLoggerRejectsEmptyPath(t *testing.T) {
	_, err := NewJSONLLogger("")
	if err == nil {
		t.Fatalf("expected error")
	}
}
