package gophkeeper

import (
	"context"
	"errors"

	"github.com/hashicorp/go-multierror"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Alena-Kurushkina/gophkeeper/internal/crypter"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"github.com/Alena-Kurushkina/gophkeeper/internal/storage"
)

type Credentialer interface {	
	SaveUserCredentials(ctx context.Context, userID primitive.Binary, encrypted []byte, metaInfo string, marking string) error
	GetUserCredentials(ctx context.Context, userID primitive.Binary) ([]*storage.CredentialsDocument, error)
}

type CredentialsCore struct {
	db Credentialer
}

func NewCredentialsCore(db Credentialer) CredentialsCore{
	return CredentialsCore{
		db: db,
	}
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

func(c *CredentialsCore) SaveCredentials(ctx context.Context, userID uuid.UUID, in CredentialsStorage) (error){
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

func(c *CredentialsCore) GetAllUserCredentials(ctx context.Context, userID uuid.UUID, password string) (AllUserCredentials,error){
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