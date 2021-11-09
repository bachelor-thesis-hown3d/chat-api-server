package k8sutil

import (
	"github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
)

func GetPodNamesFromRocket(rocket *v1alpha1.Rocket) (podNames []string) {
	for _, pod := range rocket.Status.Pods {
		podNames = append(podNames, pod.Name)
	}
	return
}
