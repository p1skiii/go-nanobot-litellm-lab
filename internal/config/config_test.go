package config

import (
	"testing"
	"time"
)

func TestLoadUsesDefaultAddr(t *testing.T) {
	t.Setenv("NANOBOT_ADDR", "")
	t.Setenv("LITELLM_BASE_URL", "")
	t.Setenv("LITELLM_MODEL", "")
	t.Setenv("LITELLM_TIMEOUT", "")
	t.Setenv("NANOBOT_MODELS_CONFIG", "")
	t.Setenv("NANOBOT_POLICIES_CONFIG", "")
	t.Setenv("NANOBOT_USAGE_LOG_PATH", "")

	cfg := Load()
	if cfg.Addr != defaultAddr {
		t.Fatalf("addr = %q, want %q", cfg.Addr, defaultAddr)
	}

	if cfg.LiteLLMBaseURL != defaultLiteLLMBaseURL {
		t.Fatalf("litellm base url = %q, want %q", cfg.LiteLLMBaseURL, defaultLiteLLMBaseURL)
	}

	if cfg.LiteLLMModel != defaultLiteLLMModel {
		t.Fatalf("litellm model = %q, want %q", cfg.LiteLLMModel, defaultLiteLLMModel)
	}

	if cfg.LiteLLMTimeout != defaultLiteLLMTimeout {
		t.Fatalf("litellm timeout = %s, want %s", cfg.LiteLLMTimeout, defaultLiteLLMTimeout)
	}
	if cfg.ModelsConfigPath != defaultModelsConfigPath {
		t.Fatalf("models config path = %q, want %q", cfg.ModelsConfigPath, defaultModelsConfigPath)
	}
	if cfg.PoliciesConfigPath != defaultPoliciesConfigPath {
		t.Fatalf("policies config path = %q, want %q", cfg.PoliciesConfigPath, defaultPoliciesConfigPath)
	}
	if cfg.UsageLogPath != defaultUsageLogPath {
		t.Fatalf("usage log path = %q, want %q", cfg.UsageLogPath, defaultUsageLogPath)
	}
}

func TestLoadUsesEnvAddr(t *testing.T) {
	t.Setenv("NANOBOT_ADDR", ":18080")
	t.Setenv("LITELLM_BASE_URL", "http://litellm.test")
	t.Setenv("LITELLM_API_KEY", "sk-test")
	t.Setenv("LITELLM_MODEL", "code-smart")
	t.Setenv("LITELLM_TIMEOUT", "2s")
	t.Setenv("NANOBOT_MODELS_CONFIG", "/tmp/models.yaml")
	t.Setenv("NANOBOT_POLICIES_CONFIG", "/tmp/policies.yaml")
	t.Setenv("NANOBOT_USAGE_LOG_PATH", "/tmp/usage.jsonl")

	cfg := Load()
	if cfg.Addr != ":18080" {
		t.Fatalf("addr = %q, want :18080", cfg.Addr)
	}

	if cfg.LiteLLMBaseURL != "http://litellm.test" {
		t.Fatalf("litellm base url = %q, want http://litellm.test", cfg.LiteLLMBaseURL)
	}

	if cfg.LiteLLMAPIKey != "sk-test" {
		t.Fatalf("litellm api key = %q, want sk-test", cfg.LiteLLMAPIKey)
	}

	if cfg.LiteLLMModel != "code-smart" {
		t.Fatalf("litellm model = %q, want code-smart", cfg.LiteLLMModel)
	}

	if cfg.LiteLLMTimeout != 2*time.Second {
		t.Fatalf("litellm timeout = %s, want 2s", cfg.LiteLLMTimeout)
	}
	if cfg.ModelsConfigPath != "/tmp/models.yaml" {
		t.Fatalf("models config path = %q, want /tmp/models.yaml", cfg.ModelsConfigPath)
	}
	if cfg.PoliciesConfigPath != "/tmp/policies.yaml" {
		t.Fatalf("policies config path = %q, want /tmp/policies.yaml", cfg.PoliciesConfigPath)
	}
	if cfg.UsageLogPath != "/tmp/usage.jsonl" {
		t.Fatalf("usage log path = %q, want /tmp/usage.jsonl", cfg.UsageLogPath)
	}
}
