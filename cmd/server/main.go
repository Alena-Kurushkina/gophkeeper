package main

import (
	"github.com/Alena-Kurushkina/gophkeeper/internal/config"
	"github.com/Alena-Kurushkina/gophkeeper/internal/logger"
	"github.com/Alena-Kurushkina/gophkeeper/internal/server"
	"github.com/Alena-Kurushkina/gophkeeper/internal/storage"
)

func main() {
    logger.MustInitialize()

    cfg:=config.InitConfig()

    db:=storage.MustCreate(cfg)

    srv:=server.CreateServer(cfg, db)

    srv.Run()
    
}