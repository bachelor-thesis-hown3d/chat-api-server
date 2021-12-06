package k8sutil

import (
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"
	certmanager "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

//NewClientsetFromKubeconfig creates a new kubernetes rest config
func NewClientsetFromKubeconfig(kubeconfig *string) (*kubernetes.Clientset, error) {
	c, err := buildConfig(*kubeconfig)
	if err != nil {
		return nil, err
	}
	// create the clientset
	return kubernetes.NewForConfig(c)
}

func NewClientsetFromToken(token string) (*kubernetes.Clientset, error) {
	c, err := buildConfig("")
	if err != nil {
		return nil, err
	}
	c.BearerToken = token
	// create the clientset
	return kubernetes.NewForConfig(c)
}

func NewChatClientsetFromKubeconfig(kubeconfig *string) (*chatv1alpha1.ChatV1alpha1Client, error) {
	// use the current context in kubeconfig
	c, err := buildConfig(*kubeconfig)
	if err != nil {
		return nil, err
	}
	// create the clientset
	return chatv1alpha1.NewForConfig(c)
}
func NewChatClientsetFromToken(token string) (*chatv1alpha1.ChatV1alpha1Client, error) {
	c, err := buildConfig("")
	if err != nil {
		return nil, err
	}
	c.BearerToken = token
	// create the clientset
	return chatv1alpha1.NewForConfig(c)
}

func NewCertManagerClientsetFromKubeconfig(kubeconfig *string) (*certmanager.CertmanagerV1Client, error) {
	c, err := buildConfig(*kubeconfig)
	if err != nil {
		return nil, err
	}
	return certmanager.NewForConfig(c)
}
func NewCertManagerClientsetFromToken(token string) (*certmanager.CertmanagerV1Client, error) {
	c, err := buildConfig("")
	if err != nil {
		return nil, err
	}
	c.BearerToken = token

	return certmanager.NewForConfig(c)
}
