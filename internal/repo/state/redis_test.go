package state

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
	dbredis "github.com/dnonakolesax/noted-auth/internal/db/redis"
)

func TestRedisStateRepo_SetThenGet_OK(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	host, portStr, err := net.SplitHostPort(mr.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	cfg := &configs.RedisConfig{
		Address:        host,
		Port:           port,
		Password:       "",
		RequestTimeout: 500 * time.Millisecond,
	}

	alive := &atomic.Bool{}
	vaultCh := make(chan string)
	t.Cleanup(func() { close(vaultCh) })

	client, err := dbredis.NewClient(cfg, alive, logger, vaultCh)
	require.NoError(t, err)

	repo := NewRedisStateRepo(client, logger)

	ctx := context.Background()
	require.NoError(t, repo.SetState(ctx, "s1", "https://example.com/cb", 2*time.Second))

	got, err := repo.GetState(ctx, "s1")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/cb", got)
}
