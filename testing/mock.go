package logtesting

import (
	"fmt"
	"sync"

	"github.com/paularlott/logger"
)

// MockLogger is a logger implementation that captures log calls for testing
type MockLogger struct {
	mu      sync.RWMutex
	Entries []LogEntry
	attrs   map[string]any
	group   string
}

// LogEntry represents a single log entry
type LogEntry struct {
	Level         string
	Message       string
	KeysAndValues []any
	Attrs         map[string]any
	Group         string
}

// New creates a new MockLogger
func New() *MockLogger {
	return &MockLogger{
		Entries: make([]LogEntry, 0),
		attrs:   make(map[string]any),
	}
}

func (m *MockLogger) log(level string, msg string, keysAndValues ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Copy attrs
	attrs := make(map[string]any, len(m.attrs))
	for k, v := range m.attrs {
		attrs[k] = v
	}

	m.Entries = append(m.Entries, LogEntry{
		Level:         level,
		Message:       msg,
		KeysAndValues: keysAndValues,
		Attrs:         attrs,
		Group:         m.group,
	})
}

func (m *MockLogger) Trace(msg string, keysAndValues ...any) {
	m.log("trace", msg, keysAndValues...)
}

func (m *MockLogger) Debug(msg string, keysAndValues ...any) {
	m.log("debug", msg, keysAndValues...)
}

func (m *MockLogger) Info(msg string, keysAndValues ...any) {
	m.log("info", msg, keysAndValues...)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...any) {
	m.log("warn", msg, keysAndValues...)
}

func (m *MockLogger) Error(msg string, keysAndValues ...any) {
	m.log("error", msg, keysAndValues...)
}

func (m *MockLogger) Fatal(msg string, keysAndValues ...any) {
	m.log("fatal", msg, keysAndValues...)
}

func (m *MockLogger) With(key string, value any) logger.Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()

	newAttrs := make(map[string]any, len(m.attrs)+1)
	for k, v := range m.attrs {
		newAttrs[k] = v
	}
	newAttrs[key] = value

	return &MockLogger{
		Entries: m.Entries, // Share entries slice for assertions
		attrs:   newAttrs,
		group:   m.group,
		mu:      sync.RWMutex{},
	}
}

func (m *MockLogger) WithError(err error) logger.Logger {
	return m.With("error", err.Error())
}

func (m *MockLogger) WithGroup(group string) logger.Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()

	newAttrs := make(map[string]any, len(m.attrs))
	for k, v := range m.attrs {
		newAttrs[k] = v
	}

	return &MockLogger{
		Entries: m.Entries, // Share entries slice for assertions
		attrs:   newAttrs,
		group:   group,
		mu:      sync.RWMutex{},
	}
}

// Reset clears all captured log entries
func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = make([]LogEntry, 0)
}

// GetEntries returns a copy of all log entries (thread-safe)
func (m *MockLogger) GetEntries() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]LogEntry, len(m.Entries))
	copy(entries, m.Entries)
	return entries
}

// HasEntry checks if an entry with the given level and message exists
func (m *MockLogger) HasEntry(level, message string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range m.Entries {
		if entry.Level == level && entry.Message == message {
			return true
		}
	}
	return false
}

// CountEntries returns the number of log entries with the given level
func (m *MockLogger) CountEntries(level string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, entry := range m.Entries {
		if entry.Level == level {
			count++
		}
	}
	return count
}

// LastEntry returns the last log entry, or nil if no entries
func (m *MockLogger) LastEntry() *LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Entries) == 0 {
		return nil
	}
	entry := m.Entries[len(m.Entries)-1]
	return &entry
}

// String returns a human-readable representation of all log entries
func (m *MockLogger) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Entries) == 0 {
		return "No log entries"
	}

	result := fmt.Sprintf("Log entries (%d):\n", len(m.Entries))
	for i, entry := range m.Entries {
		result += fmt.Sprintf("  [%d] %s: %s", i, entry.Level, entry.Message)
		if entry.Group != "" {
			result += fmt.Sprintf(" [group=%s]", entry.Group)
		}
		if len(entry.Attrs) > 0 {
			result += fmt.Sprintf(" attrs=%v", entry.Attrs)
		}
		if len(entry.KeysAndValues) > 0 {
			result += fmt.Sprintf(" kvs=%v", entry.KeysAndValues)
		}
		result += "\n"
	}
	return result
}
