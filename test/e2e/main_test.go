// +build e2e

package e2e

import (
	"flag"
	"os"
	"testing"

	"github.com/openshift-knative/serverless-operator/test"
)

var retcode = -1

func TestMain(m *testing.M) {
	// go1.13+ testing flags regression fix: https://github.com/golang/go/issues/31859
	_ = test.Flags
	flag.Parse()
	ctx := test.SetupClusterAdmin(nil)
	installServerlessOperator(ctx)
	defer undeployServerlessOperator(ctx)
	retcode = m.Run()

	os.Exit(retcode)
}

func installServerlessOperator(ctx *test.Context) {
	ctx.L.Log("Create subscription for Serverless Operator and wait for installation to succeed")
	if _, err := test.WithOperatorReady(ctx, "serverless-operator-subscription"); err != nil {
		ctx.L.Fatal("Failed to install serverless operator", err)
	}
	ctx.L.Log("Serverless Operator successfully installed")
}

func undeployServerlessOperator(ctx *test.Context) {
	ctx.L.Log("Removing Serverless Operator")
	ctx.Cleanup()
	if err := test.WaitForOperatorDepsDeleted(ctx); err != nil {
		ctx.L.Fatalf("Operators still running: %v, tests retcode was: %d", err, retcode)
	}
	ctx.L.Log("Serverless Operator removed successfully")
}
