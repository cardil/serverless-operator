package chaos

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	pkgupgrade "knative.dev/pkg/test/upgrade"
)

// Config holds configuration for the environment.
type Config struct {
	Timeout    time.Duration
	Image      string
	Namespaces []string
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
		Timeout: 10 * time.Minute,
		Image:   "quay.io/knative-e2e-test-images/chaosduck-7965e91b0d0f7fca30cc0eeb761e8bae",
		Namespaces: []string{
			// "knative-serving",
			// "knative-serving-ingress",
			"knative-eventing",
		},
	}
	err := envconfig.Process("SERVERLESS_CHAOSE2E", &defaults)
	if err != nil {
		return Config{}, err
	}
	return defaults, nil
}
