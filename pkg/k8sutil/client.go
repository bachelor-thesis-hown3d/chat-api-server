package k8sutil

import (
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//NewClientSet creates a new kubernetes rest config
func NewClientSet(kubeconfig *string) (*kubernetes.Clientset, error) {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	// create the clientset
	clientset := kubernetes.NewForConfigOrDie(config)
	return clientset, nil
}

func NewChatClientset(kubeconfig *string) (*chatv1alpha1.ChatV1alpha1Client, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	// create the clientset
	clientset := chatv1alpha1.NewForConfigOrDie(config)
	return clientset, nil
}
