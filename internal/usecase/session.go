package usecase

import (
	"context"
	"io"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/httpclient"
)

type SessionUsecase struct {
	HTTPClient *httpclient.HTTPClient
	logger     *slog.Logger
}

func NewSessionUsecase(httpClient *httpclient.HTTPClient, logger *slog.Logger) *SessionUsecase {
	return &SessionUsecase{
		HTTPClient: httpClient,
		logger:     logger,
	}
}

func (su *SessionUsecase) Get(ctx context.Context, token string) ([]byte, error) {
	sessionsResponse, err := su.HTTPClient.Get(context.TODO(), token)
	defer func() {
		_ = sessionsResponse.Body.Close()
	}()

	if err != nil {
		su.logger.ErrorContext(ctx, "Error getting sessions", slog.String(consts.ErrorLoggerKey, err.Error()))
		return nil, err
	}

	body, err := io.ReadAll(sessionsResponse.Body)

	if err != nil {
		su.logger.ErrorContext(ctx, "Error reading response body", slog.String(consts.ErrorLoggerKey, err.Error()))
		return nil, err
	}

	return body, nil
}

func (su *SessionUsecase) Delete(ctx context.Context, token string, id string) error {
	deleteResponse, err := su.HTTPClient.Delete(context.TODO(), token, id)
	defer func() {
		_ = deleteResponse.Body.Close()
	}()

	if err != nil {
		su.logger.ErrorContext(ctx, "Error deleting response", slog.String(consts.ErrorLoggerKey, err.Error()))
		return err
	}

	return nil
}
