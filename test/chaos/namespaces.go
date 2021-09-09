package chaos

import (
	"strings"

	"github.com/openshift-knative/serverless-operator/test"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func findNamespacesByPrefix(
	testCtx *test.Context,
	stateCtx *stateCtx,
	nsPrefixes ...string,
) []string {
	nss, err := testCtx.Clients.Kube.CoreV1().Namespaces().
		List(stateCtx.ctx, metav1.ListOptions{})
	assert.NilError(stateCtx.tb, err)
	found := make([]string, 0)
	for _, ns := range nss.Items {
		for _, prefix := range nsPrefixes {
			name := ns.GetName()
			if strings.HasPrefix(name, prefix) {
				found = append(found, name)
				break
			}
		}
	}
	return found
}
