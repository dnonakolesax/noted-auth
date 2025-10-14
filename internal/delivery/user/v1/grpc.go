package user

import (
	"context"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/delivery/user/v1/proto"
)

type Server struct {
	proto.UnimplementedUserServiceServer

	logger      *slog.Logger
	userUsecase usecase
}

func NewUserServer(userUsecase usecase, logger *slog.Logger) *Server {
	return &Server{
		userUsecase: userUsecase,
		logger:      logger,
	}
}

func (us *Server) GetUserCtx(ctx context.Context, req *proto.UserId) (*proto.UserInfo, error) {
	traceID, ok := ctx.Value("ReqId").(string)

	if !ok {
		us.logger.ErrorContext(ctx, "Couldn't cast trace id to string")
	}

	trace := slog.String(consts.TraceLoggerKey, traceID)
	contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
	user, err := us.userUsecase.Get(contex, req.GetUuid())

	if err != nil {
		us.logger.ErrorContext(ctx, "Error getting user")
		return nil, err
	}

	uinfo := &proto.UserInfo{
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	return uinfo, nil
}
