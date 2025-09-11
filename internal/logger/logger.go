package logger

import (
	"fmt"
	"log/slog"
	"os"
)

func NewLogger(cfgLogLevel string, addSource bool) *slog.Logger {
	commitHash, ok := os.LookupEnv("CI_COMMIT_HASH")

	if !ok {
		commitHash = "#UNDEFINED"
	}

	machineID, ok := os.LookupEnv("MACHINE_IDENTIFIER")

	if !ok {
		panic("Machine ID undefined")
	}

	var logLevel slog.Level

	switch cfgLogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		panic(fmt.Sprintf("Unknown log level: %s. Known levels: debug, info, warn, error", logLevel))
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: addSource,
		Level:     logLevel,
	})
	logger := slog.New(handler).With(
		slog.Group("exe info",
			slog.Int("pid", os.Getpid()),
			slog.String("commit hash", commitHash),
			slog.String("machine id", machineID),
		),
	)

	return logger
}
