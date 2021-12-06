package health

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/health/grpc_health_v1"
	"k8s.io/client-go/kubernetes"
)

type healthChecker struct {
	kubeclient *kubernetes.Clientset
	logger     *zap.SugaredLogger
}

func NewHealthChecker(kubeclient *kubernetes.Clientset) healthChecker {
	return healthChecker{
		kubeclient: kubeclient,
		logger:     zap.S().Named("health"),
	}
}

// Check sends a ping to the kubernetes api server and response SERVING when connectable
// and NOT_SERVING when no connection can be made
func (c healthChecker) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	response := &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
	}
	// check for connectivity, respond with serving if alive
	if _, err := c.kubeclient.ServerVersion(); err == nil {
		c.logger.Debugw("api-server is ready!", "health", "check")
		response.Status = grpc_health_v1.HealthCheckResponse_SERVING
	} else {
		c.logger.Errorw("Error contacting the kubernetes api", "health", "check")
		return nil, err
	}
	return response, nil

}

func (c healthChecker) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	c.logger.Debugw("Serving the Watch request for health check", "health", "watch")
	return server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}

// AuthFuncOverride implements the ServiceAuthFuncOverride Interface from grpc_auth.
// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/560829fc74fcf9a69b7ab01d484f8b8961dc734b/auth/auth.go#L30
// This is needed to disable auth checking for the health service
func (c healthChecker) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}
