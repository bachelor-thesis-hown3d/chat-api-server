package k8sutil

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/hown3d/chat-operator/pkg/client/internalclientset"
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

func NewChatClientSet(kubeconfig *string) {
}