package service

import (
	"context"

	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
)

// RocketService
type RocketService interface {
	Logs(name, namespace, pod string, stream rocketpb.RocketService_LogsServer) error
	GetAll(ctx context.Context, namespace string) (*v1alpha1.RocketList, error)
	Get(ctx context.Context, name, namespace string) (*v1alpha1.Rocket, error)
	Create(ctx context.Context, host, name, namespace, email, user string, databaseSize int64, replicas int32) error
	Status(name, namespace string, stream rocketpb.RocketService_StatusServer) error
	Delete(ctx context.Context, name, namespace string) error
	AvailableVersions(repo string) ([]string, error)
}
