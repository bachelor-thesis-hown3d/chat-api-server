package k8sutil

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateResourceQuotaIfNotExist creates a resource Quota in a given namespace with mem and cpu restrictions
// If the namespace already exists, return no error.
// Memory will be created in Gi and CPU as milliCores (for e.g. parsing 500 will end up as 500m).
func CreateResourceQuotaIfNotExist(ctx context.Context, cpu, mem int64, namespace string, client kubernetes.Interface) error {
	quotasClient := client.CoreV1().ResourceQuotas(namespace)

	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mem-cpu-quota",
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewMilliQuantity(cpu, resource.BinarySI),
				corev1.ResourceMemory: *resource.NewQuantity(mem*1024*1024*1024, resource.BinarySI),
			},
		},
	}
	_, err := quotasClient.Create(ctx, quota, metav1.CreateOptions{})
	err = checkIfAlreadyExistsError(err)
	return err
}
