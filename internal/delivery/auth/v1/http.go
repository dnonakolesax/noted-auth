package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/errorvals"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type usecase interface {
	GetAuthLink(ctx context.Context, retunURL string) (string, error)
	GetToken(ctx context.Context, state string, token string) (model.TokenDTO, error)
	GetLogoutLink(ctx context.Context) string
}

type Handler struct {
	basicReturnURL string
	requiredPrefix string
	authUsecase    usecase
	logger         *slog.Logger
}

func NewAuthHandler(basicReturnURL string, requiredPrefix string, authUsecase usecase, logger *slog.Logger) *Handler {
	return &Handler{
		basicReturnURL: basicReturnURL,
		requiredPrefix: requiredPrefix,
		authUsecase:    authUsecase,
		logger:         logger,
	}
}

// HandleAuth godoc
// @Summary Handle auth redirect to keycloak
// @Description Generate auth link and redirect to keycloak
// @Tags openid-connect
// @Param return_url query string true "Return url"
// @Success 301
// @Failure 400
// @Failure 500
// @Router /openid-connect/auth [get].
func (ah *Handler) handleAuth(ctx *fasthttp.RequestCtx) {
	trace := string(ctx.Request.Header.Peek(consts.HTTPHeaderXRequestID))
	contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
	returnURL := ctx.QueryArgs().Peek("return_url")
	var returnURLString string
	if returnURL == nil {
		ah.logger.DebugContext(contex, "Return url is empty")
		returnURLString = ah.basicReturnURL
	} else {
		returnURLString = string(returnURL)
	}

	if !strings.HasPrefix(returnURLString, ah.requiredPrefix) {
		ah.logger.WarnContext(contex, "Return url is not allowed", slog.String("return_url", returnURLString))
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	redirectLink, err := ah.authUsecase.GetAuthLink(contex, returnURLString)

	if err != nil {
		ah.logger.ErrorContext(contex, "Error while getting auth link", slog.Any("error", err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Redirect(redirectLink, fasthttp.StatusFound)
}

// HandleToken godoc
// @Summary Handle callback from keycloak
// @Description Receives code and state and returns access token and refresh token
// @Tags openid-connect
// @Param state query string true "State that was sent to keycloak"
// @Param code query string true "Access code from keycloak"
// @Success 301
// @Failure 400
// @Failure 500
// @Router /openid-connect/token [get].
func (ah *Handler) handleToken(ctx *fasthttp.RequestCtx) {
	trace := string(ctx.Request.Header.Peek(consts.HTTPHeaderXRequestID))
	contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
	state := ctx.QueryArgs().Peek("state")

	if state == nil {
		ah.logger.WarnContext(contex, "State is empty")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	code := ctx.QueryArgs().Peek("code")

	if code == nil {
		ah.logger.WarnContext(contex, "Code is empty")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	tokenDTO, err := ah.authUsecase.GetToken(contex, string(state), string(code))

	if err != nil {
		if errors.Is(err, errorvals.ErrObjectNotFoundInRepoError) {
			ah.logger.WarnContext(contex, "State is not found in repo")
			ctx.SetStatusCode(fasthttp.StatusRequestTimeout)
			return
		}
		ah.logger.ErrorContext(contex, "Error while getting token", slog.Any("error", err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	atCookie := fasthttp.Cookie{}
	atCookie.SetKey(consts.ATCookieKey)
	atCookie.SetValue(tokenDTO.AccessToken)
	atCookie.SetMaxAge(tokenDTO.ExpiresIn)
	atCookie.SetHTTPOnly(true)
	atCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)

	rtCookie := fasthttp.Cookie{}
	rtCookie.SetKey(consts.RTCookieKey)
	rtCookie.SetValue(tokenDTO.RefreshToken)
	rtCookie.SetMaxAge(tokenDTO.RefreshExp)
	rtCookie.SetHTTPOnly(true)
	rtCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)

	ctx.Response.Header.SetCookie(&atCookie)
	ctx.Response.Header.SetCookie(&rtCookie)

	ctx.Redirect(tokenDTO.ReturnURL, fasthttp.StatusFound)
}

func (ah *Handler) HandleLogout(ctx *fasthttp.RequestCtx) {
	trace := string(ctx.Request.Header.Peek(consts.HTTPHeaderXRequestID))
	contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
	ctx.Redirect(ah.authUsecase.GetLogoutLink(contex), fasthttp.StatusFound)
}

func (ah *Handler) RegisterRoutes(apiGroup *router.Group) {
	group := apiGroup.Group("/openid-connect")
	group.GET("/auth", ah.handleAuth)
	group.GET("/token", ah.handleToken)
}
