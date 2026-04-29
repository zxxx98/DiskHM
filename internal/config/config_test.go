package config_test

import (
	"path/filepath"
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
