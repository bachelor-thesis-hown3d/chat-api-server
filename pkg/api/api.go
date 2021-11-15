package api

import (
	"context"
	"fmt"

	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/k8sutil"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/service"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

type rocketAPIServer struct {
	service service.Interface
	logger  *zap.SugaredLogger
}

func NewAPIServer(kubeclient kubernetes.Interface, chatclient chatv1alpha1.ChatV1alpha1Interface, service service.Interface) *rocketAPIServer {
	return &rocketAPIServer{
		service: service,
		logger:  zap.S().Named("api"),
	}
}

func (r *rocketAPIServer) Create(ctx context.Context, req *rocketpb.CreateRequest) (*rocketpb.CreateResponse, error) {
	requestLogger := r.logger.With("name", req.GetName(), "namespace", req.GetNamespace(), "method", "create")
	requestLogger.Debug("New Request")
	err := r.service.Create(ctx, req.GetName(), req.GetNamespace(), req.GetName(), req.GetEmail(), req.GetDatabaseSize(), req.GetReplicas())
	if err != nil {
		return nil, err
	}
	return &rocketpb.CreateResponse{}, nil
}

func (r *rocketAPIServer) AvailableVersions(ctx context.Context, req *rocketpb.AvailableVersionsRequest) (*rocketpb.AvailableVersionsResponse, error) {
	switch i := req.Image; i {
	case rocketpb.AvailableVersionsRequest_MONGODB:
		tags, err := r.service.AvailableVersions("bitnami/mongodb")
		return &rocketpb.AvailableVersionsResponse{Tags: tags}, err

	case rocketpb.AvailableVersionsRequest_ROCKETCHAT:
		tags, err := r.service.AvailableVersions("rocketchat/rocket.chat")
		return &rocketpb.AvailableVersionsResponse{Tags: tags}, err
	}
	return &rocketpb.AvailableVersionsResponse{}, fmt.Errorf("Image doesnt match")

}

func (r *rocketAPIServer) Status(req *rocketpb.StatusRequest, stream rocketpb.RocketService_StatusServer) error {

	requestLogger := r.logger.With("name", req.GetName(), "namespace", req.GetNamespace(), "method", "status")
	requestLogger.Debug("New Request")
	if req.GetNamespace() == "" {
		return fmt.Errorf("Need to specify the namespace")
	}
	return r.service.Status(req.GetName(), req.GetNamespace(), stream)
}

func (r *rocketAPIServer) Update(_ context.Context, req *rocketpb.UpdateRequest) (*rocketpb.UpdateResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (r *rocketAPIServer) Delete(ctx context.Context, req *rocketpb.DeleteRequest) (*rocketpb.DeleteResponse, error) {

	requestLogger := r.logger.With("name", req.GetName(), "namespace", req.GetNamespace(), "method", "delete")
	requestLogger.Debug("New Request")
	if req.GetNamespace() == "" {
		return &rocketpb.DeleteResponse{}, fmt.Errorf("Need to specify the namespace")
	}
	err := r.service.Delete(ctx, req.GetName(), req.GetNamespace())
	return &rocketpb.DeleteResponse{}, err
}

func (r *rocketAPIServer) Get(ctx context.Context, req *rocketpb.GetRequest) (*rocketpb.GetResponse, error) {
	requestLogger := r.logger.With("name", req.GetName(), "namespace", req.GetNamespace(), "method", "get")
	requestLogger.Debug("New Request")
	rocket, err := r.service.Get(ctx, req.GetName(), req.GetNamespace())
	if err != nil {
		return nil, err
	}
	resp := &rocketpb.GetResponse{
		Status:           rocket.Status.Message,
		Phase:            string(rocket.Status.Phase),
		WebserverVersion: rocket.Spec.Version,
		MongodbVersion:   rocket.Spec.Database.Version,
		Pods:             k8sutil.GetPodNamesFromRocket(rocket),
		Name:             rocket.Name,
		Namespace:        rocket.Namespace,
	}

	// get databasesize if exists
	storageSpec := rocket.Spec.Database.StorageSpec
	if storageSpec != nil {
		resp.DatabaseSize = storageSpec.Status.Capacity.Storage().String()
	}
	requestLogger.Debugf("successful request: %v", req.String())
	return resp, nil
}

func (r *rocketAPIServer) GetAll(ctx context.Context, req *rocketpb.GetAllRequest) (*rocketpb.GetAllResponse, error) {
	requestLogger := r.logger.With("namespace", req.GetNamespace(), "method", "getall")
	requestLogger.Debug("New Request")
	resp := &rocketpb.GetAllResponse{}
	rocketList, err := r.service.GetAll(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}

	for _, rocket := range rocketList.Items {
		resp.Rockets = append(resp.Rockets, &rocketpb.GetResponse{
			Status:           rocket.Status.Message,
			Phase:            string(rocket.Status.Phase),
			WebserverVersion: rocket.Spec.Version,
			MongodbVersion:   rocket.Spec.Database.Version,
			Pods:             k8sutil.GetPodNamesFromRocket(&rocket),
			Name:             rocket.Name,
			Namespace:        rocket.Namespace,
		})
	}
	return resp, nil
}

func (r *rocketAPIServer) Logs(req *rocketpb.LogsRequest, stream rocketpb.RocketService_LogsServer) error {
	requestLogger := r.logger.With("namespace", req.GetNamespace(), "method", "logs")
	if req.GetNamespace() == "" {
		return fmt.Errorf("Need to specify the namespace")
	}
	requestLogger.Debug("New Request")
	return r.service.Logs(req.Name, req.Namespace, req.Pod, stream)
}
