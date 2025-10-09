package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"
)

func NewLogger(cfgLogLevel string, addSource bool) *slog.Logger {
	commitHash, ok := os.LookupEnv("CI_COMMIT_HASH")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if !ok {
		hash, err := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD").Output()
		if err != nil {
			panic(err)
		}
		commitHash = string(hash)
	}

	podName, ok := os.LookupEnv("POD_NAME")

	if !ok {
		podName = "000000"
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
			slog.String("pod name", podName),
		),
	)

	return logger
}
