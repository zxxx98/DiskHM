package config

import (
	"bytes"
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Security SecurityConfig `yaml:"security"`
	Sleep    SleepConfig    `yaml:"sleep"`
}

type ServerConfig struct {
	ListenAddr string `yaml:"listen_addr"`
}

type SecurityConfig struct {
	TokenHash string `yaml:"token_hash"`
}

type SleepConfig struct {
	QuietGraceSeconds int `yaml:"quiet_grace_seconds"`
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			ListenAddr: "0.0.0.0:9789",
		},
		Security: SecurityConfig{
			TokenHash: "bootstrap-token-hash-change-me",
		},
		Sleep: SleepConfig{
			QuietGraceSeconds: 10,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}

		return Config{}, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
