package config

import (
	"fmt"
	"os"

	"github.com/jack/yapper/go-note/internal/logging"
)

// Config captures CLI/env derived runtime options.
type Config struct {
	VaultPath string
	LogLevel  logging.Level
}

// Load reads environment variables and CLI args to produce a Config.
func Load(args []string) (Config, error) {
	return FromSources(os.Getenv("NOTE_VAULT_PATH"), os.Getenv("NOTE_DAEMON_LOG"), args)
}

// FromSources is injectable for tests, matching the Rust implementation strategy.
func FromSources(vaultEnv, logEnv string, args []string) (Config, error) {
	cfg := Config{
		VaultPath: ".",
		LogLevel:  logging.LevelInfo,
	}
	if vaultEnv != "" {
		cfg.VaultPath = vaultEnv
	}
	if logEnv != "" {
		level, err := logging.ParseLevel(logEnv)
		if err != nil {
			return cfg, err
		}
		cfg.LogLevel = level
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--vault", "-v":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("--vault expects a following path")
			}
			cfg.VaultPath = args[i]
		case "--log-level", "-l":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("--log-level expects a value")
			}
			level, err := logging.ParseLevel(args[i])
			if err != nil {
				return cfg, err
			}
			cfg.LogLevel = level
		case "--help", "-h":
			return cfg, fmt.Errorf("usage: %s", Usage())
		default:
			return cfg, fmt.Errorf("unrecognized argument %q. Usage: %s", args[i], Usage())
		}
	}

	return cfg, nil
}

// Usage returns the CLI usage text.
func Usage() string {
	return "note-daemon [--vault PATH] [--log-level error|warn|info|debug]"
}
