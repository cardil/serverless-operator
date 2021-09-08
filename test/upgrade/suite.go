package upgrade

import (
	"github.com/openshift-knative/serverless-operator/test"
	"github.com/openshift-knative/serverless-operator/test/upgrade/installation"
	kafkaupgrade "knative.dev/eventing-kafka/test/upgrade"
	"knative.dev/eventing-kafka/test/upgrade/continual"
	eventingupgrade "knative.dev/eventing/test/upgrade"
	pkgupgrade "knative.dev/pkg/test/upgrade"
	servingupgrade "knative.dev/serving/test/upgrade"
)

// Suite defines the suite of tests that are run for upgrade is tested.
func Suite(ctx *test.Context) pkgupgrade.Suite {
	return pkgupgrade.Suite{
		Tests: pkgupgrade.Tests{
			PreUpgrade:  preUpgradeTests(),
			PostUpgrade: postUpgradeTests(ctx),
			Continual: test.Merge(
				[]pkgupgrade.BackgroundOperation{
					servingupgrade.ProbeTest(),
					servingupgrade.AutoscaleSustainingWithTBCTest(),
					servingupgrade.AutoscaleSustainingTest(),
					eventingupgrade.ContinualTest(),
				},
				kafkaupgrade.ChannelContinualTests(continual.ChannelTestOptions{}),
				kafkaupgrade.SourceContinualTests(continual.SourceTestOptions{}),
			),
		},
		Installations: pkgupgrade.Installations{
			UpgradeWith: []pkgupgrade.Operation{
				pkgupgrade.NewOperation("UpgradeServerless", func(c pkgupgrade.Context) {
					if err := installation.UpgradeServerless(ctx); err != nil {
						c.T.Error("Serverless upgrade failed:", err)
					}
				}),
			},
		},
	}
}

func preUpgradeTests() []pkgupgrade.Operation {
	tests := []pkgupgrade.Operation{
		eventingupgrade.PreUpgradeTest(),
		kafkaupgrade.ChannelPreUpgradeTest(),
		kafkaupgrade.SourcePreUpgradeTest(),
	}
	// We might want to skip pre-upgrade test if we want to re-use the services
	// from the previous run. For example, to let them survive both Serverless
	// and OCP upgrades. This allows for more variants of tests, with different
	// order of upgrades.
	if test.Flags.SkipServingPreUpgrade {
		return tests
	}
	return append(tests, servingupgrade.ServingPreUpgradeTests()...)
}

func postUpgradeTests(ctx *test.Context) []pkgupgrade.Operation {
	tests := []pkgupgrade.Operation{
		waitForServicesReady(ctx),
		eventingupgrade.PostUpgradeTest(),
		kafkaupgrade.ChannelPostUpgradeTest(),
		kafkaupgrade.SourcePostUpgradeTest(),
	}
	tests = append(tests, servingupgrade.ServingPostUpgradeTests()...)
	return tests
}

func waitForServicesReady(ctx *test.Context) pkgupgrade.Operation {
	return pkgupgrade.NewOperation("WaitForServicesReady", func(c pkgupgrade.Context) {
		if err := test.WaitForReadyServices(ctx, "serving-tests"); err != nil {
			c.T.Error("Knative services not ready: ", err)
		}
		// TODO: Check if we need to sleep 30 more seconds like in the previous bash scripts.
	})
}
