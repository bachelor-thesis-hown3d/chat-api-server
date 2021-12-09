package k8sutil

import (
	"context"

	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// createUserRoleBinding creates a new rolebinding for the user role inside the cluster
func createUserRoleBinding(ctx context.Context, username string, email string, role *rbacv1.Role, kubeclient kubernetes.Interface) error {
	namespace := username
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      username,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{Kind: rbacv1.UserKind, Name: "oidc:" + email},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Name:     role.Name,
		},
	}

	_, err := kubeclient.RbacV1().RoleBindings(namespace).Create(ctx, rb, metav1.CreateOptions{})
	return err
}

func getOrCreateUserRole(ctx context.Context, username string, kubeclient kubernetes.Interface) (*rbacv1.Role, error) {
	namespace := username
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      username,
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs: []string{
					rbacv1.VerbAll,
				},
				APIGroups: []string{chatv1alpha1.SchemeGroupVersion.Group},
				Resources: []string{"rockets"},
			},
			{
				Verbs: []string{
					"get",
				},
				APIGroups: []string{""},
				Resources: []string{
					"pods", "pods/log",
				},
			}},
	}
	createdRole, err := kubeclient.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{})
	// check if role already exists, if yes, return the role by a get request
	if err != nil {
		if !apiErrors.IsAlreadyExists(err) {
			return nil, err
		}
		createdRole, err = kubeclient.RbacV1().Roles(namespace).Get(ctx, username, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}

	return createdRole, nil
}

func CreateRBAC(ctx context.Context, email, username string, kubeclient kubernetes.Interface) error {
	role, err := getOrCreateUserRole(ctx, username, kubeclient)

	err = createUserRoleBinding(ctx, username, email, role, kubeclient)
	return checkIfAlreadyExistsError(err)
}
