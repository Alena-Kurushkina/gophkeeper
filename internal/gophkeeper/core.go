package gophkeeper

import (
	"context"
	"errors"

	"github.com/hashicorp/go-multierror"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"

	"github.com/Alena-Kurushkina/gophkeeper/internal/authenticator"
	"github.com/Alena-Kurushkina/gophkeeper/internal/config"
	"github.com/Alena-Kurushkina/gophkeeper/internal/crypter"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"github.com/Alena-Kurushkina/gophkeeper/internal/logger"
	"github.com/Alena-Kurushkina/gophkeeper/internal/storage"
)

type Storager interface {
	Shutdown(ctx context.Context)
	AddUser(ctx context.Context, binaryUUID primitive.Binary, login string, password string) error
	GetUserPassword(ctx context.Context, login string) (*storage.User, error)
	SaveUserCredentials(ctx context.Context, userID primitive.Binary, encrypted []byte, metaInfo string, marking string) error
	GetUserCredentials(ctx context.Context, userID primitive.Binary) ([]storage.CredentialsDocument, error)
	CreateGridFSStream(userID string, filename, metainfo string) (*gridfs.UploadStream, error)
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
	Password string
}

type CredentialsInfo struct {
	Creds []byte
	MetaInfo string
	Marking string
}

type CredentialsStorage struct {
	CredentialsInfo
	Password string
}

type AllUserCredentials []*CredentialsInfo

func (c *GophKeeperCore) Shutdown(ctx context.Context){
	c.db.Shutdown(ctx)
	logger.Log.Info("GophKeeper core shutdown has finished")
}

func(c *GophKeeperCore) UserRegister(ctx context.Context, cr Credentials) (string, error)  {
	// генерация id пользователя
	userID := uuid.NewV4()

	token, err:=authenticator.BuildJWTString(userID)
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

func(c *GophKeeperCore) UserLogin(ctx context.Context, cr Credentials) (string, error) {
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

	token, err:=authenticator.BuildJWTString(uuid)
	if err!=nil{
		return "", err
	}

	return token, nil
}

func(c *GophKeeperCore) SaveCredentials(ctx context.Context, userID uuid.UUID, in CredentialsStorage) (error){
	encrypted, err:= crypter.Encrypt(in.Creds, in.Password)
	if err!=nil{
		return err
	}
	binaryUUID:=UUIDToBinary(userID)
	err=c.db.SaveUserCredentials(ctx, binaryUUID, encrypted, in.MetaInfo, in.Marking)
	if err!=nil{
		return err
	}
	return nil
} 

func(c *GophKeeperCore) GetAllUserCredentials(ctx context.Context, userID uuid.UUID, password string) (AllUserCredentials,error){
	binaryUUID:=UUIDToBinary(userID)
	creds, err:=c.db.GetUserCredentials(ctx, binaryUUID)
	if err!=nil{
		return nil, err
	}
	
	var errList error
	out:=AllUserCredentials{}
	for _,item:=range creds{
		decrypted, err:=crypter.Decrypt(item.Credentials, password)
		if err!=nil{			
			if errors.Is(err, gopherror.ErrDecryptAuth){
				errList=multierror.Append(errList, err)
			} else {
				return nil, err
			}
		} else {	
			out=append(out, &CredentialsInfo{
				Creds: decrypted,
				MetaInfo: item.MetaInfo,
				Marking: item.Marking,
			})
		}
	}
	return out, errList
}

func(c *GophKeeperCore) CreateGridFSStream(userID, filename, metainfo string)(*gridfs.UploadStream, error){
	return c.db.CreateGridFSStream(userID, filename, metainfo)
}
