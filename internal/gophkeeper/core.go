package gophkeeper

import (
	"context"

	"github.com/Alena-Kurushkina/gophkeeper/internal/logger"
)

type Shutdowner interface {
	Shutdown(ctx context.Context)
}

type GophKeeperCore struct {
	DB Shutdowner
	UserCore
	CredentialsCore
	GridFSCore
}

// func NewGophKeeperCore(db *storage.Database) *GophKeeperCore{
// 	return &GophKeeperCore{
// 		DB: db,
// 		UserCore: NewUserCore(db),
// 		CredentialsCore: NewCredentialsCore(db),
// 		GridFSCore: NewGridFSCore(db),
// 	}
// }

func (c *GophKeeperCore) Shutdown(ctx context.Context){
	c.DB.Shutdown(ctx)
	logger.Log.Info("GophKeeper core shutdown has finished")
}