package service_test

import (
	"context"
	"testing"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/service"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/testutils"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const TestNamespace string = "test-ns"

func TestRocket_Create(t *testing.T) {
	type args struct {
		in1 *rocketpb.CreateRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *rocketpb.CreateResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.NewRocket(fake.NewSimpleClientset(), testutils.NewFakeChatClient())
			err := s.Create(context.TODO(), nil)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestRocket_Update(t *testing.T) {
	type args struct {
		rocket *chatv1alpha1.Rocket
	}
	tests := []struct {
		name    string
		args    args
		want    *rocketpb.UpdateResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.NewRocket(fake.NewSimpleClientset(), testutils.NewFakeChatClient(tt.args.rocket))
			err := s.Update(context.TODO(), nil)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestRocket_Delete(t *testing.T) {
	type args struct {
		in1    *rocketpb.DeleteRequest
		rocket *chatv1alpha1.Rocket
	}
	tests := []struct {
		name    string
		args    args
		want    *rocketpb.DeleteResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.NewRocket(fake.NewSimpleClientset(), testutils.NewFakeChatClient(tt.args.rocket))
			err := s.Delete(context.TODO(), nil)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestRocket_Get(t *testing.T) {
	type args struct {
		name   string
		rocket *chatv1alpha1.Rocket
	}
	tests := []struct {
		name    string
		args    args
		want    *v1alpha1.Rocket
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.NewRocket(fake.NewSimpleClientset(), testutils.NewFakeChatClient(tt.args.rocket))
			rocket, err := s.Get(context.TODO(), tt.args.name, TestNamespace)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.args.rocket, rocket)
		})
	}
}

func TestRocket_GetAll(t *testing.T) {
	type args struct {
		rockets []*chatv1alpha1.Rocket
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		struct {
			name    string
			args    args
			wantErr bool
		}{
			name: "two rockets",
			args: args{
				rockets: []*chatv1alpha1.Rocket{
					&chatv1alpha1.Rocket{
						ObjectMeta: v1.ObjectMeta{
							Name:      "foo",
							Namespace: TestNamespace,
						},
					},
					&chatv1alpha1.Rocket{
						ObjectMeta: v1.ObjectMeta{
							Name:      "bar",
							Namespace: TestNamespace,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.NewRocket(fake.NewSimpleClientset(), testutils.NewFakeChatClient(tt.args.rockets...))
			rockets, err := s.GetAll(context.TODO(), TestNamespace)
			if tt.wantErr {
				assert.Error(t, err)
			}
			for _, rocket := range rockets.Items {
				assert.Contains(t, tt.args.rockets, rocket)
			}

		})
	}
}

func TestRocket_Logs(t *testing.T) {
	type args struct {
		name   string
		pod    string
		stream rocketpb.RocketService_LogsServer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.NewRocket(fake.NewSimpleClientset(), testutils.NewFakeChatClient())
			err := s.Logs(tt.args.name, TestNamespace, tt.args.pod, nil)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}
