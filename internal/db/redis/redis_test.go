package redis

import (
	"context"
	"io"
	"log/slog"
	"net"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/errorvals"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestClient_Get_NotFoundReturnsRepoNotFound(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	host, portStr, err := net.SplitHostPort(mr.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	cfg := &configs.RedisConfig{Address: host, Port: port, Password: "", RequestTimeout: 500 * time.Millisecond}
	alive := &atomic.Bool{}
	vaultCh := make(chan string)
	t.Cleanup(func() { close(vaultCh) })

	c, err := NewClient(cfg, alive, newTestLogger(), vaultCh)
	require.NoError(t, err)

	_, getErr := c.Get(context.Background(), "missing")
	require.ErrorIs(t, getErr, errorvals.ErrObjectNotFoundInRepoError)

	// redis.Nil не должен “ронять” Alive
	require.True(t, alive.Load(), "alive should stay true on redis.Nil")
}

func TestClient_SetThenGet_OK(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	host, portStr, err := net.SplitHostPort(mr.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	cfg := &configs.RedisConfig{Address: host, Port: port, Password: "", RequestTimeout: 500 * time.Millisecond}
	alive := &atomic.Bool{}
	vaultCh := make(chan string)
	t.Cleanup(func() { close(vaultCh) })

	c, err := NewClient(cfg, alive, newTestLogger(), vaultCh)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, c.Set(ctx, "k", "v", 2*time.Second))

	got, err := c.Get(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, "v", got)
	require.True(t, alive.Load())
}

func TestClient_NewClient_FailsOnWrongPassword(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	mr.RequireAuth("pass1")

	host, portStr, err := net.SplitHostPort(mr.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	cfg := &configs.RedisConfig{Address: host, Port: port, Password: "wrong", RequestTimeout: 500 * time.Millisecond}
	alive := &atomic.Bool{}
	vaultCh := make(chan string)
	t.Cleanup(func() { close(vaultCh) })

	// NewClient пингует => должен упасть
	_, err = NewClient(cfg, alive, newTestLogger(), vaultCh)
	require.Error(t, err)
}

func TestClient_MonitorVault_RotatesPasswordAndReconnects(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	mr.RequireAuth("pass1")

	host, portStr, err := net.SplitHostPort(mr.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	cfg := &configs.RedisConfig{Address: host, Port: port, Password: "pass1", RequestTimeout: 500 * time.Millisecond}
	alive := &atomic.Bool{}
	vaultCh := make(chan string, 1)
	t.Cleanup(func() { close(vaultCh) })

	c, err := NewClient(cfg, alive, newTestLogger(), vaultCh)
	require.NoError(t, err)

	// “ротация” на сервере
	mr.RequireAuth("pass2")

	// отправляем новый пароль, MonitorVault должен переподнять conn
	vaultCh <- "pass2"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	deadline := time.Now().Add(2 * time.Second)
	for {
		err = c.Set(ctx, "k2", "v2", 2*time.Second)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected reconnect after vault rotation, last error: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	got, err := c.Get(ctx, "k2")
	require.NoError(t, err)
	require.Equal(t, "v2", got)
}
