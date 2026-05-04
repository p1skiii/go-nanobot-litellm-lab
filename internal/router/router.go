package router

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var ErrNoMatchingModel = errors.New("no matching model")

type ModelProfile struct {
	Alias          string  `yaml:"alias"`
	Quality        float64 `yaml:"quality"`
	Cost           float64 `yaml:"cost"`
	Latency        float64 `yaml:"latency"`
	ContextLimit   int     `yaml:"context_limit"`
	SupportsStream bool    `yaml:"supports_stream"`
}

type Input struct {
	TaskType        string
	ContextChars    int
	StreamRequested bool
	BudgetHint      string
}

type Decision struct {
	ModelAlias    string
	FallbackChain []string
	RouteReason   string
}

type Router struct {
	models    []ModelProfile
	byAlias   map[string]ModelProfile
	defaults  defaultsConfig
	scoring   scoringConfig
	available []string
}

type modelsConfig struct {
	Models []ModelProfile `yaml:"models"`
}

type policiesConfig struct {
	Defaults defaultsConfig `yaml:"defaults"`
	Router   struct {
		Scoring scoringConfig `yaml:"scoring"`
	} `yaml:"router"`
}

type defaultsConfig struct {
	TaskType      string   `yaml:"task_type"`
	ModelAlias    string   `yaml:"model_alias"`
	FallbackChain []string `yaml:"fallback_chain"`
}

type scoringConfig struct {
	QualityWeight float64 `yaml:"quality_weight"`
	CostWeight    float64 `yaml:"cost_weight"`
	LatencyWeight float64 `yaml:"latency_weight"`
}

func NewDefault() *Router {
	models := []ModelProfile{
		{Alias: "code-cheap", Quality: 0.6, Cost: 0.2, Latency: 0.8, ContextLimit: 128000, SupportsStream: true},
		{Alias: "code-smart", Quality: 0.9, Cost: 0.7, Latency: 0.5, ContextLimit: 128000, SupportsStream: true},
	}
	return newRouter(models, defaultsConfig{
		TaskType:      "review_diff",
		ModelAlias:    "code-cheap",
		FallbackChain: []string{"code-smart"},
	}, scoringConfig{
		QualityWeight: 0.5,
		CostWeight:    0.3,
		LatencyWeight: 0.2,
	})
}

func NewFromFiles(modelsPath, policiesPath string) (*Router, error) {
	models, err := readModelsConfig(modelsPath)
	if err != nil {
		return nil, err
	}
	policies, err := readPoliciesConfig(policiesPath)
	if err != nil {
		return nil, err
	}
	return newRouter(models.Models, policies.Defaults, policies.Router.Scoring), nil
}

func (r *Router) Route(input Input) (Decision, error) {
	taskType := strings.TrimSpace(input.TaskType)
	if taskType == "" {
		taskType = strings.TrimSpace(r.defaults.TaskType)
	}
	if taskType == "" {
		taskType = "review_diff"
	}

	var chosen ModelProfile
	var bestScore float64
	found := false

	for _, model := range r.models {
		if input.StreamRequested && !model.SupportsStream {
			continue
		}
		if model.ContextLimit > 0 && input.ContextChars > model.ContextLimit {
			continue
		}

		score := r.baseScore(model) + budgetAdjustment(model, input.BudgetHint) + taskAdjustment(model, taskType)
		if !found || score > bestScore || (score == bestScore && model.Alias < chosen.Alias) {
			found = true
			chosen = model
			bestScore = score
		}
	}

	if !found {
		return Decision{}, fmt.Errorf("%w: task=%s context_chars=%d stream=%t", ErrNoMatchingModel, taskType, input.ContextChars, input.StreamRequested)
	}

	fallbacks := make([]string, 0, len(r.defaults.FallbackChain))
	for _, alias := range r.defaults.FallbackChain {
		alias = strings.TrimSpace(alias)
		if alias == "" || alias == chosen.Alias {
			continue
		}
		if _, ok := r.byAlias[alias]; ok {
			fallbacks = append(fallbacks, alias)
		}
	}

	budget := normalizedBudget(input.BudgetHint)
	if budget == "" {
		budget = "balanced"
	}
	reason := fmt.Sprintf(
		"task=%s context_chars=%d budget=%s selected=%s score=%.3f",
		taskType, input.ContextChars, budget, chosen.Alias, bestScore,
	)

	return Decision{
		ModelAlias:    chosen.Alias,
		FallbackChain: fallbacks,
		RouteReason:   reason,
	}, nil
}

func (r *Router) baseScore(model ModelProfile) float64 {
	return model.Quality*r.scoring.QualityWeight +
		(1-model.Cost)*r.scoring.CostWeight +
		model.Latency*r.scoring.LatencyWeight
}

func taskAdjustment(model ModelProfile, taskType string) float64 {
	switch taskType {
	case "review_diff":
		return model.Quality * 0.2
	default:
		return 0
	}
}

func budgetAdjustment(model ModelProfile, raw string) float64 {
	switch normalizedBudget(raw) {
	case "low":
		return (1-model.Cost)*0.35 - model.Quality*0.15
	case "high_quality":
		return model.Quality*0.35 - model.Cost*0.10
	default:
		return 0
	}
}

func normalizedBudget(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "low", "cheap", "cost":
		return "low"
	case "high", "quality", "high_quality":
		return "high_quality"
	default:
		return ""
	}
}

func newRouter(models []ModelProfile, defaults defaultsConfig, scoring scoringConfig) *Router {
	if scoring.QualityWeight == 0 && scoring.CostWeight == 0 && scoring.LatencyWeight == 0 {
		scoring = scoringConfig{QualityWeight: 0.5, CostWeight: 0.3, LatencyWeight: 0.2}
	}

	byAlias := make(map[string]ModelProfile, len(models))
	available := make([]string, 0, len(models))
	for _, m := range models {
		alias := strings.TrimSpace(m.Alias)
		if alias == "" {
			continue
		}
		m.Alias = alias
		byAlias[alias] = m
		available = append(available, alias)
	}
	sort.Strings(available)

	if defaults.ModelAlias == "" && len(available) > 0 {
		defaults.ModelAlias = available[0]
	}

	return &Router{
		models:    models,
		byAlias:   byAlias,
		defaults:  defaults,
		scoring:   scoring,
		available: available,
	}
}

func readModelsConfig(path string) (modelsConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return modelsConfig{}, fmt.Errorf("read models config: %w", err)
	}
	var cfg modelsConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return modelsConfig{}, fmt.Errorf("parse models config: %w", err)
	}
	if len(cfg.Models) == 0 {
		return modelsConfig{}, fmt.Errorf("models config has no models")
	}
	return cfg, nil
}

func readPoliciesConfig(path string) (policiesConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return policiesConfig{}, fmt.Errorf("read policies config: %w", err)
	}
	var cfg policiesConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return policiesConfig{}, fmt.Errorf("parse policies config: %w", err)
	}
	return cfg, nil
}
