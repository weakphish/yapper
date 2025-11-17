package main

import (
	"fmt"
	"os"

	"github.com/jack/yapper/go-note/internal/config"
	"github.com/jack/yapper/go-note/internal/core"
	"github.com/jack/yapper/go-note/internal/logging"
	"github.com/jack/yapper/go-note/internal/server"
)

func main() {
	cfg, err := config.Load(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	logging.SetLevel(cfg.LogLevel)
	logging.Infof("starting note-daemon (vault: %s)", cfg.VaultPath)

	vault := core.NewFileSystemVault(cfg.VaultPath)
	index := core.NewInMemoryIndex()
	parser := core.NewRegexMarkdownParser()
	manager := core.NewVaultIndexManager(vault, index, parser)
	domain := core.NewDomain(manager)

	if err := domain.ReindexAll(); err != nil {
		logging.Errorf("initial vault reindex failed: %v", err)
	} else {
		logging.Infof("vault reindex completed")
	}

	if err := server.Run(domain); err != nil {
		logging.Errorf("server error: %v", err)
		os.Exit(1)
	}
}
