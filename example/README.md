# Logger Examples

This directory contains example applications demonstrating how to use the logger library with different implementations.

## Examples

### 1. slog Example (`example/slog/`)

Demonstrates using the `logslog` implementation (standard library slog-based).

**Run:**
```bash
cd example/slog
go build .
./slog

# With options:
./slog --log-level debug --log-format console
./slog --log-level trace --log-format json
```

**Features:**
- Uses Go's standard library `log/slog`
- Colored console output
- JSON output support
- All log levels: trace, debug, info, warn, error

### 2. zerolog Example (`example/zerolog/`)

Demonstrates using the `logzerolog` implementation (zerolog-based).

**Run:**
```bash
cd example/zerolog
go build .
./zerolog

# With options:
./zerolog --log-level debug --log-format console
./zerolog --log-level trace --log-format json
```

**Features:**
- Uses `github.com/rs/zerolog`
- Colored console output (zerolog's native console writer)
- JSON output support
- All log levels: trace, debug, info, warn, error

## Command Line Options

Both examples support:
- `--log-level` : Set log level (trace|debug|info|warn|error), default: info
- `--log-format` : Set output format (console|json), default: console

## What the Examples Demonstrate

1. **Creating an application-level log package** with convenience functions
2. **Configuring the logger** at startup with command-line flags
3. **Simple logging** with key-value pairs
4. **Contextual logging** using `With()` to add fields
5. **Error logging** using `WithError()`
6. **Service integration** - passing logger to services via the interface
7. **Grouped logging** using `WithGroup()` to organize logs by component

## Output Comparison

### Console Output (slog)
```
15 Oct 25 12:28 AWST INF application starting version=1.0.0
15 Oct 25 12:28 AWST INF [user-service] user logging in user_id=123 username=john
15 Oct 25 12:28 AWST ERR operation failed error=invalid user ID
```

**Format:** `DD Mon YY HH:MM TZ LVL [group] message key=value`

Groups are shown in cyan brackets `[group]` for clear visual separation.

### Console Output (zerolog)
```
15 Oct 25 12:34 AWST INF application starting (zerolog) version=1.0.0
15 Oct 25 12:34 AWST INF user logging in group=user-service user_id=123 username=john
15 Oct 25 12:34 AWST ERR operation failed error="invalid user ID"
```

**Format:** `DD Mon YY HH:MM TZ LVL message key=value`

Groups are shown as a regular field `group=value` (zerolog's console writer limitation).

### JSON Output (slog)
```json
{"time":"2025-10-15T12:29:13.255196+08:00","level":"INFO","msg":"user logging in","group":"user-service","user_id":123,"username":"john"}
{"time":"2025-10-15T12:29:13.255260+08:00","level":"ERROR","msg":"operation failed","error":"invalid user ID"}
```

Groups appear as `"group":"name"` field.

### JSON Output (zerolog)
```json
{"level":"info","version":"1.0.0","time":"2025-10-15T12:34:26+08:00","message":"application starting (zerolog)"}
{"level":"info","group":"user-service","user_id":123,"username":"john","time":"2025-10-15T12:34:26+08:00","message":"user logging in"}
```

Groups appear as `"group":"name"` field.

## Key Differences

### slog
- **Pros:**
  - No external dependencies (uses stdlib)
  - Groups displayed in brackets `[group]` in console output
  - Built-in to Go 1.21+

- **Cons:**
  - Custom handler needed for colored output
  - Slightly more complex implementation

### zerolog
- **Pros:**
  - Very fast and efficient
  - Rich console writer with colors out of the box
  - Mature and well-tested

- **Cons:**
  - External dependency
  - Groups stored as regular fields (less nested structure)

## Use in Your Application

Copy either example's log package pattern into your application:

```go
// internal/log/log.go
package log

import (
    "github.com/paularlott/logger"
    logslog "github.com/paularlott/logger/slog"
    // OR
    // logzerolog "github.com/paularlott/logger/zerolog"
)

var defaultLogger logger.Logger

func Configure(level, format string) {
    defaultLogger = logslog.New(logslog.Config{
        Level: level, Format: format, Writer: os.Stdout,
    })
}

func Info(msg string, keysAndValues ...any) {
    defaultLogger.Info(msg, keysAndValues...)
}
// ... other convenience functions
```

Then in your main:
```go
import "yourapp/internal/log"

func main() {
    log.Configure("info", "console")
    log.Info("app started", "version", "1.0.0")
}
```

## Testing

When writing libraries, accept the `logger.Logger` interface:

```go
type MyService struct {
    log logger.Logger
}

func NewMyService(log logger.Logger) *MyService {
    if log == nil {
        log = logger.NewNullLogger() // safe default
    }
    return &MyService{log: log}
}
```

In tests, use the mock logger:

```go
import logtesting "github.com/paularlott/logger/testing"

func TestMyService(t *testing.T) {
    mockLog := logtesting.New()
    svc := NewMyService(mockLog)

    svc.DoSomething()

    if !mockLog.HasEntry("info", "something happened") {
        t.Error("expected log entry not found")
    }
}
```
