package user

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/delivery/user/v1/proto"
)

type UserServer struct {
	proto.UnimplementedUserServiceServer
	userUsecase UserUsecase
}

func (us *UserServer) GetUserCtx(ctx context.Context, req *proto.UserId) (*proto.UserInfo, error) {
	user, err := us.userUsecase.Get(req.Uuid)
	println(ctx.Value("abc"))

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

func NewUserService(userUsecase UserUsecase) *UserServer {
	return &UserServer{
		userUsecase: userUsecase,
	}
}
