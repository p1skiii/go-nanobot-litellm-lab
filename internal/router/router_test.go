package router

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRouteHighQualitySelectsSmart(t *testing.T) {
	r := NewDefault()
	decision, err := r.Route(Input{
		TaskType:     "review_diff",
		ContextChars: 4000,
		BudgetHint:   "high_quality",
	})
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	if decision.ModelAlias != "code-smart" {
		t.Fatalf("model alias = %q, want code-smart", decision.ModelAlias)
	}
	if strings.TrimSpace(decision.RouteReason) == "" {
		t.Fatalf("route reason is empty")
	}
}

func TestRouteLowBudgetSelectsCheap(t *testing.T) {
	r := NewDefault()
	decision, err := r.Route(Input{
		TaskType:     "review_diff",
		ContextChars: 4000,
		BudgetHint:   "low",
	})
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	if decision.ModelAlias != "code-cheap" {
		t.Fatalf("model alias = %q, want code-cheap", decision.ModelAlias)
	}
}

func TestRouteFiltersStreamUnsupported(t *testing.T) {
	r := newRouter([]ModelProfile{
		{Alias: "cheap-no-stream", Quality: 0.8, Cost: 0.2, Latency: 0.9, ContextLimit: 128000, SupportsStream: false},
		{Alias: "smart-stream", Quality: 0.7, Cost: 0.7, Latency: 0.5, ContextLimit: 128000, SupportsStream: true},
	}, defaultsConfig{}, scoringConfig{QualityWeight: 0.5, CostWeight: 0.3, LatencyWeight: 0.2})

	decision, err := r.Route(Input{
		TaskType:        "review_diff",
		ContextChars:    1024,
		StreamRequested: true,
	})
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	if decision.ModelAlias != "smart-stream" {
		t.Fatalf("model alias = %q, want smart-stream", decision.ModelAlias)
	}
}

func TestRouteRejectsWhenContextTooLarge(t *testing.T) {
	r := NewDefault()
	_, err := r.Route(Input{
		TaskType:     "review_diff",
		ContextChars: 200000,
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), ErrNoMatchingModel.Error()) {
		t.Fatalf("err = %v, want ErrNoMatchingModel", err)
	}
}

func TestNewFromFiles(t *testing.T) {
	dir := t.TempDir()
	modelsPath := filepath.Join(dir, "models.yaml")
	policiesPath := filepath.Join(dir, "policies.yaml")

	if err := os.WriteFile(modelsPath, []byte(`
models:
  - alias: model-a
    quality: 0.9
    cost: 0.4
    latency: 0.7
    context_limit: 16000
    supports_stream: true
`), 0o644); err != nil {
		t.Fatalf("write models: %v", err)
	}

	if err := os.WriteFile(policiesPath, []byte(`
defaults:
  task_type: review_diff
  model_alias: model-a
  fallback_chain: []
router:
  scoring:
    quality_weight: 0.5
    cost_weight: 0.3
    latency_weight: 0.2
`), 0o644); err != nil {
		t.Fatalf("write policies: %v", err)
	}

	r, err := NewFromFiles(modelsPath, policiesPath)
	if err != nil {
		t.Fatalf("new from files: %v", err)
	}
	decision, err := r.Route(Input{TaskType: "review_diff", ContextChars: 1000})
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	if decision.ModelAlias != "model-a" {
		t.Fatalf("model alias = %q, want model-a", decision.ModelAlias)
	}
}
