package logger

// Logger is the minimal interface all paularlott/* libraries accept
type Logger interface {
	Trace(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Fatal(msg string, keysAndValues ...any) // Logs and exits with status 1
	With(key string, value any) Logger
	WithError(err error) Logger
	WithGroup(group string) Logger
}
