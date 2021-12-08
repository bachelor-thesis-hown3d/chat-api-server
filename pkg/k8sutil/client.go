package k8sutil

import (
	"errors"
	"flag"
	"os"
	"path/filepath"

	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"
	certmanager "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig *string
)

func init() {
	file := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		kubeconfig = flag.String("kubeconfig", file, "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

func buildConfig(overrides *clientcmd.ConfigOverrides) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	return config.ClientConfig()
}

func buildConfigFromToken(token string) (*rest.Config, error) {
	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	overrides.AuthInfo.Token = token
	return buildConfig(overrides)
}

//NewClientsetFromKubeconfig creates a new kubernetes rest config
func NewClientsetFromKubeconfig() (*kubernetes.Clientset, error) {
	c, err := buildConfig(&clientcmd.ConfigOverrides{})
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
	c, err := buildConfig(&clientcmd.ConfigOverrides{})
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

func NewCertManagerClientsetFromKubeconfig() (*certmanager.CertmanagerV1Client, error) {
	c, err := buildConfig(&clientcmd.ConfigOverrides{})
	if err != nil {
		return nil, err
	}
	return certmanager.NewForConfig(c)
}
func NewCertManagerClientsetFromToken(token string) (*certmanager.CertmanagerV1Client, error) {
	c, err := buildConfigFromToken(token)
	if err != nil {
		return nil, err
	}

	return certmanager.NewForConfig(c)
}
