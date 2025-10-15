package logslog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/paularlott/logger"
)

// Custom slog level for TRACE (below DEBUG which is -4)
const LevelTrace = slog.Level(-8)

// Custom slog level for FATAL (above ERROR which is 8)
const LevelFatal = slog.Level(10)

// SlogLogger wraps slog.Logger to implement the logger.Logger interface
// SlogLogger wraps slog.Logger to implement the logger.Logger interface
type SlogLogger struct {
	logger         *slog.Logger
	groupFieldName string
}

// Config for creating a new SlogLogger
type Config struct {
	Level          string    // "trace", "debug", "info", "warn", "error"
	Format         string    // "console" or "json"
	Writer         io.Writer // Output writer, defaults to os.Stdout
	GroupFieldName string    // Field name for groups, defaults to "_group"
}

// New creates a new SlogLogger with the given configuration
func New(cfg Config) logger.Logger {
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}
	if cfg.Format == "" {
		cfg.Format = "console"
	}
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.GroupFieldName == "" {
		cfg.GroupFieldName = "_group"
	}

	level := parseLevel(cfg.Level)
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Replace custom level names for TRACE and FATAL in JSON output
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				if level == LevelTrace {
					return slog.String(slog.LevelKey, "TRACE")
				} else if level == LevelFatal {
					return slog.String(slog.LevelKey, "FATAL")
				}
			}
			return a
		},
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = &JSONHandler{
			handler: slog.NewJSONHandler(cfg.Writer, opts),
		}
	} else {
		handler = NewConsoleHandler(cfg.Writer, opts, cfg.GroupFieldName)
	}

	return &SlogLogger{
		logger:         slog.New(handler),
		groupFieldName: cfg.GroupFieldName,
	}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return LevelTrace
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "fatal":
		return LevelFatal
	default:
		return slog.LevelInfo
	}
}

func (l *SlogLogger) Trace(msg string, keysAndValues ...any) {
	l.log(LevelTrace, msg, keysAndValues...)
}

func (l *SlogLogger) Debug(msg string, keysAndValues ...any) {
	l.log(slog.LevelDebug, msg, keysAndValues...)
}

func (l *SlogLogger) Info(msg string, keysAndValues ...any) {
	l.log(slog.LevelInfo, msg, keysAndValues...)
}

func (l *SlogLogger) Warn(msg string, keysAndValues ...any) {
	l.log(slog.LevelWarn, msg, keysAndValues...)
}

func (l *SlogLogger) Error(msg string, keysAndValues ...any) {
	l.log(slog.LevelError, msg, keysAndValues...)
}

func (l *SlogLogger) Fatal(msg string, keysAndValues ...any) {
	l.log(LevelFatal, msg, keysAndValues...)
	os.Exit(1)
}

func (l *SlogLogger) log(level slog.Level, msg string, keysAndValues ...any) {
	l.logger.Log(context.Background(), level, msg, keysAndValues...)
}

func (l *SlogLogger) With(key string, value any) logger.Logger {
	return &SlogLogger{
		logger:         l.logger.With(key, value),
		groupFieldName: l.groupFieldName,
	}
}

func (l *SlogLogger) WithError(err error) logger.Logger {
	return &SlogLogger{
		logger:         l.logger.With("error", err),
		groupFieldName: l.groupFieldName,
	}
}

func (l *SlogLogger) WithGroup(group string) logger.Logger {
	return &SlogLogger{
		logger:         l.logger.With(l.groupFieldName, group),
		groupFieldName: l.groupFieldName,
	}
}

// JSONHandler is a wrapper around slog.JSONHandler that properly formats TRACE and FATAL levels
type JSONHandler struct {
	handler slog.Handler
}

