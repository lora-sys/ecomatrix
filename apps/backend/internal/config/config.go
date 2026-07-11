// Package config loads and validates the backend's environment configuration.
// Configuration is fail-fast: missing required vars abort startup.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr     string
	DBDSN        string
	AdminToken   string
	LogLevel     string
	WSHubBuffer  int
	WSHeartbeatS int
}

func must(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func mustInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// Load reads env vars and validates required ones.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:     must("ECOMATRIX_HTTP_ADDR", ":8080"),
		DBDSN:        must("ECOMATRIX_DB_DSN", "postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable"),
		AdminToken:   must("ECOMATRIX_ADMIN_TOKEN", "dev-admin-token"),
		LogLevel:     strings.ToLower(must("ECOMATRIX_LOG_LEVEL", "info")),
		WSHubBuffer:  mustInt("ECOMATRIX_WS_HUB_BUFFER", 64),
		WSHeartbeatS: mustInt("ECOMATRIX_WS_HEARTBEAT_SECONDS", 20),
	}
	if cfg.AdminToken == "" {
		return cfg, fmt.Errorf("ECOMATRIX_ADMIN_TOKEN is required")
	}
	return cfg, nil
}
