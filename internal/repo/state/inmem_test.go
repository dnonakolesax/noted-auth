package state

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/muesli/cache2go"
	"github.com/stretchr/testify/require"

	"github.com/dnonakolesax/noted-auth/internal/errorvals"
)

func TestInMemStateRepo_SetThenGet_OK(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	repo := NewInMemStateRepo(logger)

	ctx := context.Background()
	require.NoError(t, repo.SetState(ctx, "state-1", "https://example.com/cb", 2*time.Second))

	got, err := repo.GetState(ctx, "state-1")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/cb", got)
}

func TestInMemStateRepo_Get_NotFound(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	repo := NewInMemStateRepo(logger)

	_, err := repo.GetState(context.Background(), "missing")
	require.ErrorIs(t, err, errorvals.ErrObjectNotFoundInRepoError)
}

func TestInMemStateRepo_Get_CastError(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	repo := NewInMemStateRepo(logger)

	// кладём не-string, чтобы сломать каст
	repo.client.Add("state-bad", 2*time.Second, 123)

	_, err := repo.GetState(context.Background(), "state-bad")
	require.Error(t, err)
	require.Equal(t, "failed to cast data to string", err.Error())
}

func TestInMemStateRepo_Cache2goErrKeyNotFound_IsReal(t *testing.T) {
	t.Parallel()

	// sanity: cache2go реально возвращает ErrKeyNotFound
	_, err := cache2go.Cache("state").Value("missing")
	require.ErrorIs(t, err, cache2go.ErrKeyNotFound)
}
