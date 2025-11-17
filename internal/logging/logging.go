package logging

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Level controls which log statements are emitted.
type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

func (l Level) String() string {
	switch l {
	case LevelError:
		return "error"
	case LevelWarn:
		return "warn"
	case LevelInfo:
		return "info"
	case LevelDebug:
		return "debug"
	default:
		return "info"
	}
}

// ParseLevel converts a string level into the enum.
func ParseLevel(value string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return LevelInfo, nil
	case "error":
		return LevelError, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "info":
		return LevelInfo, nil
	case "debug":
		return LevelDebug, nil
	default:
		return LevelInfo, fmt.Errorf("invalid log level %q", value)
	}
}

var (
	currentLevel = LevelInfo
	levelMu      sync.RWMutex
)

// SetLevel updates the global logger threshold.
func SetLevel(level Level) {
	levelMu.Lock()
	defer levelMu.Unlock()
	currentLevel = level
}

func shouldLog(level Level) bool {
	levelMu.RLock()
	defer levelMu.RUnlock()
	return level <= currentLevel
}

func logf(level Level, format string, args ...interface{}) {
	if !shouldLog(level) {
		return
	}
	levelStr := strings.ToUpper(level.String())
	timestamp := time.Now().Format(time.RFC3339)
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[%s][%s] %s\n", timestamp, levelStr, msg)
}

// Debugf logs at debug level.
func Debugf(format string, args ...interface{}) {
	logf(LevelDebug, format, args...)
}

// Infof logs at info level.
func Infof(format string, args ...interface{}) {
	logf(LevelInfo, format, args...)
}

// Warnf logs at warn level.
func Warnf(format string, args ...interface{}) {
	logf(LevelWarn, format, args...)
}

// Errorf logs at error level.
func Errorf(format string, args ...interface{}) {
	logf(LevelError, format, args...)
}
