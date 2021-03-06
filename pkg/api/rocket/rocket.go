package rocket

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/k8sutil"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/service"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
)

type rocketAPIServer struct {
	service service.RocketService
}

func NewAPIServer(service service.RocketService) *rocketAPIServer {
	return &rocketAPIServer{
		service: service,
	}
}

func (r *rocketAPIServer) Create(ctx context.Context, req *rocketpb.CreateRequest) (*rocketpb.CreateResponse, error) {
	err := r.service.Create(ctx, req.GetHost(), req.GetName(), req.GetNamespace(), req.GetEmail(), req.GetUser(), req.GetDatabaseSize(), req.GetReplicas())
	if err != nil {
		return nil, err
	}
	return &rocketpb.CreateResponse{}, nil
}

func (r *rocketAPIServer) AvailableVersions(ctx context.Context, req *rocketpb.AvailableVersionsRequest) (*rocketpb.AvailableVersionsResponse, error) {
	var repo string
	switch i := req.Image; i {
	case rocketpb.AvailableVersionsRequest_IMAGE_MONGODB:
		repo = "bitnami/mongodb"
	case rocketpb.AvailableVersionsRequest_IMAGE_ROCKETCHAT:
		repo = "rocketchat/rocket.chat"
	case rocketpb.AvailableVersionsRequest_IMAGE_UNSPECIFIED:
		return &rocketpb.AvailableVersionsResponse{}, status.Error(codes.InvalidArgument, "Image doesnt match")
	default:
		return &rocketpb.AvailableVersionsResponse{}, status.Error(codes.InvalidArgument, "Image can't be empty")
	}
	tags, err := r.service.AvailableVersions(repo)
	return &rocketpb.AvailableVersionsResponse{Tags: tags}, err

}

func (r *rocketAPIServer) Status(req *rocketpb.StatusRequest, stream rocketpb.RocketService_StatusServer) error {
	if req.GetNamespace() == "" {
		return status.Error(codes.InvalidArgument, "Namespace can't be empty")
	}
	return r.service.Status(req.GetName(), req.GetNamespace(), stream)
}

func (r *rocketAPIServer) Update(_ context.Context, req *rocketpb.UpdateRequest) (*rocketpb.UpdateResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (r *rocketAPIServer) Delete(ctx context.Context, req *rocketpb.DeleteRequest) (*rocketpb.DeleteResponse, error) {

	if req.GetNamespace() == "" {
		return &rocketpb.DeleteResponse{}, status.Error(codes.InvalidArgument, "Namespace can't be empty")

	}
	err := r.service.Delete(ctx, req.GetName(), req.GetNamespace())
	if err != nil {
		return &rocketpb.DeleteResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &rocketpb.DeleteResponse{}, nil
}

func (r *rocketAPIServer) Get(ctx context.Context, req *rocketpb.GetRequest) (*rocketpb.GetResponse, error) {
	rocket, err := r.service.Get(ctx, req.GetName(), req.GetNamespace())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
	return resp, nil
}

func (r *rocketAPIServer) GetAll(ctx context.Context, req *rocketpb.GetAllRequest) (*rocketpb.GetAllResponse, error) {
	resp := &rocketpb.GetAllResponse{}
	rocketList, err := r.service.GetAll(ctx, req.Namespace)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
	if req.GetNamespace() == "" {
		return fmt.Errorf("Need to specify the namespace")
	}
	err := r.service.Logs(req.Name, req.Namespace, req.Pod, stream)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}
