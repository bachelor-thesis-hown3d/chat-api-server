package k8sutil

import (
	"context"
	"fmt"

	acmev1 "github.com/jetstack/cert-manager/pkg/apis/acme/v1"
	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	v1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	certmanagerClient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1Client "k8s.io/client-go/kubernetes/typed/core/v1"
)

type messageType string

var (
	ACME       = messageType("acme")
	SelfSigned = messageType("selfsigned")
)

// NewIssuer creates a new Issuer inside the specifed namespace for lets encrypt certificates.
// It returns the name of the created issuer and an error, if the create failed
func NewIssuer(
	ctx context.Context,
	email string,
	userName string,
	namespace string,
	issuerType messageType,
	kubeclient kubernetes.Interface,
	client certmanagerClient.CertmanagerV1Interface) (string, error) {

	secretSelector, err := privateKeySecret(ctx, userName, kubeclient.CoreV1().Secrets(namespace))
	if err != nil {
		return "", err
	}

	issuersClient := client.Issuers(namespace)

	var config certmanagerv1.IssuerConfig
	switch t := issuerType; t {
	case ACME:
		config.ACME = &acmev1.ACMEIssuer{
			Server:         "https://acme-v02.api.letsencrypt.org/directory",
			Email:          email,
			PreferredChain: "ISRG Root X1",
			PrivateKey:     secretSelector,
		}
	case SelfSigned:
		config.SelfSigned = &certmanagerv1.SelfSignedIssuer{}
	default:
		return "", fmt.Errorf("%s is not a valid IssuerType", t)
	}

	i := &certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: userName + "-issuer",
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: config,
		},
	}
	issuer, err := issuersClient.Create(ctx, i, metav1.CreateOptions{})
	if err != nil {
		if !apiErrors.IsAlreadyExists(err) {
			return "", err
		}
	}
	return issuer.Name, nil
}

func privateKeySecret(ctx context.Context, name string, client corev1Client.SecretInterface) (v1.SecretKeySelector, error) {
	secretName := name + "-issuer-private-key"
	selector := v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: secretName,
		},
	}

	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
	}

	_, err := client.Create(ctx, s, metav1.CreateOptions{})
	err = checkIfAlreadyExistsError(err)

	return selector, err
}
