package k8sutil

import (
	"context"

	v1 "github.com/jetstack/cert-manager/pkg/apis/acme/v1"
	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagerClient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewIssuer creates a new Issuer inside the specifed namespace for lets encrypt certificates.
// It returns the name of the created issuer and an error, if the create failed
func NewIssuer(ctx context.Context, email, name, namespace string, client certmanagerClient.CertmanagerV1Interface) (string, error) {
	issuersClient := client.Issuers(namespace)

	i := &certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-issuer",
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				ACME: &v1.ACMEIssuer{
					Server:         "https://acme-v02.api.letsencrypt.org/directory",
					Email:          email,
					PreferredChain: "ISRG Root X1",
				},
			},
		},
	}
	issuer, err := issuersClient.Create(ctx, i, metav1.CreateOptions{})
	return issuer.Name, err
}
