package usecase

import (
	"context"
	"io"

	"github.com/dnonakolesax/noted-auth/internal/httpclient"
)

type SessionUsecase struct {
	HTTPClient *httpclient.HTTPClient
}

func NewSessionUsecase(httpClient *httpclient.HTTPClient) *SessionUsecase {
	return &SessionUsecase{
		HTTPClient: httpClient,
	}
}

func (su *SessionUsecase) Get(token string) ([]byte, error) {
	sessionsResponse, err := su.HTTPClient.Get(context.TODO(), token)
	defer func() {
		_ = sessionsResponse.Body.Close()
	}()

	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(sessionsResponse.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (su *SessionUsecase) Delete(token string, id string) error {
	deleteResponse, err := su.HTTPClient.Delete(context.TODO(), token, id)
	defer func() {
		_ = deleteResponse.Body.Close()
	}()

	if err != nil {
		return err
	}

	return nil
}
