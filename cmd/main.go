package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/api"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/grpcui"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/health"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/k8sutil"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"k8s.io/client-go/util/homedir"
	"net"
	"path/filepath"
)


var (
	port       = flag.Int("port", 10000, "The server port")
	devel = flag.Bool("devel", false, "Set the api-server to development mode (nice log, grpcui etc.)")
	kubeconfig *string
	logger *zap.Logger
)

func main() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	if *devel {
		logger, _ = zap.NewDevelopment()
		go func() {
			err := grpcui.NewGRPCUiWebServer(context.TODO(), fmt.Sprintf("0.0.0.0:%v", *port), zap.NewStdLog(logger))
			if err != nil {
				logger.Fatal(fmt.Errorf("Failed to serve grpcui web server: %w", err).Error())
			}
		}()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	kubeclient, err := k8sutil.NewClientSet(kubeconfig)
	if err != nil {
		sugar.Fatalf("Failed to get kubernetes client from config: %v", err)
	}
	chatclient, err := k8sutil.NewChatClientset(kubeconfig)
	if err != nil {
		sugar.Fatalf("Failed to get chat kubeclient from config: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", *port))
	if err != nil {
		sugar.Fatalf("Failed to listen on port %v: %w", port, err)
	}


	grpcServer := grpc.NewServer()
	healthService := health.NewHealthChecker(kubeclient, sugar)
	service := api.NewRocketServiceServer(kubeclient, chatclient, sugar)

	rocketpb.RegisterRocketServiceServer(grpcServer, service)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)
	reflection.Register(grpcServer)

	sugar.Infof("Starting grpc server on %v ...", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		sugar.Fatalf("Failed to start grpc Server %v", err)
	}
}
