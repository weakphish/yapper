package main

import (
	"bytes"
	"flag"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/yuin/goldmark"
)

var fileDir = "~/.yapper/" // TODO: make this configurable

func runTaskMode(mdFiles []string) {
	var files []string
	for _, v := range mdFiles {
		files = append(files, path.Join(fileDir, v))
	}

	// Read each file into memory, one file per slice
	fileStrings := make([]string, 0, len(mdFiles))
	for _, file := range files {
		content, err := os.ReadFile(file) // the file is inside the local directory
		if err != nil {
			slog.Error("Error reading file", "file", file, "error", err)
			continue
		}
		fileStrings = append(fileStrings, string(content))
		slog.Debug("Read file", "file", file, "data", string(content))
	}

	// Parse files as markdown using goldmark
	buffers := make([]bytes.Buffer, 0, len(fileStrings))
	for _, fileString := range fileStrings {
		var buffer bytes.Buffer
		if err := goldmark.Convert([]byte(fileString), &buffer); err != nil {
			slog.Error("Error converting markdown", "error", err)
			continue
		}
		buffers = append(buffers, buffer)
	}
}

func runTodayMode() {
	slog.Info("Running in today mode")
	// get $EDITOR
	editor := os.Getenv("EDITOR")
	if editor == "" {
		slog.Error("No editor set in environment variable EDITOR")
		return
	}
	slog.Debug("Using editor", "editor", editor)
	// Open today's file in the editor
	today := time.Now()
	todayFile := path.Join(fileDir, today.Format("2006-01-02")+".md")
	cmd := exec.Command(editor, todayFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		slog.Error("Error starting %s: %s", editor, err)
	}

	err = cmd.Wait()
	if err != nil {
		slog.Error("Error waiting for command to finish", "error", err)
	} else {
		slog.Info("Opened today's file in editor", "file", todayFile)
	}
}
func main() {
	// Set up logging with slog
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	// Get CLI input and parse
	var mode string
	flag.CommandLine.StringVar(&mode, "mode", "today", "tasks | today")

	flag.Parse()
	slog.Debug(mode)

	// Get the directory of the user's files
	root := os.DirFS(fileDir)
	mdFiles, err := fs.Glob(root, "*.md")
	if err != nil {
		slog.Error("Error reading directory", "error", err)
	}

	// Switch on input flag - if task mode, parse directory and retrieve tasks
	switch mode {
	case "tasks":
		slog.Info("Running in tasks mode")
		runTaskMode(mdFiles)
	case "today":
		slog.Info("Running in today mode")
		runTodayMode()
	}
}
