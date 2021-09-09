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
	Timeout    time.Duration
	Wait       time.Duration
	Image      string
	Namespaces []string
	Noop       bool
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
		Timeout: 12 * time.Second,
		Wait:    5 * time.Second,
		Image:   "quay.io/knative-e2e-test-images/chaosduck-7965e91b0d0f7fca30cc0eeb761e8bae",
		Noop:    false,
		Namespaces: []string{
			"knative-serving",
			"knative-serving-ingress",
			"knative-eventing",
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
		log: s.log.With("name", name),
		cfg: s.cfg,
	}
}
