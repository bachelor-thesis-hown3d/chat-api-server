package tenant

import (
	"context"
	"fmt"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/k8sutil"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	certmanager "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	"k8s.io/client-go/kubernetes"
)

type Tenant struct {
	kubeclient        kubernetes.Interface
	certmanagerClient certmanager.CertmanagerV1Interface
}

func NewTenantServiceImpl(kubeclient kubernetes.Interface, certmanagerClient certmanager.CertmanagerV1Interface) *Tenant {
	return &Tenant{
		kubeclient:        kubeclient,
		certmanagerClient: certmanagerClient,
	}
}

func (t *Tenant) Register(ctx context.Context, name string, email string, cpu, mem int64) error {
	l := ctxzap.Extract(ctx)

	namespace := name

	err := k8sutil.CreateNamespaceIfNotExist(ctx, namespace, t.kubeclient)
	if err != nil {
		err = fmt.Errorf("Error creating namespace: %v", err)
		l.Error(err.Error())
		return err
	}

	err = k8sutil.CreateResourceQuotaIfNotExist(ctx, cpu, mem, namespace, t.kubeclient)
	if err != nil {
		err = fmt.Errorf("Error creating resource Quota: %v", err)
		l.Error(err.Error())
		return err
	}

	//TODO: Use Issuer Name for Ingress
	_, err = k8sutil.NewIssuer(ctx, email, name, namespace, k8sutil.SelfSigned, t.kubeclient, t.certmanagerClient)
	if err != nil {
		err = fmt.Errorf("Error setting rocket Client for kubernetes from token: %v", err)
		l.Error(err.Error())
		return err
	}

	return nil
}
