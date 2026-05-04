package config

import (
	"os"
	"strings"
	"time"
)

const defaultAddr = ":8080"
const defaultLiteLLMBaseURL = "http://localhost:4000"
const defaultLiteLLMModel = "code-cheap"
const defaultLiteLLMTimeout = 30 * time.Second
const defaultModelsConfigPath = "configs/models.yaml"
const defaultPoliciesConfigPath = "configs/policies.yaml"
const defaultUsageLogPath = "data/usage.jsonl"

type Config struct {
	Addr               string
	LiteLLMBaseURL     string
	LiteLLMAPIKey      string
	LiteLLMModel       string
	LiteLLMTimeout     time.Duration
	ModelsConfigPath   string
	PoliciesConfigPath string
	UsageLogPath       string
}

func Load() Config {
	addr := strings.TrimSpace(os.Getenv("NANOBOT_ADDR"))
	if addr == "" {
		addr = defaultAddr
	}

	baseURL := strings.TrimSpace(os.Getenv("LITELLM_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultLiteLLMBaseURL
	}

	model := strings.TrimSpace(os.Getenv("LITELLM_MODEL"))
	if model == "" {
		model = defaultLiteLLMModel
	}

	timeout := defaultLiteLLMTimeout
	if raw := strings.TrimSpace(os.Getenv("LITELLM_TIMEOUT")); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	modelsPath := strings.TrimSpace(os.Getenv("NANOBOT_MODELS_CONFIG"))
	if modelsPath == "" {
		modelsPath = defaultModelsConfigPath
	}
	policiesPath := strings.TrimSpace(os.Getenv("NANOBOT_POLICIES_CONFIG"))
	if policiesPath == "" {
		policiesPath = defaultPoliciesConfigPath
	}
	usageLogPath := strings.TrimSpace(os.Getenv("NANOBOT_USAGE_LOG_PATH"))
	if usageLogPath == "" {
		usageLogPath = defaultUsageLogPath
	}

	return Config{
		Addr:               addr,
		LiteLLMBaseURL:     baseURL,
		LiteLLMAPIKey:      strings.TrimSpace(os.Getenv("LITELLM_API_KEY")),
		LiteLLMModel:       model,
		LiteLLMTimeout:     timeout,
		ModelsConfigPath:   modelsPath,
		PoliciesConfigPath: policiesPath,
		UsageLogPath:       usageLogPath,
	}
}
