package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"strings"
	"encoding/json"
	"io"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/cryptos"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type StateRepo interface {
	SetState(state string, redirectURI string, timeout time.Duration) error 
	GetState(state string) (string, error)
}

type AuthUsecase struct {
	authLifetime time.Duration
	kcTimeout time.Duration
	stateRepo StateRepo
	kcConfig configs.KeycloakConfig
}

func (ac *AuthUsecase) GetAuthLink(returnUrl string) (string, error) {
	state, err := cryptos.GenRandomString(32)

	if err != nil {
		return "", err
	}

	encodedState := base64.URLEncoding.EncodeToString(state)

	err = ac.stateRepo.SetState(encodedState, returnUrl, time.Second*ac.authLifetime)

	if err != nil {
		return "", err
	}

	link := fmt.Sprintf("%s?client_id=%s&response_type=code&state=%s&redirect_uri=%s", ac.kcConfig.RealmAddress + "/auth", ac.kcConfig.ClientId, encodedState, ac.kcConfig.RedirectURI)

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

	req, err := http.NewRequestWithContext(context.TODO(), "POST", ac.kcConfig.RealmAddress + "/token", strings.NewReader(data.Encode()))
	if err != nil {
		return model.TokenDTO{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Выполняем запрос
	client := &http.Client{Timeout: ac.kcTimeout * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return model.TokenDTO{}, err
	}
	defer resp.Body.Close()

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

func NewAuthUsecase(authLifetime time.Duration, stateRepo StateRepo, kcConfig configs.KeycloakConfig) *AuthUsecase {
	return &AuthUsecase{
		authLifetime: authLifetime,
		stateRepo: stateRepo,
		kcConfig: kcConfig,
	}
}
