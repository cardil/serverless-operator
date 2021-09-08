// +build upgrade

/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package upgrade_test

import (
	"testing"

	"github.com/openshift-knative/serverless-operator/test"
	"github.com/openshift-knative/serverless-operator/test/upgrade"
	"github.com/openshift-knative/serverless-operator/test/upgrade/installation"
	"go.uber.org/zap"
	eventingupgrade "knative.dev/eventing/test/upgrade"
	_ "knative.dev/pkg/system/testing"
	pkgupgrade "knative.dev/pkg/test/upgrade"
)

// FIXME: https://github.com/knative/eventing/issues/5176 `*-config.toml` in
//        this directory are required, so that kafkaupgrade tests will see them.

func TestServerlessUpgrade(t *testing.T) {
	ctx := test.SetupClusterAdmin(t)
	test.CleanupOnInterrupt(t, func() { test.CleanupAll(t, ctx) })
	cfg := newUpgradeConfig(t)
	suite := upgrade.Suite(ctx)
	suite.Execute(cfg)
}

func TestClusterUpgrade(t *testing.T) {
	ctx := test.SetupClusterAdmin(t)
	if !test.Flags.UpgradeOpenShift {
		t.Skip("Cluster upgrade tests disabled unless enabled by a flag.")
	}
	cfg := newUpgradeConfig(t)
	suite := upgrade.Suite(ctx)
	// Do not include continual tests as they're failing across cluster upgrades.
	suite.Tests.Continual = nil
	suite.Installations.UpgradeWith = []pkgupgrade.Operation{
		pkgupgrade.NewOperation("OpenShift Upgrade", func(c pkgupgrade.Context) {
			if err := installation.UpgradeOpenShift(ctx); err != nil {
				c.T.Error("OpenShift upgrade failed:", err)
			}
		}),
	}
	suite.Execute(cfg)
}

func newUpgradeConfig(t *testing.T) pkgupgrade.Configuration {
	log, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	return pkgupgrade.Configuration{T: t, Log: log}
}

func TestMain(m *testing.M) {
	eventingupgrade.RunMainTest(m)
}
