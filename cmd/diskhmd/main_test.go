package main

import (
	"errors"
	"flag"
	"reflect"
	"testing"
)

func TestParseCommandArgsUsesDefaultConfigForDaemon(t *testing.T) {
	t.Parallel()

	cmd, err := parseCommandArgs([]string{"daemon"})
	if err != nil {
		t.Fatalf("parseCommandArgs returned error: %v", err)
	}

	if cmd.Name != commandDaemon {
		t.Fatalf("Name = %q, want %q", cmd.Name, commandDaemon)
	}

	if cmd.ConfigPath != defaultConfigPath {
		t.Fatalf("ConfigPath = %q, want %q", cmd.ConfigPath, defaultConfigPath)
	}
}

func TestParseCommandArgsUsesConfigFlagForDaemon(t *testing.T) {
	t.Parallel()

	cmd, err := parseCommandArgs([]string{"daemon", "--config", "/tmp/diskhm.yaml"})
	if err != nil {
		t.Fatalf("parseCommandArgs returned error: %v", err)
	}

	if cmd.ConfigPath != "/tmp/diskhm.yaml" {
		t.Fatalf("ConfigPath = %q, want %q", cmd.ConfigPath, "/tmp/diskhm.yaml")
	}
}

func TestParseCommandArgsParsesServiceCommand(t *testing.T) {
	t.Parallel()

	cmd, err := parseCommandArgs([]string{"start"})
	if err != nil {
		t.Fatalf("parseCommandArgs returned error: %v", err)
	}

	if cmd.Name != commandStart {
		t.Fatalf("Name = %q, want %q", cmd.Name, commandStart)
	}
}

func TestConfigPathFromArgsUsesConfigFlag(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("diskhm", flag.ContinueOnError)
	configPath, err := configPathFromArgs(fs, defaultConfigPath, []string{"--config", "/tmp/diskhm.yaml"})
	if err != nil {
		t.Fatalf("configPathFromArgs returned error: %v", err)
	}

	if configPath != "/tmp/diskhm.yaml" {
		t.Fatalf("configPath = %q, want %q", configPath, "/tmp/diskhm.yaml")
	}
}

func TestRunServiceCommandRequiresRoot(t *testing.T) {
	t.Parallel()

	err := runServiceCommand(cliCommand{Name: commandStart}, commandDeps{
		euid: func() int { return 1000 },
	})
	if err == nil {
		t.Fatal("runServiceCommand returned nil error")
	}
}

func TestRunUninstallRemovesInstalledFiles(t *testing.T) {
	t.Parallel()

	var systemctlCalls [][]string
	var removedPaths []string

	err := runServiceCommand(cliCommand{Name: commandUninstall}, commandDeps{
		euid: func() int { return 0 },
		systemctl: func(args ...string) error {
			systemctlCalls = append(systemctlCalls, append([]string(nil), args...))
			return nil
		},
		removePath: func(path string) error {
			removedPaths = append(removedPaths, path)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("runServiceCommand returned error: %v", err)
	}

	wantSystemctl := [][]string{
		{"disable", "--now", serviceUnitName},
		{"daemon-reload"},
	}
	if !reflect.DeepEqual(systemctlCalls, wantSystemctl) {
		t.Fatalf("systemctlCalls = %#v, want %#v", systemctlCalls, wantSystemctl)
	}

	wantRemoved := []string{
		installedBinaryPath,
		installedServicePath,
		installedConfigDir,
		installedDataDir,
	}
	if !reflect.DeepEqual(removedPaths, wantRemoved) {
		t.Fatalf("removedPaths = %#v, want %#v", removedPaths, wantRemoved)
	}
}

func TestRunServiceCommandPropagatesSystemctlError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	err := runServiceCommand(cliCommand{Name: commandStop}, commandDeps{
		euid: func() int { return 0 },
		systemctl: func(args ...string) error {
			return wantErr
		},
		removePath: func(path string) error {
			return nil
		},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("runServiceCommand error = %v, want %v", err, wantErr)
	}
}
