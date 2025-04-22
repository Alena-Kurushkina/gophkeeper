package gophkeeper

import (
	"context"

	"github.com/Alena-Kurushkina/gophkeeper/internal/storage"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

type GridFSer interface {
	CreateGridFSStream(userID string, filename, metainfo string, size int64) (*gridfs.UploadStream, error)
	GetFileMetadata(ctx context.Context, userID, filename string) (*storage.FileInfo, error)
	CreateDownloadStream(fileID primitive.ObjectID)(*gridfs.DownloadStream, error)
}

type GridFSCore struct {
	db GridFSer
}

func NewGridFSCore(db GridFSer) GridFSCore{
	return GridFSCore{
		db: db,
	}
}

type FileInfo struct {
	FileName string
	ID primitive.ObjectID
	Metadata string
	Size int64
}

func(c *GridFSCore) CreateGridFSStream(userID, filename, metainfo string, size int64)(*gridfs.UploadStream, error){
	return c.db.CreateGridFSStream(userID, filename, metainfo, size)
}

func(c *GridFSCore) CreateDownloadStream(fileID primitive.ObjectID)(*gridfs.DownloadStream, error){
	return c.db.CreateDownloadStream(fileID)
}

func(c *GridFSCore) GetFileMetadata(ctx context.Context, userID uuid.UUID, filename string) (*FileInfo, error) {
	fInfo, err:=c.db.GetFileMetadata(ctx, userID.String(), filename)
	if err!=nil{
		return nil, err
	}
	return &FileInfo{
		FileName: filename,
		ID: fInfo.ID,
		Metadata: fInfo.Metadata.MetaInfo,
		Size: fInfo.Length,
	}, nil
}