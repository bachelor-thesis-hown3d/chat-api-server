package main

import (
	"flag"
	"fmt"
	"net"

	rocketApi "github.com/bachelor-thesis-hown3d/chat-api-server/pkg/api/rocket"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/grpc/gateway"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/health"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/k8sutil"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/oauth"
	rocketService "github.com/bachelor-thesis-hown3d/chat-api-server/pkg/service/rocket"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
)

var (
	port           = flag.Int("port", 10000, "The server port")
	devel          = flag.Bool("devel", false, "Set the api-server to development mode (nice log, grpcui etc.)")
	oauthClientID  = flag.String("oauth-client-id", "kubernetes", "oauth Client ID of the issuer")
	oauthIssuerUrl = flag.String("oauth-issuer-url", "https://keycloak:8443/auth/realms/kubernetes", "oauth Client ID of the issuer")
	logger         *zap.Logger
)

func main() {
	flag.Parse()

	if *devel {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	grpcServer := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_auth.UnaryServerInterceptor(oauth.Middleware),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_auth.StreamServerInterceptor(oauth.Middleware),
			grpc_zap.StreamServerInterceptor(logger),
		),
	)

	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	//grpc_zap.ReplaceGrpcLoggerV2(logger)

	defer logger.Sync() // flushes buffer, if any
	kubeclient, err := k8sutil.NewClientsetFromKubeconfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get kubernetes client from config: %v", err))
	}
	chatclient, err := k8sutil.NewChatClientsetFromKubeconfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get chat kubeclient from config: %v", err))
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", *port))
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to listen on port %v: %v", port, err))
	}

	healthService := health.NewHealthChecker(kubeclient)

	// rocket proto Service
	rocketService := rocketService.NewRocketServiceImpl(kubeclient, chatclient)
	rocketAPI := rocketApi.NewAPIServer(rocketService)
	rocketpb.RegisterRocketServiceServer(grpcServer, rocketAPI)

	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)

	if *devel {
		reflection.Register(grpcServer)
		// go func() {
		// err := grpcui.NewGRPCUiWebServer(context.TODO(), fmt.Sprintf("0.0.0.0:%v", *port), zap.NewStdLog(logger))
		// if err != nil {
		// logger.Fatal(fmt.Errorf("Failed to serve grpcui web server: %w", err).Error())
		// }
		// }()
	}

	//start gateway server
	go gateway.Run(logger, lis.Addr().String(), fmt.Sprintf(":%v", *port+1))

	logger.Info(fmt.Sprintf("Starting grpc server on %v ...", lis.Addr().String()))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to start grpc Server %v", err))
	}
}
