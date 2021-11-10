package service

import (
	"context"
	"fmt"

	"github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/k8sutil"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

type Rocket struct {
	kubeclient kubernetes.Interface
	chatclient chatv1alpha1.ChatV1alpha1Interface
	logger     *zap.SugaredLogger
}

func NewRocket(kubeclient kubernetes.Interface, chatclient chatv1alpha1.ChatV1alpha1Interface) *Rocket {
	return &Rocket{
		kubeclient: kubeclient,
		chatclient: chatclient,
		logger:     zap.S().Named("service"),
	}
}

func (r *Rocket) Create(_ context.Context, _ *rocketpb.CreateRequest) error {
	panic("not implemented") // TODO: Implement
}

func (r *Rocket) Update(_ context.Context, req *rocketpb.UpdateRequest) error {
	panic("not implemented") // TODO: Implement
}

func (r *Rocket) Delete(_ context.Context, _ *rocketpb.DeleteRequest) error {
	panic("not implemented") // TODO: Implement
}

func (r *Rocket) Get(ctx context.Context, name, namespace string) (*v1alpha1.Rocket, error) {
	requestLogger := r.logger.With("name", name, "namespace", namespace, "method", "get")
	// omitting the namespace (having it set to "" will get all rockets from all namespaces)
	rocket, err := r.chatclient.Rockets(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			requestLogger.Infof("Rocket not found in cluster")
			return nil, fmt.Errorf("Rocket %v in Namespace %v was not found", name, namespace)
		}
		requestLogger.Errorf("error getting rocket from cluster api: %w", err)
		return nil, err
	}
	return rocket, nil
}

func (r *Rocket) GetAll(ctx context.Context, namespace string) (*v1alpha1.RocketList, error) {
	requestLogger := r.logger.With("namespace", namespace, "method", "getall")
	// omitting the namespace (having it set to "" will get all rockets from all namespaces)
	rockets, err := r.chatclient.Rockets(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		requestLogger.Error("Error getting rocket list from cluster api")
		return nil, err
	}

	return rockets, nil
}

func (r *Rocket) Logs(name, namespace, pod string, stream rocketpb.RocketService_LogsServer) error {
	requestLogger := r.logger.With("name", name, "namespace", namespace, "method", "logs")
	rocket, err := r.chatclient.Rockets(namespace).Get(stream.Context(), name, v1.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			requestLogger.Infof("Rocket not found in cluster")
			return fmt.Errorf("Rocket %v in Namespace %v was not found", name, namespace)
		}
		requestLogger.Errorf("error getting rocket from cluster api: %w", err)
		return err
	}
	if pod != "" {
		requestLogger.Debugw(fmt.Sprintf("Getting logs from pod %v", pod), "instance", name, "namespace", namespace)
		err = k8sutil.GetPodLogs(stream.Context(), r.kubeclient, []string{pod}, namespace, stream)
	} else {
		var podNames []string
		for _, pod := range rocket.Status.Pods {
			podNames = append(podNames, pod.Name)
		}
		requestLogger.Debugw("Getting logs from all pods", "instance", name, "namespace", namespace)
		err = k8sutil.GetPodLogs(stream.Context(), r.kubeclient, podNames, namespace, stream)
	}
	if err != nil {
		requestLogger.Errorw("Error getting pod logs", "instance", name, "namespace", namespace)
		return err
	}
	// wait for stream context to close
	<-stream.Context().Done()

	// if stream is closed, close the quit channel to close all goroutines
	return nil
}
