package explo_test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

// testLogWriter is a generic custom io.Writer that validates log entries as they're written
type testLogWriter[T any] struct {
	actualLogs []T
}

func newTestLogger[T any](logWriter *testLogWriter[T]) *slog.Logger {
	return slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{AddSource: false}))
}

func newTestLogWriter[T any]() *testLogWriter[T] {
	return &testLogWriter[T]{
		actualLogs: make([]T, 0),
	}
}

// Write implements io.Writer and parses JSON log entries
func (w *testLogWriter[T]) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	// Parse the JSON log entry
	var log T
	if err := json.Unmarshal(p, &log); err != nil {
		return 0, fmt.Errorf("failed to unmarshal log entry: %w", err)
	}

	w.actualLogs = append(w.actualLogs, log)
	return len(p), nil
}

// validate checks that all expected logs were received in order
func (w *testLogWriter[T]) assertLogEntries(t *testing.T, expectedLogs []T) {
	for _, expectedLog := range expectedLogs {
		assert.Contains(t, w.actualLogs, expectedLog, "expected log entry were not found in actual logs: %v", expectedLog)
	}
}
