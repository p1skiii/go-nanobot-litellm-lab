package contextmgr

import (
	"errors"
	"regexp"
	"strings"
)

var ErrNoUsableContext = errors.New("no usable context")

type Input struct {
	CurrentDiff     string
	RepoSummary     string
	PriorPlan       string
	OldLogs         string
	IrrelevantNotes string
}

type Output struct {
	FinalContext     string
	KeptBlocks       []string
	CompressedBlocks []string
	DroppedBlocks    []string
}

type Config struct {
	MaxFinalChars   int
	MaxPlanChars    int
	MaxLogChars     int
	MaxSummaryChars int
}

type Manager struct {
	cfg Config
}

func NewManager(cfg Config) *Manager {
	if cfg.MaxFinalChars <= 0 {
		cfg.MaxFinalChars = 16000
	}
	if cfg.MaxPlanChars <= 0 {
		cfg.MaxPlanChars = 1800
	}
	if cfg.MaxLogChars <= 0 {
		cfg.MaxLogChars = 1200
	}
	if cfg.MaxSummaryChars <= 0 {
		cfg.MaxSummaryChars = 2000
	}
	return &Manager{cfg: cfg}
}

func (m *Manager) Build(input Input) (Output, error) {
	diff := normalize(input.CurrentDiff)
	summary := normalize(input.RepoSummary)
	plan := normalize(input.PriorPlan)
	logs := normalize(input.OldLogs)
	notes := normalize(input.IrrelevantNotes)

	if diff == "" && summary == "" && plan == "" && logs == "" {
		return Output{}, ErrNoUsableContext
	}

	var out Output
	var sections []string

	if diff != "" {
		sections = append(sections, renderSection("Current Diff", diff))
		out.KeptBlocks = append(out.KeptBlocks, "current_diff")
	}
	if summary != "" {
		keptSummary := capText(summary, m.cfg.MaxSummaryChars)
		if keptSummary != summary {
			out.CompressedBlocks = append(out.CompressedBlocks, "repo_summary")
		} else {
			out.KeptBlocks = append(out.KeptBlocks, "repo_summary")
		}
		sections = append(sections, renderSection("Repository Summary", keptSummary))
	}
	if plan != "" {
		compressedPlan := compressGeneric(plan, m.cfg.MaxPlanChars)
		sections = append(sections, renderSection("Prior Plan (Compressed)", compressedPlan))
		out.CompressedBlocks = append(out.CompressedBlocks, "old_plan")
	}
	if logs != "" {
		compressedLogs := compressLogs(logs, m.cfg.MaxLogChars)
		if compressedLogs != "" {
			sections = append(sections, renderSection("Old Logs (Compressed)", compressedLogs))
			out.CompressedBlocks = append(out.CompressedBlocks, "old_logs")
		} else {
			out.DroppedBlocks = append(out.DroppedBlocks, "old_logs")
		}
	}
	if notes != "" {
		out.DroppedBlocks = append(out.DroppedBlocks, "irrelevant_notes")
	}

	final := strings.Join(sections, "\n\n")
	final = capText(final, m.cfg.MaxFinalChars)
	if strings.TrimSpace(final) == "" {
		return Output{}, ErrNoUsableContext
	}

	out.FinalContext = final
	return out, nil
}

func renderSection(title, content string) string {
	return title + ":\n" + strings.TrimSpace(content)
}

func normalize(s string) string {
	return strings.TrimSpace(s)
}

func capText(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return strings.TrimSpace(s[:n])
}

func compressGeneric(s string, maxChars int) string {
	oneLine := strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
	return capText(oneLine, maxChars)
}

var importantLogPattern = regexp.MustCompile(`(?i)(error|fail|panic|timeout|denied|refused|exception|status [45]\d{2})`)

func compressLogs(logs string, maxChars int) string {
	lines := strings.Split(logs, "\n")
	var picked []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if importantLogPattern.MatchString(line) {
			picked = append(picked, line)
		}
	}
	if len(picked) == 0 {
		for i := len(lines) - 1; i >= 0 && len(picked) < 12; i-- {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			picked = append([]string{line}, picked...)
		}
	}
	if len(picked) == 0 {
		return ""
	}
	return capText(strings.Join(picked, "\n"), maxChars)
}
