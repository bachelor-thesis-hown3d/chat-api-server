package api

import (
	"context"

	"github.com/hown3d/api-server/pkg/k8sutil"
	rocketpb "github.com/hown3d/api-server/proto/rocket/v1"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

type rocketServiceServer struct {
	kubeclient *kubernetes.Clientset
	logger     *zap.SugaredLogger
}

func NewRocketServiceServer(kubeclient *kubernetes.Clientset, logger *zap.SugaredLogger) *rocketServiceServer {
	return &rocketServiceServer{
		kubeclient: kubeclient,
		logger:     logger,
	}
}

func (r *rocketServiceServer) Create(_ context.Context, _ *rocketpb.CreateRequest) (*rocketpb.CreateResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (r *rocketServiceServer) Update(_ context.Context, req *rocketpb.UpdateRequest) (*rocketpb.UpdateResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (r *rocketServiceServer) Delete(_ context.Context, _ *rocketpb.DeleteRequest) (*rocketpb.DeleteResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (r *rocketServiceServer) Get(_ context.Context, _ *rocketpb.GetRequest) (*rocketpb.GetResponse, error) {
	panic("not implemented") // TODO: Implement
}
func (r *rocketServiceServer) GetAll(_ context.Context, _ *rocketpb.GetAllRequest) (*rocketpb.GetAllResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (r *rocketServiceServer) Logs(req *rocketpb.LogsRequest, stream rocketpb.RocketService_LogsServer) error {
	for {
		err := k8sutil.GetPodLogs(stream.Context(), r.kubeclient, req.Namespace, req.Pod, true, stream)
		if err != nil {
			r.logger.Errorw("Error getting pod logs", "instance", req.Name, "namespace", req.Namespace)
			return err
		}
	}
}
