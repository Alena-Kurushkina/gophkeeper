package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alena-Kurushkina/gophkeeper/internal/config"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gophkeeper"
	"github.com/Alena-Kurushkina/gophkeeper/internal/logger"
	"github.com/Alena-Kurushkina/gophkeeper/internal/server"
	"github.com/Alena-Kurushkina/gophkeeper/internal/storage"
) 

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
    fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
    
    logger.MustInitialize()
    defer logger.Log.Sync()

    cfg:=config.InitConfig()

    ctx := context.Background()

    db:=storage.MustCreate(ctx, cfg)

    // через канал сообщаем основному потоку, что все сетевые соединения обработаны и закрыты
	idleConnsClosed := make(chan struct{})

    core:=gophkeeper.NewGophKeeperCore(db, cfg)

    srv:=server.CreateServer(cfg, core, idleConnsClosed)

    setupShutdown(idleConnsClosed, core, srv)

    srv.Run()
}

func setupShutdown(idleConnsClosed chan struct{}, core *gophkeeper.GophKeeperCore, rpcServer *server.GophKeeperServer){
	// канал для перенаправления прерываний
	sigint := make(chan os.Signal, 1)
	// регистрируем перенаправление прерываний
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	// запускаем горутину обработки пойманных прерываний
	go func() {
		// читаем из канала прерываний
		<-sigint
		// запускаем процедуру graceful shutdown
		ctx, cancel:=context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		rpcServer.Server.GracefulStop()
		logger.Log.Info("gRPC server was stopped seccussfully")

		// сообщаем основному потоку, что все сетевые соединения обработаны и закрыты
		close(idleConnsClosed)

		core.Shutdown(ctx)
	}()
}