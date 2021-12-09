package tenant

import (
	"context"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/oauth"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/service"
	tenantpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/tenant/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type tenantAPIServer struct {
	service service.TenantService
}

func NewAPIServer(service service.TenantService) *tenantAPIServer {
	return &tenantAPIServer{
		service: service,
	}
}

func (t *tenantAPIServer) Register(ctx context.Context, req *tenantpb.RegisterRequest) (*tenantpb.RegisterResponse, error) {
	var mem, cpu int64
	switch s := req.Size; s {
	case tenantpb.RegisterRequest_SIZE_SMALL:
		mem = smallMemory
		cpu = smallCPU
	case tenantpb.RegisterRequest_SIZE_MEDIUM:
		mem = mediumMemory
		cpu = mediumCPU
	case tenantpb.RegisterRequest_SIZE_LARGE:
		mem = largeMemory
		cpu = largeCPU
	case tenantpb.RegisterRequest_SIZE_UNSPECIFIED:
		return &tenantpb.RegisterResponse{}, status.Error(codes.InvalidArgument, "Size can't be empty")
	default:
		return &tenantpb.RegisterResponse{}, status.Error(codes.InvalidArgument, "Size can't be empty")
	}

	name := ctx.Value(oauth.NameClaimKey).(string)
	email := ctx.Value(oauth.EmailClaimKey).(string)

	return &tenantpb.RegisterResponse{}, t.service.Register(ctx, name, email, cpu, mem)
}
