package contextmgr

import (
	"errors"
	"strings"
	"testing"
)

func TestBuildKeepCompressDrop(t *testing.T) {
	m := NewManager(Config{
		MaxFinalChars:   20000,
		MaxPlanChars:    40,
		MaxLogChars:     120,
		MaxSummaryChars: 80,
	})

	out, err := m.Build(Input{
		CurrentDiff:     "diff --git a/a.go b/a.go\n+fmt.Println(\"x\")",
		RepoSummary:     "small service",
		PriorPlan:       "this is a long old plan that should be compressed into a smaller one-line representation",
		OldLogs:         "INFO boot\nERROR downstream failed\nINFO retrying",
		IrrelevantNotes: "old chat note that should drop",
	})
	if err != nil {
		t.Fatalf("build context: %v", err)
	}

	if !contains(out.KeptBlocks, "current_diff") {
		t.Fatalf("kept blocks missing current_diff: %+v", out.KeptBlocks)
	}
	if !contains(out.KeptBlocks, "repo_summary") {
		t.Fatalf("kept blocks missing repo_summary: %+v", out.KeptBlocks)
	}
	if !contains(out.CompressedBlocks, "old_plan") {
		t.Fatalf("compressed blocks missing old_plan: %+v", out.CompressedBlocks)
	}
	if !contains(out.CompressedBlocks, "old_logs") {
		t.Fatalf("compressed blocks missing old_logs: %+v", out.CompressedBlocks)
	}
	if !contains(out.DroppedBlocks, "irrelevant_notes") {
		t.Fatalf("dropped blocks missing irrelevant_notes: %+v", out.DroppedBlocks)
	}
	if strings.TrimSpace(out.FinalContext) == "" {
		t.Fatalf("final context is empty")
	}
	if !strings.Contains(out.FinalContext, "Current Diff:") {
		t.Fatalf("final context missing diff section: %q", out.FinalContext)
	}
}

func TestBuildRequiresUsableContext(t *testing.T) {
	m := NewManager(Config{})
	_, err := m.Build(Input{IrrelevantNotes: "only notes"})
	if !errors.Is(err, ErrNoUsableContext) {
		t.Fatalf("err = %v, want ErrNoUsableContext", err)
	}
}

func contains(list []string, needle string) bool {
	for _, v := range list {
		if v == needle {
			return true
		}
	}
	return false
}
