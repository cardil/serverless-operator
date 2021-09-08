package chaose2e

import (
	"testing"

	"github.com/openshift-knative/serverless-operator/test"
	"go.uber.org/zap"
	kafkaupgrade "knative.dev/eventing-kafka/test/upgrade"
	"knative.dev/eventing-kafka/test/upgrade/continual"
	eventingupgrade "knative.dev/eventing/test/upgrade"
	pkgupgrade "knative.dev/pkg/test/upgrade"
	servingupgrade "knative.dev/serving/test/upgrade"
)

func TestChaos(t *testing.T) {
	ctx := test.SetupClusterAdmin(t)
	test.CleanupOnInterrupt(t, func() { test.CleanupAll(t, ctx) })
	cfg := newUpgradeConfig(t)
	suite := pkgupgrade.Suite{
		Tests: pkgupgrade.Tests{
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
				pkgupgrade.NewOperation("UnleashChaos", func(c pkgupgrade.Context) {

				}),
			},
		},
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
