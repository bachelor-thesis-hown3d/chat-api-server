package k8sutil

import (
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"
	certmanager "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//NewClientSet creates a new kubernetes rest config
func NewClientSet(kubeconfig *string) (*kubernetes.Clientset, error) {

	// use the current context in kubeconfig
	config, err := configFromFlags(kubeconfig)
	if err != nil {
		return nil, err
	}
	// create the clientset
	return kubernetes.NewForConfig(config)
}

func NewChatClientset(kubeconfig *string) (*chatv1alpha1.ChatV1alpha1Client, error) {
	// use the current context in kubeconfig
	config, err := configFromFlags(kubeconfig)
	if err != nil {
		return nil, err
	}
	// create the clientset
	return chatv1alpha1.NewForConfig(config)
}

func NewCertManagerClientset(kubeconfig *string) (*certmanager.CertmanagerV1Client, error) {
	c, err := configFromFlags(kubeconfig)
	if err != nil {
		return nil, err
	}
	return certmanager.NewForConfig(c)
}

func configFromFlags(kubeconfig *string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", *kubeconfig)
}
