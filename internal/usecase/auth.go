package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"time"

	"github.com/mailru/easyjson"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/errorvals"
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

	codeVerifier, err := rnd.GenRandomString(ac.kcConfig.CodeVerifierLength)

	if err != nil {
		ac.logger.ErrorContext(ctx, "Failed to create crypto-random string (code verifier)",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return "", err
	}
	b64cv := base64.URLEncoding.EncodeToString(codeVerifier)

	encodedState := base64.URLEncoding.EncodeToString(state)

	hasher := sha256.New()
	_, err = hasher.Write([]byte(b64cv))

	if err != nil {
		ac.logger.ErrorContext(ctx, "Failed to hash code verifier",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return "", err
	}
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	err = ac.repos[0].SetState(ctx, encodedState, returnURL, ac.authLifetime)

	if err != nil {
		ac.logger.ErrorContext(ctx, "Failed to set state",
			slog.String(consts.ErrorLoggerKey, err.Error()))
	}

	err = ac.repos[0].SetState(ctx, encodedState+":code_verifier", b64cv, ac.authLifetime)

	if err != nil {
		ac.logger.ErrorContext(ctx, "Failed to set code verifier",
			slog.String(consts.ErrorLoggerKey, err.Error()))
	}

	go func() {
		for i := 1; i < len(ac.repos); i++ {
			err = ac.repos[1].SetState(ctx, encodedState, returnURL, ac.authLifetime)

			if err != nil {
				ac.logger.ErrorContext(ctx, "Failed to set state",
					slog.String(consts.ErrorLoggerKey, err.Error()))
			}

			err = ac.repos[0].SetState(ctx, encodedState+":code_verifier", returnURL, ac.authLifetime)

			if err != nil {
				ac.logger.ErrorContext(ctx, "Failed to set code verifier",
					slog.String(consts.ErrorLoggerKey, err.Error()))
			}
		}
	}()

	data := url.Values{}
	data.Set("client_id", ac.kcConfig.ClientID)
	data.Set("redirect_uri", ac.kcConfig.RedirectURI)
	data.Set("state", encodedState)
	data.Set("response_type", "code")
	data.Set("code_challenge", sha)
	data.Set("code_challenge_method", "S256")
	ac.logger.InfoContext(ctx, data.Encode())
	link := fmt.Sprintf("%s%s?%s", ac.kcConfig.RealmAddress, ac.kcConfig.AuthEndpoint, data.Encode())
	ac.logger.DebugContext(ctx, "Created auth link", slog.String("Link", link))

	return link, nil
}

func (ac *AuthUsecase) GetToken(ctx context.Context, state string, code string) (model.TokenDTO, error) {
	var returnURL string
	var codeVerifier string
	for _, repo := range ac.repos {
		var err error
		returnURL, err = repo.GetState(ctx, state)

		if err != nil && !errors.Is(err, errorvals.ErrObjectNotFoundInRepoError) {
			ac.logger.ErrorContext(ctx, "Failed to get state", slog.String(consts.ErrorLoggerKey, err.Error()))
			return model.TokenDTO{}, err
		}

		codeVerifier, err = repo.GetState(ctx, state+":code_verifier")

		if err != nil && !errors.Is(err, errorvals.ErrObjectNotFoundInRepoError) {
			ac.logger.ErrorContext(ctx, "Failed to get code verifier", slog.String(consts.ErrorLoggerKey, err.Error()))
			return model.TokenDTO{}, err
		}
	}

	if returnURL == "" {
		ac.logger.WarnContext(ctx, "Return URL not found")
		ac.logger.DebugContext(ctx, "", slog.String("state", state))
		return model.TokenDTO{}, errors.New("return URL not found")
	}
	if codeVerifier == "" {
		ac.logger.WarnContext(ctx, "Code verifier not found")
		ac.logger.DebugContext(ctx, "", slog.String("state", state))
		return model.TokenDTO{}, errors.New("code verifier not found")
	}

	postState, err := rnd.GenRandomString(ac.kcConfig.StateLength)

	if err != nil {
		ac.logger.ErrorContext(ctx, "Failed to create crypto-random string (state)",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return model.TokenDTO{}, err
	}

	encodedState := base64.URLEncoding.EncodeToString(postState)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", ac.kcConfig.ClientID)
	data.Set("client_secret", ac.kcConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", ac.kcConfig.RedirectURI)
	data.Set("state", encodedState)
	data.Set("code_verifier", codeVerifier)

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
	err = easyjson.Unmarshal(body, &dto)
	if err != nil {
		ac.logger.ErrorContext(ctx, "Failed to unmarshal token-post response body",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		ac.logger.DebugContext(ctx, "", "body", body)
		return model.TokenDTO{}, err
	}

	if encodedState != dto.SessionState {
		ac.logger.ErrorContext(ctx, "State mismatch")
		ac.logger.DebugContext(ctx, "", "expected", encodedState, "actual", dto.SessionState)
		return model.TokenDTO{}, errors.New("state mismatch")
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
