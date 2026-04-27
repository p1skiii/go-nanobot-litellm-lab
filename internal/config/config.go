package config

import (
	"os"
	"strings"
)

const defaultAddr = ":8080"

type Config struct {
	Addr string
}

func Load() Config {
	addr := strings.TrimSpace(os.Getenv("NANOBOT_ADDR"))
	if addr == "" {
		addr = defaultAddr
	}

	return Config{
		Addr: addr,
	}
}
