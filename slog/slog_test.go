package logslog

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// A message or value that holds a line break, an escape or invalid UTF-8 is quoted, so it stays on its own line and
// cannot reach the terminal as a control sequence; ordinary text, spaces included, is written as before.
func TestConsoleQuotesControlCharacters(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"unknown user", "unknown user"},
		{"", ""},
		{"x\n2026-10-21 00:00:00 INF forged", `"x\n2026-10-21 00:00:00 INF forged"`},
		{"a\r\nb", `"a\r\nb"`},
		{"\x1b[31mred", `"\x1b[31mred"`},
		{"tab\there", `"tab\there"`},
		{"line\u2028sep", `"line\u2028sep"`},
		{"bad\x9bbyte", `"bad\x9bbyte"`},
		{"café", "café"},
	} {
		var buf bytes.Buffer
		l := slog.New(NewConsoleHandler(&buf, nil, "_group"))
		l.Info(c.in, "k", c.in)
		out := buf.String()
		if strings.Count(out, "\n") != 1 || !strings.HasSuffix(out, "\n") {
			t.Errorf("%q: not one line: %q", c.in, out)
		}
		if !strings.Contains(out, "\033[0m "+c.want) {
			t.Errorf("%q: message written as %q, want %q", c.in, out, c.want)
		}
		if !strings.HasSuffix(out, "k\033[0m="+c.want+"\n") {
			t.Errorf("%q: value written as %q, want %q", c.in, out, c.want)
		}
	}
}
