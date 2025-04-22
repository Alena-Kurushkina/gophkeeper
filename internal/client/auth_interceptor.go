package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct {
    Token string
}

func (i *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        // Пропустить для метода Login
        if method != "gophkeeper.Gophkeeper/Login" && method != "gophkeeper.Gophkeeper/Register" && i.Token != "" {
            ctx = metadata.AppendToOutgoingContext(ctx, "token", i.Token)
        }
        return invoker(ctx, method, req, reply, cc, opts...)
    }
}

// Stream Interceptor для клиента
func (i *AuthInterceptor) Stream() grpc.StreamClientInterceptor{
    return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
        // Пропустить для метода Login
        if method != "gophkeeper.Gophkeeper/Login" && method != "gophkeeper.Gophkeeper/Register" && i.Token != "" {
            
            ctx = metadata.AppendToOutgoingContext(ctx, "token", i.Token)
        }
        return streamer(ctx, desc, cc, method, opts...)
    }
}