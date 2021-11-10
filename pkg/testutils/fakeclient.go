package testutils

import (
	fakeChat "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/fake"
	chatv1alpha1Client "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
)

func NewFakeChatClient(objs ...*chatv1alpha1.Rocket) chatv1alpha1Client.ChatV1alpha1Interface {
	var rockets []runtime.Object
	// convert to runtime Object slice
	for _, obj := range objs {
		rockets = append(rockets, obj)
	}

	fk := fakeChat.NewSimpleClientset(rockets...)
	return fk.ChatV1alpha1()
}
