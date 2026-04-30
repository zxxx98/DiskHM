package main

import (
	"flag"
	"testing"
)

func TestConfigPathFromArgsDefaultsToSystemPath(t *testing.T) {
	t.Parallel()

	configPath, err := configPathFromArgs(flag.NewFlagSet("diskhmd", flag.ContinueOnError), defaultConfigPath, nil)
	if err != nil {
		t.Fatalf("configPathFromArgs returned error: %v", err)
	}

	if configPath != defaultConfigPath {
		t.Fatalf("configPath = %q, want %q", configPath, defaultConfigPath)
	}
}

func TestConfigPathFromArgsUsesConfigFlag(t *testing.T) {
	t.Parallel()

	configPath, err := configPathFromArgs(flag.NewFlagSet("diskhmd", flag.ContinueOnError), defaultConfigPath, []string{"--config", "/tmp/diskhm.yaml"})
	if err != nil {
		t.Fatalf("configPathFromArgs returned error: %v", err)
	}

	if configPath != "/tmp/diskhm.yaml" {
		t.Fatalf("configPath = %q, want %q", configPath, "/tmp/diskhm.yaml")
	}
}
