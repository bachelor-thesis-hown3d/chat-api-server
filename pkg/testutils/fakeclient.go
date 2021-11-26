package testutils

import (
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	fakeChat "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/fake"
	chatv1alpha1Client "github.com/bachelor-thesis-hown3d/chat-operator/pkg/client/clientset/versioned/typed/chat.accso.de/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"

	fakeCertmanager "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/fake"
	certmanagerv1Client "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
)

// NewFakeChatClient returns a faked chat client that responds with the specified objs
// returns a ChatV1alpha1Interface to interact with Rocket objects
func NewFakeChatClient(objs ...chatv1alpha1.Rocket) chatv1alpha1Client.ChatV1alpha1Interface {
	var rockets []runtime.Object
	// convert to runtime Object slice
	// inside a for loop, we are using a copy of the object
	// access via pointer to retrieve the real value
	for i := range objs {
		rockets = append(rockets, &objs[i])
	}

	return fakeChat.NewSimpleClientset(rockets...).ChatV1alpha1()
}

// NewFakeCertManagerClient returns a faked certmanager client that responds with the specified objs
// returns a CertmanagerV1Interface to interact with Certmanager objects
func NewFakeCertManagerClient(objs ...runtime.Object) certmanagerv1Client.CertmanagerV1Interface {
	return fakeCertmanager.NewSimpleClientset(objs...).CertmanagerV1()
}
