package api

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	gomock "github.com/golang/mock/gomock"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	_ "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"

	"github.com/Alena-Kurushkina/gophkeeper/internal/authenticator"
	"github.com/Alena-Kurushkina/gophkeeper/internal/crypter"
	_ "github.com/Alena-Kurushkina/gophkeeper/internal/crypter"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"github.com/Alena-Kurushkina/gophkeeper/internal/gophkeeper"
	pb "github.com/Alena-Kurushkina/gophkeeper/internal/grpc/proto"
	storage "github.com/Alena-Kurushkina/gophkeeper/internal/storage"
)

const (
	bufSize = 1024 * 1024
	password="123456"
	token="klfjbshjfbk"
)

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func createTestClient(t *testing.T) (pb.GophkeeperClient, *gophkeeper.GophKeeperCore) {	
	lis = bufconn.Listen(bufSize)

	th:=authenticator.NewTokenHelper([]byte(token))
	interc:=th.GRPCAuthInterceptor()

	s := grpc.NewServer(
		grpc.UnaryInterceptor(interc),
	)
	authenticator.TargetMethods["/gophkeeper.Gophkeeper/Register"]=false
	authenticator.TargetMethods["/gophkeeper.Gophkeeper/Login"]=false

	ctrl := gomock.NewController(t)
	//m := NewMockStorager(ctrl)

	m:=NewMockShutdowner(ctrl)

	// core := gophkeeper.NewGophKeeperCore(m)

	core:=&gophkeeper.GophKeeperCore{
		DB: m,
		//UserCore: gophkeeper.NewUserCore(m),
		// CredentialsCore: gophkeeper.NewCredentialsCore(m),
		// GridFSCore: gophkeeper.NewGridFSCore(m),
	}

	pb.RegisterGophkeeperServer(s, &Server{
		Core: core,
	})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return pb.NewGophkeeperClient(conn), core
}

func TestRegister(t *testing.T) {
	client, core := createTestClient(t)
	
	ctrl := gomock.NewController(t)
	m:=NewMockUserer(ctrl)

	m.EXPECT().AddUser(gomock.Any(), gomock.Any(), "login_test1", gomock.Any()).Return(nil)
	m.EXPECT().AddUser(gomock.Any(), gomock.Any(), "login_test_existed", gomock.Any()).Return(gopherror.ErrLoginAlreadyExists)

	core.UserCore=gophkeeper.NewUserCore(m, []byte(token))

	tests := []struct {
		name    string
		request *pb.Credentials
		wantErr error
	}{
		{
			name:    "valid request",
			request: &pb.Credentials{Login: "login_test1", Password: password},
			wantErr: nil,
		},
		{
			name:    "request with existed login",
			request: &pb.Credentials{Login: "login_test_existed", Password: password},
			wantErr: status.Errorf(codes.AlreadyExists, "Login is already used by another user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Register(context.Background(), tt.request)
			if err != nil {
				if tt.wantErr!=nil {					
					assert.ErrorIs(t,err,tt.wantErr)					
					return
				} else {
					assert.NoError(t, err)
				}
			}
			claims := &authenticator.Claims{}
			token, err := jwt.ParseWithClaims(resp.Token, claims,
				func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}
					return []byte(token), nil
				})
			assert.NoError(t, err)
			assert.Equal(t,true, token.Valid)
		})
	}
}

func TestLogin(t *testing.T) {
	client, core := createTestClient(t)

	ctrl := gomock.NewController(t)
	m:=NewMockUserer(ctrl)

	hashedPas, err:=crypter.HashPassword(password)
	if err!=nil{
		panic(err)
	}
	m.EXPECT().GetUserPassword(gomock.Any(), "login_test1").Return(&storage.User{
		Login: "login_test1",
		PasswordHash: hashedPas,
		CreatedAt: time.Now(),
		UserID: gophkeeper.UUIDToBinary(uuid.NewV4()),
	}, nil).AnyTimes()
	m.EXPECT().GetUserPassword(gomock.Any(), "login_test_unexisted").Return(nil, gopherror.ErrUnregisteredUser)

	core.UserCore=gophkeeper.NewUserCore(m, []byte(token))

	tests := []struct {
		name    string
		request *pb.Credentials
		wantErr error
	}{
		{
			name:    "valid request",
			request: &pb.Credentials{Login: "login_test1", Password: password},
			wantErr: nil,
		},
		{
			name:    "request with unexisted login",
			request: &pb.Credentials{Login: "login_test_unexisted", Password: password},
			wantErr: status.Errorf(codes.NotFound, "Specified user login isn't found. Register the user."),
		},
		{
			name:    "request with wrong password",
			request: &pb.Credentials{Login: "login_test1", Password: "11111"},
			wantErr: status.Errorf(codes.Unauthenticated , "Wrong login or password"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Login(context.Background(), tt.request)
			if err != nil {
				if tt.wantErr!=nil {					
					assert.ErrorIs(t,err,tt.wantErr)					
					return
				} else {
					assert.NoError(t, err)
				}
			}
			claims := &authenticator.Claims{}
			token, err := jwt.ParseWithClaims(resp.Token, claims,
				func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}
					return []byte(token), nil
				})
			assert.NoError(t, err)
			assert.Equal(t,true, token.Valid)
		})
	}
}

