package invocation

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"go-nanobot-litellm-lab/internal/usage"
)

func TestJSONLLedgerWritesAndReadsRecentRecords(t *testing.T) {
	ledger, err := NewJSONLLedger(filepath.Join(t.TempDir(), "invocations.jsonl"))
	if err != nil {
		t.Fatalf("new ledger: %v", err)
	}

	for _, taskID := range []string{"task_1", "task_2", "task_3"} {
		if err := ledger.Log(Record{RunID: "run_1", AttemptID: taskID, TaskID: taskID, TaskStatus: StatusSuccess}); err != nil {
			t.Fatalf("log %s: %v", taskID, err)
		}
	}

	records, err := ledger.Recent(2)
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

func TestJSONLLedgerFiltersByTaskAndRun(t *testing.T) {
	ledger, err := NewJSONLLedger(filepath.Join(t.TempDir(), "invocations.jsonl"))
	if err != nil {
		t.Fatalf("new ledger: %v", err)
	}

	records := []Record{
		{RunID: "run_1", AttemptID: "attempt_1", TaskID: "task_1", TaskStatus: StatusFailed},
		{RunID: "run_2", AttemptID: "attempt_2", TaskID: "task_2", TaskStatus: StatusSuccess},
		{RunID: "run_1", AttemptID: "attempt_3", TaskID: "task_1", TaskStatus: StatusSuccess},
	}
	for _, record := range records {
		if err := ledger.Log(record); err != nil {
			t.Fatalf("log: %v", err)
		}
	}

	byTask, err := ledger.ForTask("task_1")
	if err != nil {
		t.Fatalf("for task: %v", err)
	}
	if len(byTask) != 2 {
		t.Fatalf("task records = %d, want 2", len(byTask))
	}

	byRun, err := ledger.ForRun("run_1")
	if err != nil {
		t.Fatalf("for run: %v", err)
	}
	if len(byRun) != 2 {
		t.Fatalf("run records = %d, want 2", len(byRun))
	}
}

func TestUsageProjectionWrapsUsageRecord(t *testing.T) {
	finishedAt := time.Date(2026, 5, 6, 1, 2, 3, 0, time.UTC)
	record := Record{
		TaskID:        "task_1",
		RequestID:     "req_1",
		ModelAlias:    "code-cheap",
		ReturnedModel: "code-cheap",
		RouteReason:   "selected=code-cheap",
		LatencyMS:     123,
		TaskStatus:    StatusSuccess,
		FinishedAt:    finishedAt,
		Usage: &usage.Record{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	got := record.UsageProjection()
	if got.TaskID != "task_1" {
		t.Fatalf("task id = %q, want task_1", got.TaskID)
	}
	if got.TotalTokens != 15 {
		t.Fatalf("total tokens = %d, want 15", got.TotalTokens)
	}
	if !got.CreatedAt.Equal(finishedAt) {
		t.Fatalf("created at = %s, want %s", got.CreatedAt, finishedAt)
	}
}

func TestProjectUsageSkipsRejectedRecordsWithoutTask(t *testing.T) {
	records := []Record{
		{RunID: "run_1", AttemptID: "attempt_1", TaskStatus: StatusRejected},
		{RunID: "run_1", AttemptID: "attempt_2", TaskID: "task_1", TaskStatus: StatusSuccess},
	}

	got := ProjectUsage(records)
	if len(got) != 1 {
		t.Fatalf("usage records = %d, want 1", len(got))
	}
	if got[0].TaskID != "task_1" {
		t.Fatalf("task id = %q, want task_1", got[0].TaskID)
	}
}

func TestConcurrentLogAndRecentDoesNotReadPartialRecords(t *testing.T) {
	ledger, err := NewJSONLLedger(filepath.Join(t.TempDir(), "invocations.jsonl"))
	if err != nil {
		t.Fatalf("new ledger: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := ledger.Log(Record{RunID: "run_1", AttemptID: "attempt_1", TaskID: "task_1", TaskStatus: StatusSuccess}); err != nil {
				t.Errorf("log: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := ledger.Recent(10); err != nil && strings.Contains(err.Error(), "decode invocation log") {
				t.Errorf("recent saw partial record: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestJSONLLedgerRejectsEmptyPath(t *testing.T) {
	if _, err := NewJSONLLedger(""); err == nil {
		t.Fatalf("expected error")
	}
}
