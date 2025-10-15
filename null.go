package logger

// NullLogger is a no-op logger implementation
type NullLogger struct{}

func NewNullLogger() Logger {
	return NullLogger{}
}

func (NullLogger) Trace(msg string, keysAndValues ...any) {}
func (NullLogger) Debug(msg string, keysAndValues ...any) {}
func (NullLogger) Info(msg string, keysAndValues ...any)  {}
func (NullLogger) Warn(msg string, keysAndValues ...any)  {}
func (NullLogger) Error(msg string, keysAndValues ...any) {}
func (NullLogger) Fatal(msg string, keysAndValues ...any) {} // No-op: does not exit
func (n NullLogger) With(key string, value any) Logger    { return n }
func (n NullLogger) WithError(err error) Logger           { return n }
func (n NullLogger) WithGroup(group string) Logger        { return n }
