package storage

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Login        string           `bson:"login"`
	PasswordHash string           `bson:"password_hash"`
	CreatedAt    time.Time        `bson:"created_at"`
	UserID       primitive.Binary `bson:"_id"`
}

type CredentialsDocument struct {
	UserID      primitive.Binary `bson:"user_id"`
	Marking     string           `bson:"name"`
	Credentials []byte           `bson:"credentials"`
	MetaInfo    string           `bson:"metainfo"`
}

type FileInfo struct {
	ID         primitive.ObjectID `bson:"_id"`
	Length     int64              `bson:"length"`
	UploadDate time.Time          `bson:"uploadDate"`
	Metadata   struct {
		UserID string `bson:"user_id"`
		MetaInfo string `bson:"metainfo"`
	} `bson:"metadata"`
}
