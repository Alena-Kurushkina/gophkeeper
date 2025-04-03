package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Alena-Kurushkina/gophkeeper/internal/config"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"github.com/Alena-Kurushkina/gophkeeper/internal/logger"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	client *mongo.Client //TODO: нужно ли хранить это в структуре
	database *mongo.Database
}

const (
	collectionUsers = "users"
)

func MustCreate(cfg *config.Config) *Database {
	ctx, cancel:= context.WithTimeout(context.Background(), 10*time.Second)
	//TODO: ???
	defer cancel()

	// Подключение к MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.ConnectionStr))
	if err != nil {
		logger.Log.Fatalf("Failed to connect to DB:",err)
	}

	db:=client.Database(cfg.DBName)

	//TODO: in other function
	// create collections
	colUsers:=db.Collection(collectionUsers)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "login", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = colUsers.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal("Failed to create index:", err)
	}


	return &Database{
		client: client,
		database: db,
	}
}

func (d *Database) Shutdown(){
	d.client.Disconnect(context.TODO())
}

func(d *Database) AddUser(ctx context.Context, userID uuid.UUID, login string, hashedPassword string) error{
	// Преобразование в Binary
    binaryUUID := primitive.Binary{
        Subtype: 0x04, // UUID subtype
        Data:    userID[:],
    }

	newUser := User{
		Login:        login,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now().UTC(),
		UserID: binaryUUID,
	}

	// Вставка документа в MongoDB
	collection := d.database.Collection(collectionUsers)
	_, err := collection.InsertOne(context.Background(), newUser)

	// Обработка ошибок
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return gopherror.ErrLoginAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Для binary UUID
// uuidBytes := uuid.MustParse("a6bb8f0d-7e7a-4a9a-b88d-5b5e5b5e5b5e").Bytes()
// cursor, err := collection.Find(ctx, bson.M{"_id": primitive.Binary{Subtype: 0x04, Data: uuidBytes}})




	// // Получение коллекции
	// collection := client.Database("gophkeeper").Collection("user")

	// // Получение всех документов
	// cursor, err := collection.Find(context.TODO(), bson.D{})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer cursor.Close(context.TODO())

	// // Декодирование результатов в slice мап
	// var results []bson.M
	// if err = cursor.All(context.TODO(), &results); err != nil {
	// 	log.Fatal(err)
	// }

	// // Конвертация в красивый JSON
	// jsonData, err := json.MarshalIndent(results, "", "  ")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Вывод JSON
	// fmt.Println(string(jsonData))