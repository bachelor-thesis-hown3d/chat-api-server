package gateway

import (
	"context"
	"net/http"

	rocketgw "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	tenantgw "github.com/bachelor-thesis-hown3d/chat-api-server/proto/tenant/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Run(logger *zap.Logger, grpcServerEndpoint, addr string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := rocketgw.RegisterRocketServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts)
	if err != nil {
		return err
	}
	err = tenantgw.RegisterTenantServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	logger.Info("Starting gateway on " + addr)
	err = http.ListenAndServe(addr, mux)
	return err
}
