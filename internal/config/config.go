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

type Config struct {
	Addr           string
	LiteLLMBaseURL string
	LiteLLMAPIKey  string
	LiteLLMModel   string
	LiteLLMTimeout time.Duration
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

	return Config{
		Addr:           addr,
		LiteLLMBaseURL: baseURL,
		LiteLLMAPIKey:  strings.TrimSpace(os.Getenv("LITELLM_API_KEY")),
		LiteLLMModel:   model,
		LiteLLMTimeout: timeout,
	}
}
