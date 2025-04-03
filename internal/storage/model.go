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
