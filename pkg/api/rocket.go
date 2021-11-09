package api

import (
	"context"
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/internalclientset/typed/chat.accso.de/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/k8sutil"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

type rocketServiceServer struct {
	kubeclient *kubernetes.Clientset
	chatclient *chatv1alpha1.ChatV1alpha1Client
	logger     *zap.SugaredLogger
}

func NewRocketServiceServer(kubeclient *kubernetes.Clientset, chatclient *chatv1alpha1.ChatV1alpha1Client, logger *zap.SugaredLogger) *rocketServiceServer {
	return &rocketServiceServer{
		logger:     logger,
		kubeclient: kubeclient,
		chatclient: chatclient,
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

func (r *rocketServiceServer) Get(ctx context.Context, req *rocketpb.GetRequest) (*rocketpb.GetResponse, error) {
	requestLogger := r.logger.With("name", req.GetName(), "namespace", req.GetNamespace(), "method", "get")
	rocket, err := r.chatclient.Rockets(req.Namespace).Get(ctx, req.Name, v1.GetOptions{})
	if err != nil {
		requestLogger.Errorf("error getting rocket from cluster api: %w", err)
		return nil, err
	}

	resp := &rocketpb.GetResponse{
		Status:           rocket.Status.Message,
		Phase:            string(rocket.Status.Phase),
		WebserverVersion: rocket.Spec.Version,
		MongodbVersion:   rocket.Spec.Database.Version,
		Pods:             k8sutil.GetPodNamesFromRocket(rocket),
		Name: rocket.Name,
	}

	// get databasesize if exists
	storageSize := rocket.Spec.Database.StorageSpec.Status.Capacity.Storage()
	if storageSize != nil {
		resp.DatabaseSize = storageSize.String()
	}
	requestLogger.Debugf("successful request: %v", req.String())
	return resp, nil
}
func (r *rocketServiceServer) GetAll(ctx context.Context, req *rocketpb.GetAllRequest) (*rocketpb.GetAllResponse, error) {
	requestLogger := r.logger.With("namespace", req.GetNamespace(), "method", "getall")
	resp := &rocketpb.GetAllResponse{}
	rockets, err := r.chatclient.Rockets(req.Namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		requestLogger.Error("Error getting rocket list from cluster api")
		return nil, err
	}
	for _, rocket := range rockets.Items {
		resp.Rockets = append(resp.Rockets, &rocketpb.GetResponse{
			Status: rocket.Status.Message,
			Phase: string(rocket.Status.Phase),
			WebserverVersion: rocket.Spec.Version,
			MongodbVersion: rocket.Spec.Database.Version,
			Pods: k8sutil.GetPodNamesFromRocket(&rocket),
			Name:  rocket.Name,
		})
	}
	requestLogger.Debugf("successful request: %v", req.String())
	return resp, nil
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
