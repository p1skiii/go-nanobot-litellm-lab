package tasks

import "testing"

func TestStoreSavesAndGetsTask(t *testing.T) {
	store := NewStore()
	task := Task{
		ID:        "task_test",
		RequestID: "req_test",
		Status:    StatusSuccess,
		Result:    "ok",
		Model:     "code-cheap",
		LatencyMS: 12,
	}

	saved := store.Save(task)
	got, ok := store.Get(task.ID)
	if !ok {
		t.Fatalf("task not found")
	}
	if got.ID != saved.ID {
		t.Fatalf("id = %q, want %q", got.ID, saved.ID)
	}
	if got.CreatedAt.IsZero() {
		t.Fatalf("created at is zero")
	}
	if got.UpdatedAt.IsZero() {
		t.Fatalf("updated at is zero")
	}
}

func TestNewIDUsesPrefix(t *testing.T) {
	id := NewID("task")
	if len(id) <= len("task_") || id[:len("task_")] != "task_" {
		t.Fatalf("id = %q, want task_ prefix", id)
	}
}
