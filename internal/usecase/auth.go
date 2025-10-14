package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"time"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/httpclient"
	"github.com/dnonakolesax/noted-auth/internal/model"
	"github.com/dnonakolesax/noted-auth/internal/rnd"
)

type StateRepo interface {
	SetState(ctx context.Context, state string, redirectURI string, timeout time.Duration) error
	GetState(ctx context.Context, state string) (string, error)
}

type AuthUsecase struct {
	authLifetime time.Duration
	kcTimeout    time.Duration
	repos        []StateRepo
	kcConfig     configs.KeycloakConfig
	httpClient   *httpclient.HTTPClient
	logger       *slog.Logger
}

func NewAuthUsecase(authLifetime time.Duration, repos []StateRepo, kcConfig configs.KeycloakConfig,
	httpClient *httpclient.HTTPClient, logger *slog.Logger) *AuthUsecase {
	return &AuthUsecase{
		authLifetime: authLifetime,
		repos:        repos,
		kcConfig:     kcConfig,
		httpClient:   httpClient,
		logger:       logger,
	}
}

func (ac *AuthUsecase) GetAuthLink(ctx context.Context, returnURL string) (string, error) {
	state, err := rnd.GenRandomString(ac.kcConfig.StateLength)

	if err != nil {
		ac.logger.ErrorContext(ctx, "Failed to create crypto-random string",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return "", err
	}

	encodedState := base64.URLEncoding.EncodeToString(state)

	err = ac.repos[0].SetState(ctx, encodedState, returnURL, ac.authLifetime)

	go func() {
		for i := 1; i < len(ac.repos); i++ {
			err = ac.repos[1].SetState(ctx, encodedState, returnURL, ac.authLifetime)

			if err != nil {
				ac.logger.ErrorContext(ctx, "Failed to set state",
					slog.String(consts.ErrorLoggerKey, err.Error()))
			}
		}
	}()

	data := url.Values{}
	data.Set("client_id", ac.kcConfig.ClientID)
	data.Set("redirect_uri", ac.kcConfig.RedirectURI)
	data.Set("state", encodedState)
	data.Set("response_type", "code")
	ac.logger.InfoContext(ctx, data.Encode())
	link := fmt.Sprintf("%s%s?%s", ac.kcConfig.RealmAddress, ac.kcConfig.AuthEndpoint, data.Encode())
	ac.logger.DebugContext(ctx, "Created auth link", slog.String("Link", link))

	return link, nil
}

func (ac *AuthUsecase) GetToken(ctx context.Context, state string, code string) (model.TokenDTO, error) {
	var returnURL string
	for _, repo := range ac.repos {
		var err error
		returnURL, err = repo.GetState(ctx, state)

		if err != nil {
			ac.logger.ErrorContext(ctx, "Failed to get state", slog.String(consts.ErrorLoggerKey, err.Error()))
			return model.TokenDTO{}, err
		}
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", ac.kcConfig.ClientID)
	data.Set("client_secret", ac.kcConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", ac.kcConfig.RedirectURI)

	pCtx, cancel := context.WithTimeout(context.Background(), ac.kcTimeout)
	defer cancel()
	resp, err := ac.httpClient.PostForm(pCtx, data)
	defer func() {
		_ = resp.Body.Close()
	}()

	if err != nil {
		ac.logger.ErrorContext(ctx, "Failed to get token", slog.String(consts.ErrorLoggerKey, err.Error()))
		return model.TokenDTO{}, err
	}
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		ac.logger.ErrorContext(ctx, "Failed to read token-post response body",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return model.TokenDTO{}, err
	}
	var dto model.TokenDTO
	err = json.Unmarshal(body, &dto)
	if err != nil {
		ac.logger.ErrorContext(ctx, "Failed to unmarshal token-post response body",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		ac.logger.DebugContext(ctx, "", "body", body)
		return model.TokenDTO{}, err
	}
	dto.ReturnURL = returnURL

	return dto, nil
}

func (ac *AuthUsecase) GetLogoutLink(ctx context.Context) string {
	trace, _ := ctx.Value(consts.TraceContextKey).(slog.Attr)
	link := fmt.Sprintf("%s/%s?post_logout_redirect_uri=%s",
		ac.kcConfig.RealmAddress, ac.kcConfig.LogoutEndpoint, ac.kcConfig.PostLogoutRedirectURI)
	ac.logger.DebugContext(ctx, "Created logout link", slog.String("link", link), trace)
	return link
}
