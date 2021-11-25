package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/api"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/grpcui"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/health"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/k8sutil"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/service"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"k8s.io/client-go/util/homedir"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
)

var (
	port       = flag.Int("port", 10000, "The server port")
	devel      = flag.Bool("devel", false, "Set the api-server to development mode (nice log, grpcui etc.)")
	kubeconfig *string
	logger     *zap.Logger
)

func main() {
	file := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		kubeconfig = flag.String("kubeconfig", file, "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	if *devel {
		logger, _ = zap.NewDevelopment()

	} else {
		logger, _ = zap.NewProduction()
	}

	grpcServer := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_zap.UnaryServerInterceptor(logger),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_zap.StreamServerInterceptor(logger),
		),
	)

	reflection.Register(grpcServer)
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	grpc_zap.ReplaceGrpcLoggerV2(logger)
	zap.ReplaceGlobals(logger)

	defer logger.Sync() // flushes buffer, if any
	kubeclient, err := k8sutil.NewClientSet(kubeconfig)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get kubernetes client from config: %v", err))
	}
	chatclient, err := k8sutil.NewChatClientset(kubeconfig)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get chat kubeclient from config: %v", err))
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", *port))
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to listen on port %v: %v", port, err))
	}

	healthService := health.NewHealthChecker(kubeclient)

	service := service.NewRocket(kubeclient, chatclient)
	api := api.NewAPIServer(kubeclient, chatclient, service)

	rocketpb.RegisterRocketServiceServer(grpcServer, api)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)

	if *devel {
		go func() {
			err := grpcui.NewGRPCUiWebServer(context.TODO(), fmt.Sprintf("0.0.0.0:%v", *port), zap.NewStdLog(logger))
			if err != nil {
				logger.Fatal(fmt.Errorf("Failed to serve grpcui web server: %w", err).Error())
			}
		}()
	}

	logger.Info(fmt.Sprintf("Starting grpc server on %v ...", lis.Addr().String()))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to start grpc Server %v", err))
	}

}
