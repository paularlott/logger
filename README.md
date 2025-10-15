# Logger

A minimal, flexible logging interface for Go libraries with multiple implementation backends. This package provides a common logging interface that can be used across all your Go libraries while allowing applications to choose their preferred logging implementation.

## Design Philosophy

- **Interface-based**: Libraries depend only on the `logger.Logger` interface
- **Implementation-agnostic**: Applications choose slog, zerolog, or custom implementations
- **Minimal**: Simple interface with just the essentials
- **Zero-dependency** for the core interface
- **Performance-conscious**: No unnecessary allocations or complexity

## Installation

```bash
go get github.com/paularlott/logger
```

For specific implementations:

```bash
# For slog (standard library, no additional dependencies)
# Already available in Go 1.21+

# For zerolog
go get github.com/rs/zerolog

# For testing
# No additional dependencies needed
```

## The Interface

```go
type Logger interface {
    Trace(msg string, keysAndValues ...any)
    Debug(msg string, keysAndValues ...any)
    Info(msg string, keysAndValues ...any)
    Warn(msg string, keysAndValues ...any)
    Error(msg string, keysAndValues ...any)
    With(key string, value any) Logger
    WithError(err error) Logger
    WithGroup(group string) Logger
}
```

## Available Implementations

### 1. Null Logger (No-op)

Perfect for tests or when logging is disabled:

```go
import "github.com/paularlott/logger"

log := logger.NewNullLogger()
log.Info("this does nothing")
```

### 2. Slog Logger (Standard Library)

Uses Go's standard `log/slog` package with colored console output:

```go
import logslog "github.com/paularlott/logger/slog"

log := logslog.New(logslog.Config{
    Level:  "info",      // trace, debug, info, warn, error
    Format: "console",   // console (colored) or json
    Writer: os.Stdout,   // any io.Writer
})

log.Info("server started", "port", 8080)
// Output: 15:04:05 INF server started port=8080
```

### 3. Zerolog Logger

Uses the popular zerolog package:

```go
import logzerolog "github.com/paularlott/logger/zerolog"

log := logzerolog.New(logzerolog.Config{
    Level:  "debug",
    Format: "console",   // console or json
    Writer: os.Stdout,
})

log.Debug("processing request", "user_id", 123)
```

### 4. Mock Logger (Testing)

Captures log calls for assertions in tests:

```go
import logtesting "github.com/paularlott/logger/testing"

func TestMyFunction(t *testing.T) {
    mock := logtesting.New()

    // Pass to your code
    myFunction(mock)

    // Assert logs
    if !mock.HasEntry("info", "operation complete") {
        t.Error("expected log entry not found")
    }

    if mock.CountEntries("error") > 0 {
        t.Error("unexpected errors:", mock.String())
    }
}
```

## Usage Patterns

### In Libraries

Libraries should accept the `logger.Logger` interface and provide a sensible default:

```go
package mylib

import "github.com/paularlott/logger"

type Service struct {
    log logger.Logger
}

func NewService(log logger.Logger) *Service {
    if log == nil {
        log = logger.NewNullLogger() // Sensible default
    }
    return &Service{
        log: log.WithGroup("mylib"),
    }
}

func (s *Service) DoWork() error {
    s.log.Info("starting work")

    if err := s.process(); err != nil {
        s.log.WithError(err).Error("work failed")
        return err
    }

    s.log.Info("work completed")
    return nil
}
```

### In Applications - Creating a Log Package

Applications should create their own `log` package that wraps the chosen implementation and provides package-level functions:

**Step 1**: Create `internal/log/log.go` in your application:

```go
package log

import (
    "io"
    "os"

    "github.com/paularlott/logger"
    logslog "github.com/paularlott/logger/slog"
)

var defaultLogger logger.Logger

func init() {
    // Initialize with default configuration
    defaultLogger = logslog.New(logslog.Config{
        Level:  "info",
        Format: "console",
        Writer: os.Stdout,
    })
}

// Configure sets up the logger with the given settings
// Call this early in your main() function
func Configure(level, format string, writer io.Writer) {
    if writer == nil {
        writer = os.Stdout
    }

    defaultLogger = logslog.New(logslog.Config{
        Level:  level,
        Format: format,
        Writer: writer,
    })
}

// GetLogger returns the configured logger instance
// Use this when passing to libraries
func GetLogger() logger.Logger {
    return defaultLogger
}

// Package-level convenience functions
func Trace(msg string, keysAndValues ...any) {
    defaultLogger.Trace(msg, keysAndValues...)
}

func Debug(msg string, keysAndValues ...any) {
    defaultLogger.Debug(msg, keysAndValues...)
}

func Info(msg string, keysAndValues ...any) {
    defaultLogger.Info(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...any) {
    defaultLogger.Warn(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...any) {
    defaultLogger.Error(msg, keysAndValues...)
}

func With(key string, value any) logger.Logger {
    return defaultLogger.With(key, value)
}

func WithError(err error) logger.Logger {
    return defaultLogger.WithError(err)
}

func WithGroup(group string) logger.Logger {
    return defaultLogger.WithGroup(group)
}
```

**Step 2**: Use in your application:

```go
package main

import (
    "flag"
    "os"

    "yourapp/internal/log"
    "yourapp/internal/service"
)

func main() {
    level := flag.String("log-level", "info", "Log level (trace|debug|info|warn|error)")
    format := flag.String("log-format", "console", "Log format (console|json)")
    flag.Parse()

    // Configure logging
    log.Configure(*level, *format, os.Stdout)

    log.Info("application starting", "version", "1.0.0")

    // Pass logger to libraries
    svc := service.New(log.GetLogger())

    if err := svc.Run(); err != nil {
        log.WithError(err).Error("service failed")
        os.Exit(1)
    }

    log.Info("application stopped")
}
```

