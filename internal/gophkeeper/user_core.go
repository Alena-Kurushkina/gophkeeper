package gophkeeper

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Alena-Kurushkina/gophkeeper/internal/authenticator"
	"github.com/Alena-Kurushkina/gophkeeper/internal/crypter"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"github.com/Alena-Kurushkina/gophkeeper/internal/storage"
)

type Userer interface {
	AddUser(ctx context.Context, binaryUUID primitive.Binary, login string, password string) error
	GetUserPassword(ctx context.Context, login string) (*storage.User, error)
}

type UserCore struct {
	db Userer
	hlp authenticator.TokenHelper
}

func NewUserCore(db Userer, token []byte) UserCore{
	return UserCore{
		db: db,
		hlp: *authenticator.NewTokenHelper(token),
	}
}

type Credentials struct {
	Login string
	Password string
}

func(c *UserCore) UserRegister(ctx context.Context, cr Credentials) (string, error)  {
	// генерация id пользователя
	userID := uuid.NewV4()

	token, err:=c.hlp.BuildJWTString(userID)
	if err!=nil{
		return "", err
	}

	hashedPassword, err:=crypter.HashPassword(cr.Password)
	if err!=nil{
		return "", err
	}

	// добавляем пользователя в базу
	binaryUUID:=UUIDToBinary(userID)
	err = c.db.AddUser(ctx, binaryUUID, cr.Login, hashedPassword )
	if err!=nil{
		return "", err
	}

	return token, nil
}

func(c *UserCore) UserLogin(ctx context.Context, cr Credentials) (string, error) {
	user,err := c.db.GetUserPassword(ctx, cr.Login)
	if err!=nil{
		return "", err
	}

	check:=crypter.CompareHashPassword(user.PasswordHash, cr.Password)
 	if !check{
		return "", gopherror.ErrUnauthorized
	}

	uuid, err:=BinaryToUUID(user.UserID)
	if err!=nil{
		return "", err
	}

	token, err:=c.hlp.BuildJWTString(uuid)
	if err!=nil{
		return "", err
	}

	return token, nil
}