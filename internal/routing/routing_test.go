package routing

import (
	"sync/atomic"
	"testing"

	"github.com/fasthttp/router"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

type handlerStub struct {
	calls atomic.Int32
}

func (h *handlerStub) RegisterRoutes(_ *router.Group) {
	h.calls.Add(1)
}

type regHandler struct{}

func (regHandler) RegisterRoutes(g *router.Group) {
	g.GET("/ping", func(_ *fasthttp.RequestCtx) {})
}

func TestNewAPIGroup_PathActuallyMatches(t *testing.T) {
	t.Parallel()

	rr := NewRouter()
	rr.NewAPIGroup("/users", "1", regHandler{})

	r := rr.Router()

	var ctx fasthttp.RequestCtx
	ctx.Request.SetRequestURI("/api/v1/users/ping")
	ctx.Request.Header.SetMethod("GET")

	r.Handler(&ctx)

	require.NotEqual(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
}

func TestNewAPIGroup_CallsAllHandlers(t *testing.T) {
	t.Parallel()

	rr := NewRouter()

	h1 := &handlerStub{}
	h2 := &handlerStub{}

	rr.NewAPIGroup("/users", "1", h1, h2)

	require.Equal(t, int32(1), h1.calls.Load())
	require.Equal(t, int32(1), h2.calls.Load())
}
