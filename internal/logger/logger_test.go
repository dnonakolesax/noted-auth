package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dnonakolesax/noted-auth/internal/configs"
)

func TestNewLogger_UsesEnvCommitHashAndPodName(t *testing.T) {
	t.Setenv("CI_COMMIT_HASH", "abc123")
	t.Setenv("POD_NAME", "pod-1")

	tmp := t.TempDir()

	cfg := &configs.LoggerConfig{
		LogDir:         tmp,
		LogLevel:       "info",
		LogAddSource:   false,
		LogMaxFileSize: 1,
		LogMaxBackups:  1,
		LogMaxAge:      1,
	}

	l := NewLogger(cfg, "service")
	require.NotNil(t, l)

	// Smoke: файл создастся при первом логе
	l.Info("hello")

	// проверяем, что лог-файл в tempdir, а не в /var/log
	_, err := os.Stat(filepath.Join(tmp, "service.log"))
	require.NoError(t, err)
}

func TestNewLogger_UnknownLevel_Panics(t *testing.T) {
	cfg := &configs.LoggerConfig{
		LogDir:         t.TempDir(),
		LogLevel:       "nope",
		LogAddSource:   false,
		LogMaxFileSize: 1, LogMaxBackups: 1, LogMaxAge: 1,
	}

	require.Panics(t, func() {
		_ = NewLogger(cfg, "service")
	})
}
