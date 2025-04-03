package gophkeeper

import (
	"context"

	"github.com/Alena-Kurushkina/gophkeeper/internal/authenticator"
	"github.com/Alena-Kurushkina/gophkeeper/internal/config"
	uuid "github.com/satori/go.uuid"
)

type Storager interface {
	AddUser(ctx context.Context, userID uuid.UUID, login string, password string) error
}

type GophKeeperCore struct {
	db Storager
	cfg *config.Config
}

func NewGophKeeperCore(db Storager, cfg *config.Config) *GophKeeperCore{
	return &GophKeeperCore{
		db: db,
		cfg: cfg,
	}
}

type Credentials struct {
	Login string
	HashedPassword string
}

func(c *GophKeeperCore) UserRegister(ctx context.Context, cr Credentials) (string, error)  {
	// генерация id пользователя
	userID := uuid.NewV4()

	token, err:=authenticator.BuildJWTString(userID, []byte(c.cfg.TokenKey))
	if err!=nil{
		return "", err
	}

	// добавляем пользователя в базу
	err = c.db.AddUser(ctx, userID, cr.Login, cr.HashedPassword)
	if err!=nil{
		return "", err
	}

	return token, nil
}