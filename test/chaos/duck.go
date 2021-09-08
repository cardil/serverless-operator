package chaos

import (
	"context"
	"testing"
	"time"

	"github.com/openshift-knative/serverless-operator/test"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	"knative.dev/pkg/signals"
	pkgupgrade "knative.dev/pkg/test/upgrade"
)

const chaosduckName = "chaosduck"

func NewDuck(ctx pkgupgrade.Context, cfg Config) Duck {
	return &duffy{
		ctx:        signals.NewContext(),
		upgradeCtx: ctx,
		cfg:        cfg,
	}
}

type Duck interface {
	Quak()
}

type duffy struct {
	ctx        context.Context
	upgradeCtx pkgupgrade.Context
	cfg        Config
}

func (d *duffy) Quak() {
	t := d.upgradeCtx.T
	clusterCtx := test.SetupClusterAdmin(t)
	defer test.CleanupAll(t, clusterCtx)

	for _, namespace := range d.cfg.Namespaces {
		d.quakOnNamespace(namespace, clusterCtx)
	}
	d.watchTheWorldBurn()
}

func (d *duffy) quakOnNamespace(namespace string, clusterCtx *test.Context) {
	log := d.upgradeCtx.Log
	log.Infof("Unleashing duck on namespace: %s", namespace)
	sd := dddddDuffy{
		duffy:      d,
		clusterCtx: clusterCtx,
		namespace:  namespace,
	}
	sd.quak()
}

type dddddDuffy struct {
	*duffy
	namespace  string
	clusterCtx *test.Context
}

func (d *dddddDuffy) quak() {
	d.serviceAccount()
	d.role()
	d.roleBinding()
	d.deployment()
}

func (d *dddddDuffy) t() testing.TB {
	return d.upgradeCtx.T
}

func (d *dddddDuffy) serviceAccount() {
	k := d.clusterCtx.Clients.Kube
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: chaosduckName},
	}
	serviceAccounts := k.CoreV1().ServiceAccounts(d.namespace)
	_, err := serviceAccounts.Create(d.ctx, sa, metav1.CreateOptions{})
	assert.NoError(d.t(), err)
	d.clusterCtx.AddToCleanup(func() error {
		return serviceAccounts.Delete(d.ctx, sa.GetName(), metav1.DeleteOptions{})
	})
}

func (d *dddddDuffy) role() {
	k := d.clusterCtx.Clients.Kube
	roles := k.RbacV1().Roles(d.namespace)
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: chaosduckName},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get", "delete"},
		}, {
			APIGroups: []string{"coordination.k8s.io"},
			Resources: []string{"leases"},
			Verbs:     []string{"get", "list"},
		}},
	}
	_, err := roles.Create(d.ctx, role, metav1.CreateOptions{})
	assert.NoError(d.t(), err)
	d.clusterCtx.AddToCleanup(func() error {
		return roles.Delete(d.ctx, role.GetName(), metav1.DeleteOptions{})
	})
}

func (d *dddddDuffy) roleBinding() {
	k := d.clusterCtx.Clients.Kube
	roleBindings := k.RbacV1().RoleBindings(d.namespace)
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: chaosduckName},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     chaosduckName,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      chaosduckName,
			Namespace: d.namespace,
		}},
	}
	_, err := roleBindings.Create(d.ctx, rb, metav1.CreateOptions{})
	assert.NoError(d.t(), err)
	d.clusterCtx.AddToCleanup(func() error {
		return roleBindings.Delete(d.ctx, rb.GetName(), metav1.DeleteOptions{})
	})
}

func (d *dddddDuffy) deployment() {
	k := d.clusterCtx.Clients.Kube
	var falseVal = false
	deployments := k.AppsV1().Deployments(d.namespace)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: chaosduckName},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": chaosduckName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": chaosduckName},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: chaosduckName,
					Containers: []corev1.Container{{
						Name: chaosduckName,
						// TODO: use dynamic image resolution
						Image: d.cfg.Image,
						Env: []corev1.EnvVar{{
							Name: "SYSTEM_NAMESPACE",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
							},
						}},
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: &falseVal,
						},
					}},
				},
			},
		},
	}
	_, err := deployments.Create(d.ctx, deployment, metav1.CreateOptions{})
	assert.NoError(d.t(), err)
	d.clusterCtx.AddToCleanup(func() error {
		return deployments.Delete(d.ctx, deployment.GetName(), metav1.DeleteOptions{})
	})
	err = d.waitForDeploymentReady(deployment, deployments)
	assert.NoError(d.t(), err)
}

func (d *dddddDuffy) waitForDeploymentReady(
	deployment *appsv1.Deployment,
	deployments appsv1client.DeploymentInterface,
) error {
	return wait.Poll(250*time.Millisecond, 2*time.Minute, func() (done bool, err error) {
		depl, err := deployments.Get(d.ctx, deployment.GetName(), metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return depl.Status.Replicas == depl.Status.ReadyReplicas, nil
	})
}

func (d *duffy) watchTheWorldBurn() {
	// https://i.pinimg.com/originals/6d/59/3c/6d593c1be37b1e89089e93dfaccfaf37.gif
	d.upgradeCtx.Log.Infof(
		"Chaos duck unleashed. Watching the world burn for %v...", d.cfg.Timeout)
	time.Sleep(d.cfg.Timeout)
}
