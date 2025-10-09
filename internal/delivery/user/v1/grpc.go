package user

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/delivery/user/v1/proto"
)

type Server struct {
	proto.UnimplementedUserServiceServer

	userUsecase usecase
}

func NewUserServer(userUsecase usecase) *Server {
	return &Server{
		userUsecase: userUsecase,
	}
}

func (us *Server) GetUserCtx(ctx context.Context, req *proto.UserId) (*proto.UserInfo, error) {
	user, err := us.userUsecase.Get(req.GetUuid())

	if err != nil {
		slog.Error(fmt.Sprintf("Error getting user: %v, id: %s", err, ctx.Value("ReqId")))
		return nil, err
	}

	uinfo := &proto.UserInfo{
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	return uinfo, nil
}
