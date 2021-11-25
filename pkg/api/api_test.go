package api

import (
	"context"
	"log"
	"net"
	"testing"

	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/service"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/testutils"
	fakeChat "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/fake"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	bufSize       = 1024 * 1024
	TestNamespace = "test-ns"
)

var lis *bufconn.Listener

func apiInit(service service.Interface) {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	rocketpb.RegisterRocketServiceServer(s, NewAPIServer(fake.NewSimpleClientset(), fakeChat.NewSimpleClientset().ChatV1alpha1(), service))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func connCreation(t *testing.T, ctx context.Context, testService *testutils.MockedRocket) rocketpb.RocketServiceClient {

	// create server
	apiInit(testService)
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return rocketpb.NewRocketServiceClient(conn)
}

func TestGetAll(t *testing.T) {
	testName := "test-getall"
	rockets := []v1alpha1.Rocket{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      testName,
				Namespace: TestNamespace,
			},
		},
	}

	expectedResponse := rocketpb.GetAllResponse{
		Rockets: []*rocketpb.GetResponse{{Name: testName, Namespace: TestNamespace}},
	}

	// create an instance of our test object
	testService := new(testutils.MockedRocket)

	// setup expectations
	testService.
		On("GetAll", mock.MatchedBy(func(_ context.Context) bool { return true }), TestNamespace).
		Return(&v1alpha1.RocketList{
			Items: rockets,
		}, nil)

	ctx := context.Background()
	client := connCreation(t, ctx, testService)
	resp, err := client.GetAll(ctx, &rocketpb.GetAllRequest{Namespace: TestNamespace})
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	// assert that the expectations were met
	testService.AssertExpectations(t)
	assert.Equal(t, expectedResponse.Rockets, resp.Rockets)
}

func TestGet_exists(t *testing.T) {
	testName := "test-get"
	rocket := &v1alpha1.Rocket{
		ObjectMeta: v1.ObjectMeta{
			Name:      testName,
			Namespace: TestNamespace,
		},
	}

	expectedResponse := rocketpb.GetResponse{
		Name:      testName,
		Namespace: TestNamespace,
	}

	// create an instance of our test object
	testService := new(testutils.MockedRocket)

	// setup expectations
	testService.
		On("Get", mock.MatchedBy(func(_ context.Context) bool { return true }), testName, TestNamespace).
		Return(rocket, nil)

	ctx := context.Background()
	client := connCreation(t, ctx, testService)
	resp, err := client.Get(ctx, &rocketpb.GetRequest{Namespace: TestNamespace, Name: testName})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	// assert that the expectations were met
	testService.AssertExpectations(t)
	assert.Equal(t, expectedResponse.Name, resp.Name)
	assert.Equal(t, expectedResponse.Namespace, resp.Namespace)
}

func TestGet_doesnt_exists(t *testing.T) {
	testName := "test-get"

	// create an instance of our test object
	testService := new(testutils.MockedRocket)

	// setup expectations
	testService.
		On("Get", mock.MatchedBy(func(_ context.Context) bool { return true }), testName, TestNamespace).
		Return(nil, assert.AnError)

	ctx := context.Background()
	client := connCreation(t, ctx, testService)
	_, err := client.Get(ctx, &rocketpb.GetRequest{Namespace: TestNamespace, Name: testName})
	// assert that the expectations were met
	testService.AssertExpectations(t)
	assert.Error(t, err)
}
