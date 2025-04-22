package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Alena-Kurushkina/gophkeeper/internal/config"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"github.com/Alena-Kurushkina/gophkeeper/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	client *mongo.Client //TODO: нужно ли хранить это в структуре
	database *mongo.Database
}

const (
	collectionUsers = "users"
	collectionCredentials = "credentials"
	collectionFiles = "files"
)

func MustCreate(ctx context.Context, cfg *config.Config) *Database {
	ctxTO, cancel:= context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// connect to mogodb
	client, err := mongo.Connect(ctxTO, options.Client().ApplyURI(cfg.ConnectionStr))
	if err != nil {
		logger.Log.Fatalf("Failed to connect to DB:",err)
	}

	db:=client.Database(cfg.DBName)

	// create collections
	err=createCollections(db)
	if err!=nil{
		logger.Log.Fatalf("error creating collections", err)
	}

	return &Database{
		client: client,
		database: db,
	}
}

func createCollections(db *mongo.Database) error {
	// create collections
	colUsers:=db.Collection(collectionUsers)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "login", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := colUsers.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	colCreds:=db.Collection(collectionCredentials)
	indexModelCr := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}, {Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = colCreds.Indexes().CreateOne(context.Background(), indexModelCr)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

func (d *Database) Shutdown(ctx context.Context){
	d.client.Disconnect(ctx)
}

func(d *Database) AddUser(ctx context.Context, binaryUUID primitive.Binary, login string, hashedPassword string) error{
	// // Преобразование в Binary
    // binaryUUID := primitive.Binary{
    //     Subtype: 0x04, // UUID subtype
    //     Data:    userID[:],
    // }

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

func(d *Database) GetUserPassword(ctx context.Context, login string) (*User, error){
	collection := d.database.Collection(collectionUsers)
	user:=User{}
	err:=collection.FindOne(ctx, bson.M{"login": login}).Decode(&user)
	if err!=nil{
		if errors.Is(err, mongo.ErrNoDocuments){
			return nil, gopherror.ErrUnregisteredUser
		}
		return nil,err
	}
	if user.Login==""{
		return nil, gopherror.ErrUnregisteredUser
	}
	return &user, nil
}

func(d *Database) SaveUserCredentials(
	ctx context.Context, 
	userID primitive.Binary, 
	encrypted []byte, 
	metaInfo string, 
	marking string,
) error {
	collection := d.database.Collection(collectionCredentials)
	doc:=CredentialsDocument{
		UserID: userID,
		Marking: marking,
		MetaInfo: metaInfo,
		Credentials: encrypted,
	}
	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return gopherror.ErrAlreadyExists
		}
		return fmt.Errorf("failed to add credentials: %w", err)
	}
	return nil
}

func(d *Database) GetUserCredentials(ctx context.Context, userID primitive.Binary) ([]*CredentialsDocument, error){
	collection := d.database.Collection(collectionCredentials)

	filter := bson.M{"user_id": userID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var creds []*CredentialsDocument
	if err = cursor.All(context.TODO(), &creds); err != nil {
		return nil, err
	}

	if len(creds)==0{
		return nil, gopherror.ErrNoData
	}

	return creds, nil
}

func(d *Database) CreateGridFSStream(userID string, filename string, metainfo string, size int64) (*gridfs.UploadStream, error) {
	// Подключаемся к GridFS
	bucket, err := gridfs.NewBucket(
		d.database,
		options.GridFSBucket().SetName(collectionFiles),
	)
	if err != nil {
		return nil,err
	}

	// Создаем файл в GridFS с метаданными пользователя
	opts := options.GridFSUpload().
		SetMetadata(bson.M{
			"user_id":   userID,
			"metainfo": metainfo,
			"size": size,
		})

	uploadStream, err := bucket.OpenUploadStream(filename, opts)
	if err != nil {
		return nil, err
	}

	return uploadStream, nil
}

func(d *Database) GetFileMetadata(ctx context.Context, userID, filename string) (*FileInfo, error) {
	// Ищем файл в GridFS
	filter := bson.M{
		"metadata.user_id": userID,
		"filename": filename,
	}

	var fileInfo FileInfo

	err := d.database.Collection("files."+collectionFiles).FindOne(ctx, filter).Decode(&fileInfo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gopherror.ErrNoData
		}
		return nil, err
	}

	return &fileInfo, nil
}

func(d *Database) CreateDownloadStream(fileID primitive.ObjectID)(*gridfs.DownloadStream, error) {
	// Открываем поток для чтения файла
	bucket, _ := gridfs.NewBucket(d.database,
		options.GridFSBucket().SetName("files"), // Указываем имя бакета
	)
	downloadStream, err := bucket.OpenDownloadStream(fileID)
	if err != nil {
		return nil, err
	}
	return downloadStream, nil
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