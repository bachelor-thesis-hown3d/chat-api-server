package rocket

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	chatClient "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/k8sutil"
	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/oauth"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/fields"
)

type Rocket struct {
	kubeclient kubernetes.Interface
	chatclient chatClient.ChatV1alpha1Interface
}

func NewRocketServiceImpl(kubeclient kubernetes.Interface, chatclient chatClient.ChatV1alpha1Interface) *Rocket {
	return &Rocket{
		kubeclient: kubeclient,
		chatclient: chatclient,
	}
}

func (r *Rocket) setRocketClientToUserClient(ctx context.Context) error {
	userToken, err := oauth.GetAuthTokenFromContext(ctx)
	if err != nil {
		return fmt.Errorf("Error getting token: %v", err)
	}
	userClient, err := k8sutil.NewChatClientsetFromToken(userToken)
	if err != nil {
		return fmt.Errorf("Error creating new chatClient: %v", err)
	}
	r.chatclient = userClient
	return nil
}

func (r *Rocket) setKubeClientToUserClient(ctx context.Context) error {
	userToken := ctx.Value(oauth.TokenKey).(string)
	userClient, err := k8sutil.NewClientsetFromToken(userToken)
	if err != nil {
		return fmt.Errorf("Error creating new chatClient: %v", err)
	}
	r.kubeclient = userClient
	return nil
}

func (r *Rocket) Status(name, namespace string, stream rocketpb.RocketService_StatusServer) error {
	l := ctxzap.Extract(stream.Context())
	selectors := fields.SelectorFromSet(fields.Set{
		"metadata.name":      name,
		"metadata.namespace": namespace,
	})

	err := r.setRocketClientToUserClient(stream.Context())
	if err != nil {
		err = fmt.Errorf("Error setting rocket Client for kubernetes from token: %v", err)
		l.Error(err.Error())
		return err
	}

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

func (r *Rocket) Create(ctx context.Context, host, name, namespace, email, user string, databaseSize int64, replicas int32) error {
	l := ctxzap.Extract(ctx)

	err := r.setRocketClientToUserClient(ctx)
	if err != nil {
		err = fmt.Errorf("Error setting rocket Client for kubernetes from token: %v", err)
		l.Error(err.Error())
		return err
	}

	rocket := &chatv1alpha1.Rocket{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: chatv1alpha1.RocketSpec{
			IngressSpec: chatv1alpha1.RocketIngressSpec{
				Host: host,
				Annotations: map[string]string{
					// TODO: Maybe dynamicly get the ingress class
					"kubernetes.io/ingress.class": "nginx",
					"cert-manager.io/issuer":      user + "-issuer",
				},
			},
			Replicas: replicas,
			AdminSpec: &chatv1alpha1.RocketAdminSpec{
				Email:    email,
				Username: user,
			},
			Database: chatv1alpha1.RocketDatabase{
				Replicas: replicas,
				StorageSpec: &chatv1alpha1.EmbeddedPersistentVolumeClaim{
					//TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"},
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
	l.Info("Creating rocket")
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
	l := ctxzap.Extract(ctx)
	err := r.setRocketClientToUserClient(ctx)
	if err != nil {
		err = fmt.Errorf("Error setting rocket Client for kubernetes from token: %v", err)
		l.Error(err.Error())
		return err
	}

	err = r.setKubeClientToUserClient(ctx)
	if err != nil {
		err = fmt.Errorf("Error setting kube Client for kubernetes from token: %v", err)
		l.Error(err.Error())
		return err
	}

	chatclient := r.chatclient.Rockets(namespace)
	rocket, err := chatclient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	err = k8sutil.DeleteVolumeClaim(ctx, rocket, namespace, r.kubeclient)
	if err != nil {
		return err
	}

	err = r.chatclient.Rockets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (r *Rocket) Get(ctx context.Context, name, namespace string) (*v1alpha1.Rocket, error) {
	l := ctxzap.Extract(ctx)
	err := r.setRocketClientToUserClient(ctx)
	if err != nil {
		err = fmt.Errorf("Error setting rocket Client for kubernetes from token: %v", err)
		l.Error(err.Error())
		return nil, err
	}

	// omitting the namespace (having it set to "" will get all rockets from all namespaces)
	rocket, err := r.chatclient.Rockets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			err = fmt.Errorf("Rocket %v in Namespace %v was not found", name, namespace)
			l.Error(err.Error())
			return nil, err
		}
		err = fmt.Errorf("error getting rocket from cluster api: %w", err)
		l.Error(err.Error())
		return nil, err
	}
	return rocket, nil
}

func (r *Rocket) GetAll(ctx context.Context, namespace string) (*v1alpha1.RocketList, error) {
	l := ctxzap.Extract(ctx)

	err := r.setRocketClientToUserClient(ctx)
	if err != nil {
		err = fmt.Errorf("Error setting rocket Client for kubernetes from token: %v", err)
		l.Error(err.Error())
		return nil, err
	}
	// omitting the namespace (having it set to "" will get all rockets from all namespaces)
	rockets, err := r.chatclient.Rockets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		err = fmt.Errorf("Error getting rocket list from cluster api: %v", err)
		l.Error(err.Error())
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
	l := ctxzap.Extract(stream.Context())

	err := r.setRocketClientToUserClient(stream.Context())
	if err != nil {
		err = fmt.Errorf("Error setting rocket Client for kubernetes from token: %v", err)
		l.Error(err.Error())
		return err
	}

	err = r.setKubeClientToUserClient(stream.Context())
	if err != nil {
		err = fmt.Errorf("Error setting kube Client for kubernetes from token: %v", err)
		l.Error(err.Error())
		return err
	}

	rocket, err := r.chatclient.Rockets(namespace).Get(stream.Context(), name, metav1.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			err = fmt.Errorf("Rocket %v in Namespace %v was not found", name, namespace)
			l.Error(err.Error())
			return err
		}
		err = fmt.Errorf("error getting rocket from cluster api: %w", err)
		l.Error(err.Error())
		return err
	}

	errChan := make(chan error, 0)

	if pod != "" {
		l.Debug(fmt.Sprintf("Getting logs from pod %v", pod))
		k8sutil.GetPodLogs(r.kubeclient, []string{pod}, namespace, stream, errChan)
	} else {
		var podNames []string
		for _, pod := range rocket.Status.Pods {
			podNames = append(podNames, pod.Name)
		}
		l.Debug("Getting logs from all pods")
		k8sutil.GetPodLogs(r.kubeclient, podNames, namespace, stream, errChan)
	}
	// wait for stream context to close
	l.Debug("Waiting for grpc context to close")
	select {
	case <-stream.Context().Done():
		l.Debug("Context done!")
		return nil
	case err = <-errChan:
		l.Debug("Error recieving pod logs in routines")
		return err
	}

}