func (h *JSONHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *JSONHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

func (h *JSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &JSONHandler{
		handler: h.handler.WithAttrs(attrs),
	}
}

func (h *JSONHandler) WithGroup(name string) slog.Handler {
	return &JSONHandler{
		handler: h.handler.WithGroup(name),
	}
}

// ConsoleHandler is a custom slog handler that outputs colored logs similar to zerolog's console output
type ConsoleHandler struct {
	opts           *slog.HandlerOptions
	writer         io.Writer
	attrs          []slog.Attr
	groups         []string
	groupFieldName string
}

// NewConsoleHandler creates a new console handler with colored output
func NewConsoleHandler(w io.Writer, opts *slog.HandlerOptions, groupFieldName string) *ConsoleHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ConsoleHandler{
		opts:           opts,
		writer:         w,
		attrs:          []slog.Attr{},
		groups:         []string{},
		groupFieldName: groupFieldName,
	}
}

func (h *ConsoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *ConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	var buf strings.Builder

	// Date and time with timezone: "15 Oct 25 12:23 AWST"
	buf.WriteString("\033[90m")
	buf.WriteString(r.Time.Format("02 Jan 06 15:04 MST"))
	buf.WriteString("\033[0m ")

	// Level with color
	levelColor := getLevelColor(r.Level)
	levelStr := getLevelString(r.Level)
	buf.WriteString(levelColor)
	buf.WriteString(levelStr)
	buf.WriteString("\033[0m ")

	// Group in brackets if present (from handler attrs or record attrs)
	var group string

	// Check handler-level attributes first
	for _, attr := range h.attrs {
		if attr.Key == h.groupFieldName {
			group = attr.Value.String()
			break
		}
	}

	// Check record-level attributes if not found
	if group == "" {
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == h.groupFieldName {
				group = a.Value.String()
				return false
			}
			return true
		})
	}

	if group != "" {
		buf.WriteString("\033[36m[")
		buf.WriteString(group)
		buf.WriteString("]\033[0m ")
	}

	// Message
	buf.WriteString(r.Message)

	// Handler-level attributes (skip group field as it's already displayed)
	for _, attr := range h.attrs {
		if attr.Key != h.groupFieldName {
			appendAttr(&buf, attr, h.groups)
		}
	}

	// Record attributes (skip group field as it's already displayed)
	r.Attrs(func(a slog.Attr) bool {
		if a.Key != h.groupFieldName {
			appendAttr(&buf, a, h.groups)
		}
		return true
	})

	buf.WriteString("\n")
	_, err := h.writer.Write([]byte(buf.String()))
	return err
}

func appendAttr(buf *strings.Builder, attr slog.Attr, groups []string) {
	// Handle group nesting
	key := attr.Key
	if len(groups) > 0 {
		key = strings.Join(groups, ".") + "." + key
	}

	// Handle group attributes
	if attr.Value.Kind() == slog.KindGroup {
		for _, groupAttr := range attr.Value.Group() {
			appendAttr(buf, groupAttr, append(groups, attr.Key))
		}
		return
	}

	buf.WriteString(" \033[36m")
	buf.WriteString(key)
	buf.WriteString("\033[0m=")
	buf.WriteString(attr.Value.String())
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &ConsoleHandler{
		opts:           h.opts,
		writer:         h.writer,
		attrs:          newAttrs,
		groups:         h.groups,
		groupFieldName: h.groupFieldName,
	}
}

func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &ConsoleHandler{
		opts:           h.opts,
		writer:         h.writer,
		attrs:          h.attrs,
		groups:         newGroups,
		groupFieldName: h.groupFieldName,
	}
}

// ANSI color codes
func getLevelColor(level slog.Level) string {
	switch level {
	case LevelTrace:
		return "\033[35m" // Magenta
	case slog.LevelDebug:
		return "\033[33m" // Yellow
	case slog.LevelInfo:
		return "\033[32m" // Green
	case slog.LevelWarn:
		return "\033[33m" // Yellow
	case slog.LevelError, LevelFatal:
		return "\033[31m" // Red
	default:
		return "\033[0m" // Reset
	}
}

func getLevelString(level slog.Level) string {
	switch level {
	case LevelTrace:
		return "TRC"
	case slog.LevelDebug:
		return "DBG"
	case slog.LevelInfo:
		return "INF"
	case slog.LevelWarn:
		return "WRN"
	case slog.LevelError:
		return "ERR"
	case LevelFatal:
		return "FTL"
	default:
		return "???"
	}
}
