package k8sutil

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateNamespaceIfNotExist creates a namespace for the given name and returns an error if request wasn't successful
// If the namespace already exists, return no error
func CreateNamespaceIfNotExist(ctx context.Context, name string, client kubernetes.Interface) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	err = checkIfAlreadyExistsError(err)
	return err
}
