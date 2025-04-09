package api

import (
	"context"
	"errors"
	"io"

	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/Alena-Kurushkina/gophkeeper/internal/authenticator"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gophkeeper"
	pb "github.com/Alena-Kurushkina/gophkeeper/internal/grpc/proto"
	"github.com/hashicorp/go-multierror"
	uuid "github.com/satori/go.uuid"
)

type IGophKeeperCore interface {
	UserRegister(ctx context.Context, creds gophkeeper.Credentials) (string, error)
	UserLogin(ctx context.Context, cr gophkeeper.Credentials) (string, error)
	SaveCredentials(ctx context.Context, userID uuid.UUID, in gophkeeper.CredentialsStorage) error
	GetAllUserCredentials(ctx context.Context, userID uuid.UUID, password string) (gophkeeper.AllUserCredentials,error)
	CreateGridFSStream(userID string, filename, metainfo string)(*gridfs.UploadStream, error)
}

type Server struct {
    pb.UnimplementedGophkeeperServer
	Core IGophKeeperCore
}

func (s *Server) CheckConnection(ctx context.Context, in *pb.Hello) (*pb.HelloResponse, error) {
    return &pb.HelloResponse{Response: "Secure Hello " + in.Name}, nil
}

// Register регистрирует нового пользователя в сервисе, проверяет уникальность логина
func (s *Server) Register(ctx context.Context, in *pb.Credentials) (*pb.TokenResponse, error) {
	if len(in.Password) == 0 || len(in.Login) == 0 {
		// неверный формат запроса
		return nil, status.Errorf(codes.InvalidArgument, "Empty login or password")
	}

	creds:=gophkeeper.Credentials{
		Login: in.Login,
		Password: in.Password,
	}
	token, err:=s.Core.UserRegister(ctx, creds)
	if err != nil {
		if errors.Is(err, gopherror.ErrLoginAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "Login is already used by another user")			
		} else {
			return nil, err
		}
	}

	return &pb.TokenResponse{
		Token: token,
	}, nil
}

// Login реализует вход пользователя в систему. Метод проверяет правильность пароля для указанного логина, 
// генерирует id пользователя и возвращает его в JWT токене
func (s *Server) Login(ctx context.Context, in *pb.Credentials)(*pb.TokenResponse, error) {
	if len(in.Password) == 0 || len(in.Login) == 0 {
		// неверный формат запроса
		return nil, status.Errorf(codes.InvalidArgument, "Empty login or password")
	}
	creds:=gophkeeper.Credentials{
		Login: in.Login,
		Password: in.Password,
	}
	token, err:=s.Core.UserLogin(ctx, creds)
	if err!=nil{
		switch {
		case errors.Is(err, gopherror.ErrUnregisteredUser):
			return nil, status.Errorf(codes.NotFound, "Specified user login isn't found. Register the user.")
		case errors.Is(err, gopherror.ErrUnauthorized):
			return nil, status.Errorf(codes.Unauthenticated , "Wrong login or password")
		default:
			return nil, err
		}
	}

	return &pb.TokenResponse{
		Token: token,
	}, nil
}

// SaveCredentials позволяет сохранить пары логин и пароль, а также доп. информацию в хранилище
// Пара метка и id пользователя должны быть уникальными, иначе вернётся ошибка и сохранения не произойдёт.
// Метод принимает на вход пароль для шифрования пары логин и пароль методом AES GSM, тот же пароль необходимо
// использовать для прочтения этих данных в соответствующем методе
func (s *Server) SaveCredentials(ctx context.Context, in *pb.SaveCredsRequest)(*pb.None, error){
	userID, err:=authenticator.ExtractUserIDFromCtx(ctx)
	if err!=nil{
		return nil, status.Errorf(codes.Internal, "Fail while getting user id from context")
	}
	credsBytes, err:=proto.Marshal(in.Creds)
	if err!=nil{
		return nil, err
	}
	err=s.Core.SaveCredentials(ctx, userID, gophkeeper.CredentialsStorage{
		CredentialsInfo: gophkeeper.CredentialsInfo{
			Creds: credsBytes,
			MetaInfo: in.Metainfo,
			Marking: in.Marking,
		},
		Password: in.Password,
	})
	if err!=nil{
		if errors.Is(err, gopherror.ErrAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "Specified marking is already used by this user")			
		} else {
			return nil, err
		}
	}

	return &pb.None{}, nil
}

// GetUserAllCredentials возвращает все пары логин/пароль и доп. данные, сохранённые пользователем
// Для расшифровки используется пароль, который передаётся на вход метода
func (s *Server) GetUserAllCredentials(ctx context.Context, in *pb.GetCredsRequest)(*pb.UserCredsResponse, error){
	userID, err:=authenticator.ExtractUserIDFromCtx(ctx)
	if err!=nil{
		return nil, status.Errorf(codes.Internal, "Fail while getting user id from context")
	}

	creds,errList:=s.Core.GetAllUserCredentials(ctx, userID, in.Password)
	if errList!=nil && !errors.Is(errList, gopherror.ErrDecryptAuth){
		if errors.Is(errList, gopherror.ErrNoData) {
			return nil, status.Errorf(codes.NotFound, "No data found in storage")			
		} else {
			return nil, errList
		}
	}

	//var errList error 
	respItems:=make([]*pb.UserCredsResponse_ResponseItem,0,len(creds))
	for _, item:=range creds{
		protoCreds:=pb.Credentials{}
		err=proto.Unmarshal(item.Creds,&protoCreds)
		if err!=nil{
			errList=multierror.Append(errList, err)
		} else {
			respItems=append(respItems, &pb.UserCredsResponse_ResponseItem{
				Marking: item.Marking,
				Creds: &protoCreds,
				Metainfo: item.MetaInfo,
			})
		}
	}

	return &pb.UserCredsResponse{
		Items: respItems,
	}, errList
}

// UploadBigData позволяет сохранять данные неограниченного объёма в хранилище
// Также метод сохраняет переданную метаинформацию
func (s *Server) UploadBigData(stream pb.Gophkeeper_UploadBigDataServer) error {
	userID, err:=authenticator.ExtractUserIDFromCtx(stream.Context())
	if err!=nil{
		return status.Errorf(codes.Internal, "Fail while getting user id from context")
	}

	// Получаем метаданные из первого сообщения
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	metadata := req.GetMetadata()
	if metadata == nil {
		return stream.SendAndClose(&pb.UploadResponse{
			Status: "error: metadata required as first message",
		})
	}

	// Подключаемся к GridFS
	uploadStream, err:=s.Core.CreateGridFSStream(userID.String(), metadata.Filename, metadata.Metainfo)
	if err!=nil{
		return status.Errorf(codes.Internal, "Fail to create GridFS stream %v", err.Error())
	}
	defer uploadStream.Close()

	// Принимаем и записываем чанки
	var totalSize int64
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		chunk := req.GetChunk()
		if chunk == nil {
			continue // Пропускаем некорректные сообщения
		}

		n, err := uploadStream.Write(chunk)
		if err != nil {
			return err
		}
		totalSize += int64(n)
	}

	// 5. Возвращаем ответ с ID файла
	fileID := uploadStream.FileID.(primitive.ObjectID).Hex()
	return stream.SendAndClose(&pb.UploadResponse{
		FileId: fileID,
		Size:   totalSize,
		Status: "success",
	})
}
