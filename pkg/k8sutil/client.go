package k8sutil

import (
	"errors"
	"flag"
	"os"
	"path/filepath"

	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/clientauthentication"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig *string
)

func CreateKubeconfigFlag() {
	file := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		kubeconfig = flag.String("kubeconfig", file, "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

func buildConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	return config.ClientConfig()
}

func buildConfigFromToken(token string) (*rest.Config, error) {
	restConfig, err := buildConfig()
	if err != nil {
		return nil, err
	}

	restConfig, err = rest.ExecClusterToConfig(&clientauthentication.Cluster{
		Server:                   restConfig.Host,
		TLSServerName:            restConfig.ServerName,
		CertificateAuthorityData: restConfig.CAData,
	})

	if err != nil {
		return nil, err
	}

	restConfig.BearerToken = token
	return restConfig, nil
}

//NewClientsetFromKubeconfig creates a new kubernetes rest config
func NewClientsetFromKubeconfig() (*kubernetes.Clientset, error) {
	c, err := buildConfig()
	if err != nil {
		return nil, err
	}
	// create the clientset
	return kubernetes.NewForConfig(c)
}

func NewClientsetFromToken(token string) (*kubernetes.Clientset, error) {
	c, err := buildConfigFromToken(token)
	if err != nil {
		return nil, err
	}
	// create the clientset
	return kubernetes.NewForConfig(c)
}

func NewChatClientsetFromKubeconfig() (*chatv1alpha1.ChatV1alpha1Client, error) {
	// use the current context in kubeconfig
	c, err := buildConfig()
	if err != nil {
		return nil, err
	}
	// create the clientset
	return chatv1alpha1.NewForConfig(c)
}
func NewChatClientsetFromToken(token string) (*chatv1alpha1.ChatV1alpha1Client, error) {
	c, err := buildConfigFromToken(token)
	if err != nil {
		return nil, err
	}
	// create the clientset
	return chatv1alpha1.NewForConfig(c)
}
