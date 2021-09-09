package chaos

import (
	"sync"
	"time"

	"github.com/openshift-knative/serverless-operator/test"
	"gotest.tools/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
)

const chaosduck = "chaosduck"

// NewRegularDuck creates a regular chaos duck that target configured namespaces.
func NewRegularDuck() Duck {
	return &duffy{}
}

// see: https://p.kindpng.com/picc/s/190-1902027_daffy-duck-png-background-daffy-duck-et-bugs.png
type duffy struct {
	*stateCtx
	synchronized
}

func (d *duffy) takenCareOf(sctx *stateCtx, synch synchronized) Duck {
	return &duffy{
		stateCtx:     sctx.named("Duffy"),
		synchronized: synch,
	}
}

func (d *duffy) Quak() {
	t := d.tb
	testCtx := test.SetupClusterAdmin(t)
	var wg *sync.WaitGroup
	defer func() {
		testCtx.Cleanup(t)
		d.log.Info("Clean up done.")
		if wg != nil {
			d.log.Info("Done, sir.")
			wg.Done()
		}
	}()

	for _, namespace := range d.cfg.Leaderelection.Namespaces {
		d.quakOnNamespace(namespace, testCtx)
	}
	wg = d.watchTheWorldBurn()
}

func (d *duffy) quakOnNamespace(namespace string, testCtx *test.Context) {
	if d.cfg.Noop {
		d.log.Warnf("Skipping quaking on a namespace: %s", namespace)
		return
	}
	d.log.Infof("Unleashing duck on namespace: %s", namespace)
	sd := dddddDuffy{
		duffy:     d,
		testCtx:   testCtx,
		namespace: namespace,
	}
	sd.quak()
}

type dddddDuffy struct {
	*duffy
	namespace string
	testCtx   *test.Context
}

func (d *dddddDuffy) quak() {
	d.serviceAccount()
	d.role()
	d.roleBinding()
	d.deployment()
}

func (d *dddddDuffy) serviceAccount() {
	d.log.Infof("Creating service account: %s", d.namespace)
	k := d.testCtx.Clients.Kube
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: chaosduck},
	}
	serviceAccounts := k.CoreV1().ServiceAccounts(d.namespace)
	_, err := serviceAccounts.Create(d.ctx, sa, metav1.CreateOptions{})
	assert.NilError(d.tb, err)
	d.testCtx.AddToCleanup(func() error {
		d.log.Infof("Removing service account: %s", d.namespace)
		return serviceAccounts.Delete(d.ctx, sa.GetName(), metav1.DeleteOptions{})
	})
}

func (d *dddddDuffy) role() {
	d.log.Infof("Creating role: %s", d.namespace)
	k := d.testCtx.Clients.Kube
	roles := k.RbacV1().Roles(d.namespace)
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: chaosduck},
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
	assert.NilError(d.tb, err)
	d.testCtx.AddToCleanup(func() error {
		d.log.Infof("Removing role: %s", d.namespace)
		return roles.Delete(d.ctx, role.GetName(), metav1.DeleteOptions{})
	})
}

func (d *dddddDuffy) roleBinding() {
	d.log.Infof("Creating role binding: %s", d.namespace)
	k := d.testCtx.Clients.Kube
	roleBindings := k.RbacV1().RoleBindings(d.namespace)
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: chaosduck},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     chaosduck,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      chaosduck,
			Namespace: d.namespace,
		}},
	}
	_, err := roleBindings.Create(d.ctx, rb, metav1.CreateOptions{})
	assert.NilError(d.tb, err)
	d.testCtx.AddToCleanup(func() error {
		d.log.Infof("Removing role binding: %s", d.namespace)
		return roleBindings.Delete(d.ctx, rb.GetName(), metav1.DeleteOptions{})
	})
}

func (d *dddddDuffy) deployment() {
	d.log.Infof("Creating deployment: %s", d.namespace)
	k := d.testCtx.Clients.Kube
	var falseVal = false
	deployments := k.AppsV1().Deployments(d.namespace)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: chaosduck},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": chaosduck},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": chaosduck},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: chaosduck,
					Containers: []corev1.Container{{
						Name: chaosduck,
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
	assert.NilError(d.tb, err)
	d.testCtx.AddToCleanup(func() error {
		d.log.Infof("Removing deployment: %s", d.namespace)
		return deployments.Delete(d.ctx, deployment.GetName(), metav1.DeleteOptions{})
	})
	err = d.waitForDeploymentReady(deployment, deployments)
	assert.NilError(d.tb, err)
}

func (d *dddddDuffy) waitForDeploymentReady(
	deployment *appsv1.Deployment,
	deployments appsv1client.DeploymentInterface,
) error {
	d.log.Infof("Waiting for deployment ready: %s", d.namespace)
	return wait.Poll(250*time.Millisecond, 2*time.Minute, func() (done bool, err error) {
		depl, err := deployments.Get(d.ctx, deployment.GetName(), metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return depl.Status.Replicas == depl.Status.ReadyReplicas, nil
	})
}

func (d *duffy) watchTheWorldBurn() *sync.WaitGroup {
	// https://i.pinimg.com/originals/6d/59/3c/6d593c1be37b1e89089e93dfaccfaf37.gif
	d.log.Info("Chaos duck unleashed. Watching the world burn...")
	d.ready().Done()
	wg := <-d.finished()
	d.log.Info("Stopping chaos duck")
	return wg
}
