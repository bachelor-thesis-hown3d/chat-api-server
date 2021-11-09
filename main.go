package main

import (
	"flag"
	"fmt"
	"net"
	"path/filepath"

	"github.com/hown3d/api-server/pkg/api"
	"github.com/hown3d/api-server/pkg/health"
	"github.com/hown3d/api-server/pkg/k8sutil"
	rocketpb "github.com/hown3d/api-server/proto/rocket/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"k8s.io/client-go/util/homedir"
)

var (
	port       = flag.Int("port", 10000, "The server port")
	kubeconfig *string
)

func main() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	logger, _ := zap.Development()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	kubeclient, err := k8sutil.NewClientSet(kubeconfig)
	if err != nil {
		sugar.Fatalf("Failed to get kubernetes client from config: %w", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", *port))
	if err != nil {
		sugar.Fatalf("Failed to listen on port %v: %w", port, err)
	}

	grpcServer := grpc.NewServer()
	healthService := health.NewHealthChecker(kubeclient, sugar)
	service := api.NewRocketServiceServer(kubeclient, sugar)

	rocketpb.RegisterRocketServiceServer(grpcServer, service)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)
	reflection.Register(grpcServer)

	sugar.Infof("Starting grpc server on %v ...", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		sugar.Fatalf("Failed to start grpc Server %v", err)
	}
}
