package testutils

import (
	"context"

	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	"github.com/stretchr/testify/mock"
)

/*
  Test objects
*/

// MyMockedObject is a mocked object that implements an interface
// that describes an object that the code I am testing relies on.
type MockedRocket struct {
	mock.Mock
}

// DoSomething is a method on MyMockedObject that implements some interface
// and just records the activity, and returns what the Mock object tells it to.
//
// In the real object, this method would do something useful, but since this
// is a mocked object - we're just going to stub it out.
//
// NOTE: This method is not being tested here, code that uses this object is.
func (m *MockedRocket) Get(ctx context.Context, name, namespace string) (*v1alpha1.Rocket, error) {

	args := m.Called(ctx, name, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1alpha1.Rocket), args.Error(1)

}
func (m *MockedRocket) AvailableVersions(repo string) ([]string, error) {
	args := m.Called(repo)
	return args.Get(0).([]string), args.Error(1)
}
func (m *MockedRocket) Delete(ctx context.Context, name, namespace string) error {
	args := m.Called(name, namespace)
	return args.Error(0)
}
func (m *MockedRocket) Status(name, namespace string, stream rocketpb.RocketService_StatusServer) error {
	args := m.Called(name, namespace, stream)
	return args.Error(0)
}

func (m *MockedRocket) Create(ctx context.Context, name, namespace, user, email string, databaseSize int64, replicas int32) error {
	args := m.Called(ctx, name, namespace, user, email, databaseSize)
	return args.Error(0)

}
func (m *MockedRocket) Logs(name, namespace, pod string, stream rocketpb.RocketService_LogsServer) error {
	args := m.Called(name, namespace, pod, stream)
	return args.Error(0)
}

func (m *MockedRocket) GetAll(ctx context.Context, namespace string) (*v1alpha1.RocketList, error) {

	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1alpha1.RocketList), args.Error(1)

}


func (m *MockedRocket) Register(ctx context.Context, user string, cpu, mem int64) error {
	args := m.Called(ctx, user, cpu, mem)
	return args.Error(0)
}