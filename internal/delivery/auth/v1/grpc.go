package auth

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/metadata"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	auth "github.com/dnonakolesax/noted-auth/internal/delivery/auth/v1/proto"
)

type Server struct {
	auth.UnimplementedAuthServiceServer

	logger      *slog.Logger
	authUsecase usecase
}

func NewUserServer(authUsecase usecase, logger *slog.Logger) *Server {
	return &Server{
		authUsecase: authUsecase,
		logger:      logger,
	}
}

func (us *Server) AuthUserIDCtx(ctx context.Context, req *auth.UserTokens) (*auth.TokenData, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		us.logger.ErrorContext(ctx, "Couldn't parse request metadata")
		return nil, errors.New("couldn't parse request metadata")
	}
	traceID := md["trace_id"]
	us.logger.DebugContext(ctx, "got request", slog.String("trace", traceID[0]))

	trace := slog.String(consts.TraceLoggerKey, traceID[0])
	contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
	tokenData, err := us.authUsecase.GetUserID(contex, req.GetAuth(), req.GetRefresh())

	if err != nil {
		us.logger.ErrorContext(ctx, "Error getting user", slog.String(consts.ErrorLoggerKey, err.Error()))
		return nil, err
	}

	uinfo := &auth.TokenData{
		ID: tokenData.UserID,
		At: nil,
		Rt: nil,
		It: nil,
	}

	if tokenData.AccessToken != "" && tokenData.RefreshToken != "" {
		uinfo.At = &tokenData.AccessToken
		uinfo.Rt = &tokenData.RefreshToken
		uinfo.It = &tokenData.IDToken
	}

	return uinfo, nil
}
