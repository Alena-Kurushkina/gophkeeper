package gophkeeper

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func BinaryToUUID(b primitive.Binary) (uuid.UUID, error){
	if len(b.Data)!=16{
		return uuid.Nil, fmt.Errorf("invalid UUID length: expected 16 bytes, got %d", len(b.Data))
	}
	if b.Subtype != 0x04{ // 0x04 - UUID подтип в MongoDB
		return uuid.Nil, fmt.Errorf("invalid binary subtype for UUID: %v", b.Subtype)
	}

	return uuid.FromBytes(b.Data)
}

func UUIDToBinary(u uuid.UUID) primitive.Binary{
	// Преобразование в Binary
    return primitive.Binary{
        Subtype: 0x04, // UUID subtype
        Data:    u[:],
    }
}