**Step 3**: Use package-level functions throughout your application:

```go
package handlers

import "yourapp/internal/log"

func HandleRequest(w http.ResponseWriter, r *http.Request) {
    log.Info("handling request", "method", r.Method, "path", r.URL.Path)

    // Use With for contextual logging
    reqLog := log.With("request_id", getRequestID(r))
    reqLog.Debug("processing")

    // Error handling
    if err := process(r); err != nil {
        reqLog.WithError(err).Error("processing failed")
        http.Error(w, "Internal error", 500)
        return
    }

    reqLog.Info("request completed")
}
```

### Switching Implementations

To switch from slog to zerolog, just change your `log` package's `Configure` function:

```go
import logzerolog "github.com/paularlott/logger/zerolog"

func Configure(level, format string, writer io.Writer) {
    if writer == nil {
        writer = os.Stdout
    }

    // Changed from logslog.New to logzerolog.New
    defaultLogger = logzerolog.New(logzerolog.Config{
        Level:  level,
        Format: format,
        Writer: writer,
    })
}
```

No other code needs to change!

## Structured Logging

All implementations support structured key-value logging:

```go
// Inline key-value pairs
log.Info("user logged in", "user_id", 123, "username", "john")

// With creates a child logger with persistent fields
userLog := log.With("user_id", 123).With("session", "abc")
userLog.Info("action performed")  // Includes user_id and session
userLog.Debug("another action")   // Includes user_id and session

// WithError for error context
if err := doSomething(); err != nil {
    log.WithError(err).Error("operation failed")
}

// WithGroup for component context
dbLog := log.WithGroup("database")
dbLog.Info("connected", "host", "localhost")
// Output: 15:04:05 INF database: connected host=localhost
```

## Log Levels

- **Trace**: Very detailed diagnostic information
- **Debug**: Detailed information for debugging
- **Info**: General informational messages
- **Warn**: Warning messages for concerning but non-critical issues
- **Error**: Error messages for failures

## Output Formats

### Console (Colored)

Provides human-readable colored output similar to zerolog:

```
15:04:05 INF server started port=8080 version=1.0
15:04:05 DBG connection opened conn_id=123
15:04:05 WRN cache miss key=user:456
15:04:05 ERR request failed error="connection timeout"
```

### JSON

Machine-readable structured logs:

```json
{"time":"2025-10-15T15:04:05Z","level":"info","msg":"server started","port":8080}
{"time":"2025-10-15T15:04:05Z","level":"error","msg":"request failed","error":"timeout"}
```

## Best Practices

### 1. Accept the Interface in Libraries

```go
// Good: Accepts interface
func NewService(log logger.Logger) *Service

// Bad: Depends on specific implementation
func NewService(log *slog.Logger) *Service
```

### 2. Provide Sensible Defaults

```go
func NewService(log logger.Logger) *Service {
    if log == nil {
        log = logger.NewNullLogger()
    }
    // ...
}
```

### 3. Use Groups for Components

```go
func NewDatabase(log logger.Logger) *Database {
    return &Database{
        log: log.WithGroup("database"),
    }
}
```

### 4. Don't Log and Return Errors

```go
// Bad: Logs AND returns
if err != nil {
    log.Error("failed", "error", err)
    return err
}

// Good: Let caller decide
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Good: Log when handling
if err := doWork(); err != nil {
    log.WithError(err).Error("work failed")
    // Handle the error
}
```

### 5. Use Appropriate Log Levels

```go
log.Trace("entering function", "args", args)           // Very detailed
log.Debug("cache miss", "key", key)                    // Debugging info
log.Info("server started", "port", port)               // Important events
log.Warn("deprecated feature used", "feature", name)   // Warnings
log.Error("request failed", "error", err)              // Errors
```

## Testing

Use the mock logger to verify logging behavior:

```go
func TestService(t *testing.T) {
    mock := logtesting.New()
    svc := NewService(mock)

    err := svc.ProcessData("invalid")

    // Verify error was logged
    if !mock.HasEntry("error", "invalid data") {
        t.Error("expected error log")
    }

    // Check attributes
    lastEntry := mock.LastEntry()
    if lastEntry.Attrs["data"] != "invalid" {
        t.Error("expected data attribute")
    }

    // Print all logs for debugging
    t.Log(mock.String())
}
```

## Migration from Other Loggers

### From logrus:

```go
// Before
log.WithFields(log.Fields{"user": 123}).Info("logged in")

// After
log.With("user", 123).Info("logged in")
```

### From zap:

```go
// Before
log.Info("logged in", zap.Int("user", 123))

// After
log.Info("logged in", "user", 123)
```

### From standard log:

```go
// Before
log.Printf("user %d logged in", userID)

// After
log.Info("user logged in", "user_id", userID)
```

## Examples

Complete working examples are available in the `example/` directory:

- **[example/slog/](example/slog/)** - Full application using the slog implementation
- **[example/zerolog/](example/zerolog/)** - Full application using the zerolog implementation

Both examples demonstrate:
- Creating an application-level log package with convenience functions
- Configuring the logger with command-line flags
- Different log levels and output formats (console with colors, JSON)
- Contextual logging with `With()` and `WithGroup()`
- Integration with services that accept the logger interface

See [example/README.md](example/README.md) for details on running the examples and comparing outputs.

## License

See LICENSE.txt

## Contributing

This is a minimal interface by design. New methods should only be added if they're essential for all logging use cases.
