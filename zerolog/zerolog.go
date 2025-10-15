package logzerolog

import (
	"io"
	"os"
	"strings"

	"github.com/paularlott/logger"
	"github.com/rs/zerolog"
)

// ZerologLogger wraps zerolog.Logger to implement the logger.Logger interface
type ZerologLogger struct {
	logger zerolog.Logger
}

// Config for creating a new ZerologLogger
type Config struct {
	Level  string    // "trace", "debug", "info", "warn", "error"
	Format string    // "console" or "json"
	Writer io.Writer // Output writer, defaults to os.Stdout
}

// New creates a new ZerologLogger with the given configuration
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

	var zlog zerolog.Logger

	// Configure output format
	if cfg.Format == "console" {
		output := zerolog.ConsoleWriter{
			Out:        cfg.Writer,
			TimeFormat: "02 Jan 06 15:04 MST",
		}
		zlog = zerolog.New(output).With().Timestamp().Logger()
	} else {
		zlog = zerolog.New(cfg.Writer).With().Timestamp().Logger()
	}

	// Set log level
	level := parseLevel(cfg.Level)
	zlog = zlog.Level(level)

	return &ZerologLogger{
		logger: zlog,
	}
}

func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

func (l *ZerologLogger) Trace(msg string, keysAndValues ...any) {
	l.log(l.logger.Trace(), msg, keysAndValues...)
}

func (l *ZerologLogger) Debug(msg string, keysAndValues ...any) {
	l.log(l.logger.Debug(), msg, keysAndValues...)
}

func (l *ZerologLogger) Info(msg string, keysAndValues ...any) {
	l.log(l.logger.Info(), msg, keysAndValues...)
}

func (l *ZerologLogger) Warn(msg string, keysAndValues ...any) {
	l.log(l.logger.Warn(), msg, keysAndValues...)
}

func (l *ZerologLogger) Error(msg string, keysAndValues ...any) {
	l.log(l.logger.Error(), msg, keysAndValues...)
}

func (l *ZerologLogger) log(event *zerolog.Event, msg string, keysAndValues ...any) {
	// Add key-value pairs
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string)
			if !ok {
				continue
			}
			event.Interface(key, keysAndValues[i+1])
		}
	}
	event.Msg(msg)
}

func (l *ZerologLogger) With(key string, value any) logger.Logger {
	return &ZerologLogger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

func (l *ZerologLogger) WithError(err error) logger.Logger {
	return &ZerologLogger{
		logger: l.logger.With().Err(err).Logger(),
	}
}

func (l *ZerologLogger) WithGroup(group string) logger.Logger {
	return &ZerologLogger{
		logger: l.logger.With().Str("group", group).Logger(),
	}
}
