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
	"github.com/dnonakolesax/noted-auth/internal/cryptos"
	"github.com/dnonakolesax/noted-auth/internal/httpclient"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type StateRepo interface {
	SetState(state string, redirectURI string, timeout time.Duration) error
	GetState(state string) (string, error)
}

type AuthUsecase struct {
	authLifetime time.Duration
	kcTimeout    time.Duration
	stateRepo    StateRepo
	kcConfig     configs.KeycloakConfig
	httpClient   *httpclient.HTTPClient
}

func (ac *AuthUsecase) GetAuthLink(returnUrl string) (string, error) {
	state, err := cryptos.GenRandomString(ac.kcConfig.StateLength)

	if err != nil {
		return "", err
	}

	encodedState := base64.URLEncoding.EncodeToString(state)

	err = ac.stateRepo.SetState(encodedState, returnUrl, time.Second*ac.authLifetime)

	if err != nil {
		return "", err
	}

	data := url.Values{}
	data.Set("client_id", ac.kcConfig.ClientId)
	data.Set("redirect_uri", ac.kcConfig.RedirectURI)
	data.Set("state", encodedState)
	data.Set("response_type", "code")
	slog.Info(data.Encode())
	link := fmt.Sprintf("%s%s?%s", ac.kcConfig.RealmAddress, ac.kcConfig.AuthEndpoint, data.Encode())

	return link, nil
}

func (ac *AuthUsecase) GetToken(state string, code string) (model.TokenDTO, error) {
	returnURL, err := ac.stateRepo.GetState(state)

	if err != nil {
		return model.TokenDTO{}, err
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", ac.kcConfig.ClientId)
	data.Set("client_secret", ac.kcConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", ac.kcConfig.RedirectURI)

	pCtx, cancel := context.WithTimeout(context.Background(), time.Second*ac.kcTimeout)
	defer cancel()
	resp, err := ac.httpClient.PostForm(pCtx, data)
	defer resp.Body.Close()

	if err != nil {
		return model.TokenDTO{}, err
	}
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return model.TokenDTO{}, err
	}
	var dto model.TokenDTO
	err = json.Unmarshal(body, &dto)
	if err != nil {
		return model.TokenDTO{}, err
	}
	dto.ReturnURL = returnURL

	return dto, nil
}

func (ac *AuthUsecase) GetLogoutLink() string {
	return fmt.Sprintf("%s/%s?post_logout_redirect_uri=%s", ac.kcConfig.RealmAddress, ac.kcConfig.LogoutEndpoint, ac.kcConfig.PostLogoutRedirectURI)
}

func NewAuthUsecase(authLifetime time.Duration, stateRepo StateRepo, kcConfig configs.KeycloakConfig, httpClient *httpclient.HTTPClient) *AuthUsecase {
	return &AuthUsecase{
		authLifetime: authLifetime,
		stateRepo:    stateRepo,
		kcConfig:     kcConfig,
		httpClient:   httpClient,
	}
}
