package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/diskhm/internal/config"
)

func TestLoadReturnsDefaultsWhenFileMissing(t *testing.T) {
	t.Parallel()

	missingPath := filepath.Join(t.TempDir(), "missing.yaml")

	cfg, err := config.Load(missingPath)
	if err != nil {
		t.Fatalf("Load returned error for missing file: %v", err)
	}

	defaults := config.Default()

	if cfg.Server.ListenAddr != defaults.Server.ListenAddr {
		t.Fatalf("Server.ListenAddr = %q, want %q", cfg.Server.ListenAddr, defaults.Server.ListenAddr)
	}

	if cfg.Security.TokenHash != defaults.Security.TokenHash {
		t.Fatalf("Security.TokenHash = %q, want %q", cfg.Security.TokenHash, defaults.Security.TokenHash)
	}

	if cfg.Sleep.QuietGraceSeconds != defaults.Sleep.QuietGraceSeconds {
		t.Fatalf("Sleep.QuietGraceSeconds = %d, want %d", cfg.Sleep.QuietGraceSeconds, defaults.Sleep.QuietGraceSeconds)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	configBody := "server:\n  listen_addr: 127.0.0.1:9790\n  extra: true\n"
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err := config.Load(configPath)
	if err == nil {
		t.Fatal("Load returned nil error for unknown field")
	}

	if !strings.Contains(err.Error(), "field extra not found") {
		t.Fatalf("Load error = %q, want unknown field detail", err)
	}
}

func TestLoadOverlaysPartialConfigOnDefaults(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	configBody := "server:\n  listen_addr: 0.0.0.0:9790\n"
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	defaults := config.Default()

	if cfg.Server.ListenAddr != "0.0.0.0:9790" {
		t.Fatalf("Server.ListenAddr = %q, want %q", cfg.Server.ListenAddr, "0.0.0.0:9790")
	}

	if cfg.Security.TokenHash != defaults.Security.TokenHash {
		t.Fatalf("Security.TokenHash = %q, want %q", cfg.Security.TokenHash, defaults.Security.TokenHash)
	}

	if cfg.Sleep.QuietGraceSeconds != defaults.Sleep.QuietGraceSeconds {
		t.Fatalf("Sleep.QuietGraceSeconds = %d, want %d", cfg.Sleep.QuietGraceSeconds, defaults.Sleep.QuietGraceSeconds)
	}
}
