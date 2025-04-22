// Package logger realises middleware for logging HTTP requests and responces
package logger

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Log represents global var for logging.
// By default Log is no-op logger.
var Log *zap.SugaredLogger = zap.NewNop().Sugar()

// Initialize creates logger Log.
func MustInitialize() {
	zl, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	sugar := zl.Sugar()
	Log = sugar
}

func GRPCLogInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	Log.Infof("gRPC request to method: %s", info.FullMethod)
	return handler(ctx, req)
}

func GRPCStreamLogInterceptor(
    srv interface{},
    ss grpc.ServerStream,
    info *grpc.StreamServerInfo,
    handler grpc.StreamHandler,
) error {
	Log.Infof("gRPC request to method: %s", info.FullMethod)
	return handler(srv, ss)
}