func TestSaveCredentials(t *testing.T) {
	client, core := createTestClient(t)

	ctrl := gomock.NewController(t)
	m:=NewMockCredentialer(ctrl)

	m.EXPECT().SaveUserCredentials(
		gomock.Any(), 
		gomock.Any(), 
		gomock.Any(), 
		gomock.Any(), 
		"credentials mark",
	).Return(nil)

	m.EXPECT().SaveUserCredentials(
		gomock.Any(), 
		gomock.Any(), 
		gomock.Any(), 
		gomock.Any(), 
		"existed user credentials mark",
	).Return(gopherror.ErrAlreadyExists)

	core.CredentialsCore=gophkeeper.NewCredentialsCore(m)

	userID:=uuid.NewV4()
	hl:=authenticator.NewTokenHelper([]byte(token))
	token, err:=hl.BuildJWTString(userID)
	if err!=nil{
		panic(token)
	}

	tests := []struct {
		name    string
		request *pb.SaveCredsRequest
		wantErr error
	}{
		{
			name:    "valid request",
			request: &pb.SaveCredsRequest{Marking: "credentials mark", Creds: &pb.Credentials{}, Encrpassword: "1234", Metainfo: "info"},
			wantErr: nil,
		},
		{
			name:    "request with duplicate key",
			request: &pb.SaveCredsRequest{Marking: "existed user credentials mark", Creds: &pb.Credentials{}, Encrpassword: "1234", Metainfo: "info"},
			wantErr: status.Errorf(codes.AlreadyExists, "Specified marking is already used by this user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx:= metadata.AppendToOutgoingContext(context.Background(), "token", token)
			_,err := client.SaveCredentials(ctx, tt.request)
			if err != nil {
				if tt.wantErr!=nil {					
					assert.ErrorIs(t,err,tt.wantErr)					
					return
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestGetCredentials(t *testing.T) {
	client, core := createTestClient(t)

	ctrl := gomock.NewController(t)
	m:=NewMockCredentialer(ctrl)

	validUserID:=uuid.NewV4()
	emptyUserID:=uuid.NewV4()
	
	credsBytes, err:=proto.Marshal(&pb.Credentials{Login: "test", Password: "pas"})
	if err!=nil{
		panic(err)
	}
	encrypted, err:=crypter.Encrypt(credsBytes, password)
	if err!=nil{
		panic(err)
	}

	ret:=[]*storage.CredentialsDocument{}
	ret=append(ret, &storage.CredentialsDocument{
		Credentials: encrypted,
		UserID: gophkeeper.UUIDToBinary(validUserID),
		Marking: "",
		MetaInfo: "",
	})
	

	m.EXPECT().GetUserCredentials(
		gomock.Any(), 
		gophkeeper.UUIDToBinary(validUserID),
	).Return(ret, nil).AnyTimes()

	m.EXPECT().GetUserCredentials(
		gomock.Any(), 
		gophkeeper.UUIDToBinary(emptyUserID),
	).Return(nil,gopherror.ErrNoData)

	core.CredentialsCore=gophkeeper.NewCredentialsCore(m)

	hl:=authenticator.NewTokenHelper([]byte(token))
	tokenValidUser, err:=hl.BuildJWTString(validUserID)
	if err!=nil{
		panic(err)
	}
	tokenEmptyUser, err:=hl.BuildJWTString(emptyUserID)
	if err!=nil{
		panic(err)
	}

	tests := []struct {
		name    string
		request *pb.GetCredsRequest
		token string
		wantErr error
	}{
		{
			name:    "valid request",
			request: &pb.GetCredsRequest{Password:password},
			token: tokenValidUser,
			wantErr: nil,
		},
		{
			name:    "request with wrong password",
			token: tokenValidUser,
			request: &pb.GetCredsRequest{Password:"dhbth"},
			wantErr: status.Errorf(codes.Unknown,"1 error occurred:\n\t* message authentication failed while decrypt\n\n"),
		},
		{
			name:    "request for empty data",
			token: tokenEmptyUser,
			request: &pb.GetCredsRequest{Password:password},
			wantErr: status.Errorf(codes.NotFound, "No data found in storage"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx:= metadata.AppendToOutgoingContext(context.Background(), "token", tt.token)
			_,err := client.GetUserAllCredentials(ctx, tt.request)
			if err != nil {
				if tt.wantErr!=nil {					
					assert.ErrorIs(t,err,tt.wantErr)					
					return
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}