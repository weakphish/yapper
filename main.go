package main

import (
	"log/slog"
	"os"

	"github.com/weakphish/yapper/cmd"
	"github.com/weakphish/yapper/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})) // TODO: configure logger with options and to file
	slog.SetDefault(logger)

	config.InitConfig()
	cmd.Execute()
}
