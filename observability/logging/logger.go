// Package logging provides structured JSON logging with enriched context
// for all straitKubegateway components.
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// Level represents log severity.
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Entry is a structured JSON log entry.
type Entry struct {
	Timestamp  string           `json:"timestamp"`
	Level      Level            `json:"level"`
	Message    string           `json:"message"`
	Context    *types.Metadata  `json:"context,omitempty"`
	Error      *types.ErrorInfo `json:"error,omitempty"`
	DurationMs float64          `json:"duration_ms,omitempty"`
}

// Logger outputs structured JSON log entries.
type Logger struct {
	mu     sync.Mutex
	out    io.Writer
	minLvl Level
}

var (
	defaultLogger *Logger
	logOnce       sync.Once
)

// DefaultLogger returns the singleton structured JSON logger.
func DefaultLogger() *Logger {
	logOnce.Do(func() {
		defaultLogger = &Logger{
			out:    os.Stdout,
			minLvl: LevelInfo,
		}
	})
	return defaultLogger
}

// SetOutput changes the output destination.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// SetMinLevel sets the minimum log level.
func (l *Logger) SetMinLevel(lvl Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLvl = lvl
}

// Log writes a structured JSON log entry.
func (l *Logger) Log(lvl Level, msg string, ctx *types.Metadata, errInfo *types.ErrorInfo) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     lvl,
		Message:   msg,
		Context:   ctx,
		Error:     errInfo,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.out, `{"timestamp":"%s","level":"ERROR","message":"failed to marshal log"}`+"\n", entry.Timestamp)
		return
	}
	l.out.Write(append(data, '\n'))
}

// Info logs an informational message.
func (l *Logger) Info(msg string, ctx *types.Metadata) {
	l.Log(LevelInfo, msg, ctx, nil)
}

// Error logs an error message with error details.
func (l *Logger) Error(msg string, errInfo *types.ErrorInfo, ctx *types.Metadata) {
	l.Log(LevelError, msg, ctx, errInfo)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, ctx *types.Metadata) {
	l.Log(LevelDebug, msg, ctx, nil)
}
