package k8sutil

import (
	"context"

	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DeleteVolumeClaim(ctx context.Context, rocket *chatv1alpha1.Rocket, namespace string, kubeclient kubernetes.Interface) error {
	coreClient := kubeclient.CoreV1()
	for _, pod := range rocket.Status.Pods {
		pod, err := coreClient.Pods(namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				claimName := volume.PersistentVolumeClaim.ClaimName
				err = coreClient.PersistentVolumeClaims(namespace).Delete(ctx, claimName, metav1.DeleteOptions{})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
