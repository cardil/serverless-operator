package chaos

import (
	"context"
	"testing"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	pkgupgrade "knative.dev/pkg/test/upgrade"
)

// Config holds configuration for the environment.
type Config struct {
	Timeout        time.Duration
	Wait           time.Duration
	Image          string
	Noop           bool
	Leaderelection LeaderelectionCfg
	KafkaSource    KafkaSourceCfg
}

// LeaderelectionCfg holds configuration for leaderelection chaosduck.
type LeaderelectionCfg struct {
	Namespaces []string
}

// KafkaSourceCfg holds configuration for Kafka source based chaosduck.
type KafkaSourceCfg struct {
	NamespacePrefixes []string
}

// NewConfigOfFail creates new configuration or fail doing so.
func NewConfigOfFail(ctx pkgupgrade.Context) Config {
	c, err := newConfig()
	if err != nil {
		ctx.T.Fatal(err)
	}
	return c
}

func newConfig() (Config, error) {
	defaults := Config{
		Timeout: 5 * time.Minute,
		Wait:    10 * time.Second,
		// TODO: replace image with actual resolvable one.
		Image: "quay.io/knative-e2e-test-images/chaosduck-7965e91b0d0f7fca30cc0eeb761e8bae",
		Noop:  false,
		Leaderelection: LeaderelectionCfg{
			Namespaces: []string{
				"knative-serving",
				"knative-serving-ingress",
				"knative-eventing",
			},
		},
		KafkaSource: KafkaSourceCfg{
			NamespacePrefixes: []string{
				"eventing-e2e",
			},
		},
	}
	err := envconfig.Process("SERVERLESS_CHAOSE2E", &defaults)
	if err != nil {
		return Config{}, err
	}
	return defaults, nil
}

type stateCtx struct {
	ctx context.Context
	tb  testing.TB
	log *zap.SugaredLogger
	cfg Config
}

func (s *stateCtx) named(name string) *stateCtx {
	return &stateCtx{
		ctx: s.ctx,
		tb:  s.tb,
		log: s.log.With("said", name),
		cfg: s.cfg,
	}
}
