package server

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Alena-Kurushkina/gophkeeper/internal/config"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gophkeeper"
	"github.com/Alena-Kurushkina/gophkeeper/internal/grpc/api"
	pb "github.com/Alena-Kurushkina/gophkeeper/internal/grpc/proto"
	"github.com/Alena-Kurushkina/gophkeeper/internal/logger"
	"github.com/Alena-Kurushkina/gophkeeper/internal/storage"
)

type GophKeeperServer struct {
	server *grpc.Server
	config *config.Config
}

func CreateServer(cfg *config.Config, db *storage.Database) *GophKeeperServer {
	// Загрузка TLS-конфигурации
	creds, err:=credentials.NewServerTLSFromFile(cfg.CertPath, cfg.CertKeyPath)
	if err != nil {
		log.Fatal("Failed to load TLS: ", err)
	}

	// Создание gRPC-сервера с TLS
	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterGophkeeperServer(s, &api.Server{
		Core: gophkeeper.NewGophKeeperCore(db,cfg),
	})

	return &GophKeeperServer{
		server: s,
		config: cfg,
	}
}

func (s *GophKeeperServer) Run(){
	// Запуск сервера
	lis, err := net.Listen("tcp", s.config.ServerAddress)
	if err != nil {
		logger.Log.Fatalf("failed to listen: %v", err)
	}

	logger.Log.Infof("Server is starting on %s", s.config.ServerAddress)
	if err := s.server.Serve(lis); err != nil {
		logger.Log.Fatalf("failed to serve: %v", err)
	}
}