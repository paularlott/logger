package main

import (
	"os"

	logslog "github.com/paularlott/logger/slog"
)

func main() {
	// Example 1: Default _group
	log1 := logslog.New(logslog.Config{
		Level:  "info",
		Format: "json",
		Writer: os.Stdout,
	})
	log1.WithGroup("service-a").Info("using default _group field")

	// Example 2: Custom @group
	log2 := logslog.New(logslog.Config{
		Level:          "info",
		Format:         "json",
		Writer:         os.Stdout,
		GroupFieldName: "@group",
	})
	log2.WithGroup("service-b").Info("using custom @group field")

	// Example 3: Using 'component' to avoid collisions
	log3 := logslog.New(logslog.Config{
		Level:          "info",
		Format:         "json",
		Writer:         os.Stdout,
		GroupFieldName: "component",
	})
	log3.WithGroup("service-c").Info("using component field")
}
