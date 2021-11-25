package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	chatClient "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"

	certmanager "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/k8sutil"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/fields"
)

type Rocket struct {
	kubeclient        kubernetes.Interface
	chatclient        chatClient.ChatV1alpha1Interface
	certmanagerClient certmanager.CertmanagerV1Interface
	logger            *zap.SugaredLogger
}

func NewRocket(kubeclient kubernetes.Interface, chatclient chatClient.ChatV1alpha1Interface) *Rocket {
	logger := zap.S().Named("service")
	return &Rocket{
		kubeclient: kubeclient,
		chatclient: chatclient,
		logger:     logger,
	}
}

func (r *Rocket) Status(name, namespace string, stream rocketpb.RocketService_StatusServer) error {
	selectors := fields.SelectorFromSet(fields.Set{
		"metadata.name":      name,
		"metadata.namespace": namespace,
	})
	watcher, err := r.chatclient.Rockets(namespace).Watch(stream.Context(), metav1.ListOptions{FieldSelector: selectors.String()})
	if err != nil {
		return err
	}
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			event := <-watcher.ResultChan()
			rocket, ok := event.Object.(*chatv1alpha1.Rocket)
			if !ok {
				// should never be the case
				return fmt.Errorf("Watch event is not of type Rocket")
			}
			err = stream.Send(&rocketpb.StatusResponse{Status: rocket.Status.Message, Ready: rocket.Status.Ready})
			if err != nil {
				return err
			}
		}
	}
}

func (r *Rocket) Create(ctx context.Context, name, namespace, user, email string, databaseSize int64, replicas int32) error {
	requestLogger := r.logger.With("name", name, "namespace", namespace, "method", "create")

	//TODO: Use Issuer Name for Ingress
	_, err := k8sutil.NewIssuer(ctx, email, name, namespace, r.certmanagerClient)
	if err != nil {
		return err
	}

	rocket := &chatv1alpha1.Rocket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: chatv1alpha1.RocketSpec{
			Replicas: replicas,
			AdminSpec: &chatv1alpha1.RocketAdminSpec{
				Email:    email,
				Username: user,
			},
			Database: chatv1alpha1.RocketDatabase{
				Replicas: replicas,
				StorageSpec: &chatv1alpha1.EmbeddedPersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								// storage in Gi
								v1.ResourceStorage: *resource.NewQuantity(databaseSize*1024*1024*1024, resource.BinarySI),
							},
						},
					},
				},
			},
		},
	}
	requestLogger.Info("Creating rocket")
	_, err = r.chatclient.Rockets(namespace).Create(ctx, rocket, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (r *Rocket) Update(_ context.Context, req *rocketpb.UpdateRequest) error {
	panic("not implemented") // TODO: Implement
}

func (r *Rocket) Delete(ctx context.Context, name, namespace string) error {
	coreClient := r.kubeclient.CoreV1()
	// get pods from status
	chatclient := r.chatclient.Rockets(namespace)
	rocket, err := chatclient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for _, pod := range rocket.Status.Pods {
		pod, err := coreClient.Pods(namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				claimName := volume.PersistentVolumeClaim.ClaimName
				err = coreClient.PersistentVolumeClaims(namespace).Delete(ctx, claimName, metav1.DeleteOptions{})
			}
		}
	}

	err = r.chatclient.Rockets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (r *Rocket) Get(ctx context.Context, name, namespace string) (*v1alpha1.Rocket, error) {
	requestLogger := r.logger.With("name", name, "namespace", namespace, "method", "get")
	// omitting the namespace (having it set to "" will get all rockets from all namespaces)
	rocket, err := r.chatclient.Rockets(namespace).Get(ctx, name, metav1.GetOptions{})
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
	rockets, err := r.chatclient.Rockets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		requestLogger.Error("Error getting rocket list from cluster api")
		return nil, err
	}

	return rockets, nil
}

type tag struct {
	Name string `json:"name"`
}

func (r *Rocket) AvailableVersions(repo string) (tagNames []string, err error) {
	url := fmt.Sprintf("https://registry.hub.docker.com/v1/repositories/%v/tags", repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tags []tag
	err = json.Unmarshal(bytes, &tags)
	if err != nil {
		return nil, err
	}
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	return
}

func (r *Rocket) Logs(name, namespace, pod string, stream rocketpb.RocketService_LogsServer) error {
	requestLogger := r.logger.With("name", name, "namespace", namespace, "method", "logs")
	rocket, err := r.chatclient.Rockets(namespace).Get(stream.Context(), name, metav1.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			requestLogger.Infof("Rocket not found in cluster")
			return fmt.Errorf("Rocket %v in Namespace %v was not found", name, namespace)
		}
		requestLogger.Errorf("error getting rocket from cluster api: %w", err)
		return err
	}

	errChan := make(chan error, 0)

	if pod != "" {
		requestLogger.Debugf("Getting logs from pod %v", pod)
		k8sutil.GetPodLogs(r.kubeclient, []string{pod}, namespace, stream, errChan)
	} else {
		var podNames []string
		for _, pod := range rocket.Status.Pods {
			podNames = append(podNames, pod.Name)
		}
		requestLogger.Debug("Getting logs from all pods")
		k8sutil.GetPodLogs(r.kubeclient, podNames, namespace, stream, errChan)
	}
	// wait for stream context to close
	requestLogger.Debugf("Waiting for grpc context to close")
	select {
	case <-stream.Context().Done():
		requestLogger.Debugf("Context done!")
		return nil
	case err = <-errChan:
		requestLogger.Debug("Error recieving pod logs in routines")
		return err
	}

}
