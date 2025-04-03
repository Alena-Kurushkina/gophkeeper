package api

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gophkeeper"
	pb "github.com/Alena-Kurushkina/gophkeeper/internal/grpc/proto"
)

type IGophKeeperCore interface {
	UserRegister(ctx context.Context, creds gophkeeper.Credentials) (string, error)
}

type Server struct {
    pb.UnimplementedGophkeeperServer
	Core IGophKeeperCore
}

func (s *Server) CheckConnection(ctx context.Context, in *pb.Hello) (*pb.HelloResponse, error) {
    return &pb.HelloResponse{Response: "Secure Hello " + in.Name}, nil
}

func (s *Server) Register(ctx context.Context, in *pb.Credentials) (*pb.Token, error) {
	if len(in.Password) == 0 || len(in.Login) == 0 {
		// неверный формат запроса
		return nil, status.Errorf(codes.InvalidArgument, "Empty login or password")
	}

	// пароль был захэширован на клиенте

	creds:=gophkeeper.Credentials{
		Login: in.Login,
		HashedPassword: in.Password,
	}

	token, err:=s.Core.UserRegister(ctx, creds)
	if err != nil {
		if errors.Is(err, gopherror.ErrLoginAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "Login is already used by another user")			
		} else {
			return nil, err
		}
	}

	return &pb.Token{
		Token: token,
	}, nil
}