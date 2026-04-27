package config

import "testing"

func TestLoadUsesDefaultAddr(t *testing.T) {
	t.Setenv("NANOBOT_ADDR", "")

	cfg := Load()
	if cfg.Addr != defaultAddr {
		t.Fatalf("addr = %q, want %q", cfg.Addr, defaultAddr)
	}
}

func TestLoadUsesEnvAddr(t *testing.T) {
	t.Setenv("NANOBOT_ADDR", ":18080")

	cfg := Load()
	if cfg.Addr != ":18080" {
		t.Fatalf("addr = %q, want :18080", cfg.Addr)
	}
}
