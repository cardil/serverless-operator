// +build chaos

package chaos_test

import (
	"testing"

	"github.com/openshift-knative/serverless-operator/test"
	"github.com/openshift-knative/serverless-operator/test/chaos"
	"github.com/openshift-knative/serverless-operator/test/upgrade"
	"go.uber.org/zap"
	kafkaupgrade "knative.dev/eventing-kafka/test/upgrade"
	"knative.dev/eventing-kafka/test/upgrade/continual"
	eventingupgrade "knative.dev/eventing/test/upgrade"
	pkgupgrade "knative.dev/pkg/test/upgrade"
)

func TestChaos(t *testing.T) {
	ctx := test.SetupClusterAdmin(t)
	test.CleanupOnInterrupt(t, func() { test.CleanupAll(t, ctx) })
	cfg := newUpgradeConfig(t)
	suite := upgrade.Suite(ctx)

	// TODO: use all continual tests, not only eventing ones.
	suite.Tests = pkgupgrade.Tests{
		Continual: test.Merge(
			[]pkgupgrade.BackgroundOperation{
				eventingupgrade.ContinualTest(),
			},
			kafkaupgrade.ChannelContinualTests(continual.ChannelTestOptions{}),
			kafkaupgrade.SourceContinualTests(continual.SourceTestOptions{}),
		),
	}
	suite.Installations.UpgradeWith = []pkgupgrade.Operation{
		pkgupgrade.NewOperation("UnleashChaosDuck", func(c pkgupgrade.Context) {
			h := chaos.NewHerder(c, chaos.NewConfigOfFail(c))
			h.LookAfter(chaos.NewRegularDuck())
			h.LookAfter(chaos.NewKafkaSourceDuck())
			h.Herd()
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
