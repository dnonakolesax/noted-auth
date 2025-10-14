package healthcheck

import (
	"encoding/json"
	"log/slog"
	"sync/atomic"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type Handler struct {
	redisAlive    *atomic.Bool
	pSQLAlive     *atomic.Bool
	keycloakAlive *atomic.Bool
	vaultAlive    *atomic.Bool
	logger        *slog.Logger
}

func NewHealthCheckHandler(redisAlive *atomic.Bool, pSQLAlive *atomic.Bool, keycloakAlive *atomic.Bool,
	vaultAlive *atomic.Bool, logger *slog.Logger) *Handler {
	return &Handler{
		redisAlive:    redisAlive,
		pSQLAlive:     pSQLAlive,
		keycloakAlive: keycloakAlive,
		vaultAlive:    vaultAlive,
		logger:        logger,
	}
}

func (hh *Handler) Handle(ctx *fasthttp.RequestCtx) {
	r := hh.redisAlive.Load()
	p := hh.pSQLAlive.Load()
	k := hh.keycloakAlive.Load()
	v := hh.vaultAlive.Load()

	dto := model.HealthcheckDTO{
		RedisAlive:    r,
		PostgresAlive: p,
		KeycloakAlive: k,
		VaultAlive:    v,
	}

	bts, err := json.Marshal(dto)

	if err != nil {
		hh.logger.Error("Error marshaling healtcheck: ", slog.String(consts.ErrorLoggerKey, err.Error()))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetBody(bts)

	if r && p && k && v {
		ctx.SetStatusCode(fasthttp.StatusOK)
	} else {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
	}
}

func (hh *Handler) RegisterRoutes(apiGroup *router.Group) {
	group := apiGroup.Group("/users")
	group.GET("", hh.Handle)
}
