package authenticator

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	
	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"github.com/Alena-Kurushkina/gophkeeper/internal/logger"
)

const tokenExp = time.Hour * 3

type (
	Claims struct {
		jwt.RegisteredClaims
		UserID uuid.UUID
	}
	userUUID string
)

var (
	UserUUIDKey userUUID = "userUUID"
	TargetMethods map[string]bool
)

func init(){
	TargetMethods=make(map[string]bool)
}

type TokenHelper struct{
	tokenKey []byte
}

func NewTokenHelper(key []byte) *TokenHelper{
	return &TokenHelper{
		tokenKey: key,
	}
}

// BuildJWTString makes token and returns it as a string.
func(h *TokenHelper) BuildJWTString(id uuid.UUID) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		// собственное утверждение
		UserID: id,
	})

	// создаём строку токена
	tokenString, err := token.SignedString(h.tokenKey)
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func(h *TokenHelper) getUserID(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(h.tokenKey), nil
		})
	if err != nil {
		v, _ := err.(*jwt.ValidationError)
		if v.Errors == jwt.ValidationErrorExpired || v.Errors == jwt.ValidationErrorSignatureInvalid {
			return uuid.Nil, gopherror.ErrTokenInvalid
		}
		return uuid.Nil, err
	}
	if !token.Valid {
		return uuid.Nil, gopherror.ErrTokenInvalid
	}
	if claims.UserID == uuid.Nil {
		return uuid.Nil, gopherror.ErrNoUserIDInToken
	}
	logger.Log.Infof("User token is valid")
	return claims.UserID, nil
}

// GRPCAuthInterceptor realises gRPC interceptor for user authentication.
// It try to get user UUID from context.
func(h *TokenHelper) GRPCAuthInterceptor() grpc.UnaryServerInterceptor {
	return func (ctx context.Context,req any,info *grpc.UnaryServerInfo,handler grpc.UnaryHandler) (interface{},error) {
		// filter grpc methods that need authorization
		if applyed, exists:=TargetMethods[info.FullMethod]; exists && !applyed{
			return handler(ctx, req)
		}

		var token string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("token")
			if len(values) > 0 {
				token = values[0]
			}
		}
		if len(token) == 0 {
			logger.Log.Infof("No token in request")
			return nil, status.Error(codes.Unauthenticated, "No token in request")
		}

		userID, err := h.getUserID(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())		
		}

		ctx = context.WithValue(ctx, UserUUIDKey, userID.String())
		logger.Log.Infof("Got user id %s from token", userID)

		return handler(ctx, req)
	}
}

func (h *TokenHelper) GRPCStreamAuthInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// фильтруем grpc методы которым не нужна авторизация
		if applyed, exists:=TargetMethods[info.FullMethod]; exists && !applyed{
			return handler(srv, ss)
		}

		// создаем обертку вокруг ServerStream
		// wrappedStream := &wrappedServerStream{ss, ss.Context()}

		var token string
		if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
			values := md.Get("token")
			if len(values) > 0 {
				token = values[0]
			}
		}
		if len(token) == 0 {
			logger.Log.Infof("No token in request")
			return status.Error(codes.Unauthenticated, "No token in request")
		}

		userID, err := h.getUserID(token)
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())		
		}

		//ctx := context.WithValue(context.TODO(), UserUUIDKey, userID.String())
		logger.Log.Infof("Got user id %s from token", userID.String())

		// Добавляем user_id в контекст
		newCtx := context.WithValue(ss.Context(), UserUUIDKey, userID.String())
		
		// Обертка для ServerStream с новым контекстом
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx: newCtx,
		}

		return handler(srv, wrappedStream)
	}
}

// Обертка для ServerStream с переопределением контекста
type wrappedServerStream struct {
    grpc.ServerStream
    ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
    return w.ctx
}

func (w *wrappedServerStream) RecvMsg(m interface{}) error {
    return w.ServerStream.RecvMsg(m)
}

func (w *wrappedServerStream) SendMsg(m interface{}) error {
    return w.ServerStream.SendMsg(m)
}

func ExtractUserIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(UserUUIDKey).(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("can't extract user id from context")
	}
	userID, err := uuid.FromString(id)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}