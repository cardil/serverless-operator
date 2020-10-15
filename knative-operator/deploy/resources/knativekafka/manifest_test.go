package knativekafka

import (
	"testing"

	mf "github.com/manifestival/manifestival"
)

var notAllowed = mf.Any(
	mf.All(mf.ByKind("ConfigMap"), mf.ByName("config-tracing")),
	mf.All(mf.ByKind("ConfigMap"), mf.ByName("config-observability")),
	mf.All(mf.ByKind("ConfigMap"), mf.ByName("config-logging")),
	mf.All(mf.ByKind("ConfigMap"), mf.ByName("config-leader-election-kafka")),
)

func TestUnallowedResourcesInManifest(t *testing.T) {
	tests := []struct {
		path  string
		fails bool
	}{{
		path:  "./kafkachannel-latest.yaml",
		fails: false,
	}, {
		path:  "./kafkasource-latest.yaml",
		fails: false,
	}, {
		path:  "./testdata/config-logging.yaml",
		fails: true,
	}, {
		path:  "./testdata/config-observability.yaml",
		fails: true,
	}, {
		path:  "./testdata/config-tracing.yaml",
		fails: true,
	}}

	for _, test := range tests {
		manifest, err := mf.ManifestFrom(mf.Path(test.path))
		if err != nil {
			t.Fatalf("Unable to load manifest at path '%s' for testing: %v", test.path, err)
		}
		manifest = manifest.Filter(notAllowed)
		if len(manifest.Resources()) > 0 && !test.fails {
			t.Fatalf("Manifest at path '%s' has unallowed resources", test.path)
		}
		if len(manifest.Resources()) == 0 && test.fails {
			t.Fatalf("Manifest at path '%s' should have unallowed resources, but it does not. Perhaps the check for unallowed resources is not working?", test.path)
		}
	}
}
