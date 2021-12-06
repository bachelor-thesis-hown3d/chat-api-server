package k8sutil

import (
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"
	certmanager "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildConfig(masterUrl string, kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

//NewClientSetFromKubeconfig creates a new kubernetes rest config
func NewClientSetFromKubeconfig(kubeconfig *string) (*kubernetes.Clientset, error) {
	config, err := buildConfig("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	// create the clientset
	return kubernetes.NewForConfig(config)
}

func NewClientSetFromToken(token string) (*kubernetes.Clientset, error) {
	config, err := buildConfig("", "")
	if err != nil {
		return nil, err
	}
	config.BearerToken = token
	// create the clientset
	clientset := kubernetes.NewForConfigOrDie(config)
	return clientset, nil
}

func NewChatClientsetFromKubeconfig(kubeconfig *string) (*chatv1alpha1.ChatV1alpha1Client, error) {
	// use the current context in kubeconfig
	config, err := buildConfig("", *kubeconfig)
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

func NewChatClientSetFromToken(token string) (*chatv1alpha1.ChatV1alpha1Client, error) {
	config, err := buildConfig("", "")
	if err != nil {
		return nil, err
	}
	config.BearerToken = token
	// create the clientset
	clientset := chatv1alpha1.NewForConfigOrDie(config)
	return clientset, nil
}
