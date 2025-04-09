package server

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Alena-Kurushkina/gophkeeper/internal/authenticator"
	"github.com/Alena-Kurushkina/gophkeeper/internal/config"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gophkeeper"
	"github.com/Alena-Kurushkina/gophkeeper/internal/grpc/api"
	pb "github.com/Alena-Kurushkina/gophkeeper/internal/grpc/proto"
	"github.com/Alena-Kurushkina/gophkeeper/internal/logger"
)

type GophKeeperServer struct {
	Server *grpc.Server
	config *config.Config
	idleConnsClosed chan struct{}
}

func CreateServer(cfg *config.Config, core *gophkeeper.GophKeeperCore, idleConnsChan chan struct{}) *GophKeeperServer {
	// Загрузка TLS-конфигурации
	creds, err:=credentials.NewServerTLSFromFile(cfg.CertPath, cfg.CertKeyPath)
	if err != nil {
		log.Fatal("Failed to load TLS: ", err)
	}

	authenticator.TargetMethods["/gophkeeper.Gophkeeper/Register"]=false
	authenticator.TargetMethods["/gophkeeper.Gophkeeper/Login"]=false

	chain := ChainUnaryInterceptors(
		authenticator.GRPCAuthInterceptor,
		logger.GRPCLogInterceptor,
	)

	streamChain := ChainStreamInterceptors(
        authenticator.GRPCStreamAuthInterceptor,
        logger.GRPCStreamLogInterceptor,
    )

	// Создание gRPC-сервера с TLS
	s := grpc.NewServer( 
		grpc.Creds(creds),
		grpc.UnaryInterceptor(chain),
		grpc.StreamInterceptor(streamChain),
		grpc.MaxRecvMsgSize(100<<20), // 100 MB
	)
	pb.RegisterGophkeeperServer(s, &api.Server{
		Core: core,
	})

	return &GophKeeperServer{
		Server: s,
		config: cfg,
		idleConnsClosed: idleConnsChan,
	}
}

func (s *GophKeeperServer) Run(){
	// Запуск сервера
	lis, err := net.Listen("tcp", s.config.ServerAddress)
	if err != nil {
		logger.Log.Fatalf("failed to listen: %v", err)
	}

	logger.Log.Infof("Server is starting on %s", s.config.ServerAddress)
	if err := s.Server.Serve(lis); err != nil {
		logger.Log.Fatalf("failed to serve: %v", err)
	}
}

// ChainUnaryInterceptors объединяет несколько унарных перехватчиков в один
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Создаем цепочку вызовов
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = func(next grpc.UnaryHandler, interceptor grpc.UnaryServerInterceptor) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return interceptor(ctx, req, info, next)
				}
			}(chain, interceptors[i])
		}
		return chain(ctx, req)
	}
}

func ChainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
    return func(
        srv interface{},
        ss grpc.ServerStream,
        info *grpc.StreamServerInfo,
        handler grpc.StreamHandler,
    ) error {
        chain := handler
        for i := len(interceptors) - 1; i >= 0; i-- {
            chain = func(next grpc.StreamHandler, interceptor grpc.StreamServerInterceptor) grpc.StreamHandler {
                return func(srv interface{}, stream grpc.ServerStream) error {
                    return interceptor(srv, stream, info, next)
                }
            }(chain, interceptors[i])
        }
        return chain(srv, ss)
    }
}