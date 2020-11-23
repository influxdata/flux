package embed_test

import (
	"strings"
	"testing"

	"github.com/influxdata/flux/stdlib/embed"
)

// Verify that we can reference a package by name.
func TestAssetDir(t *testing.T) {
	names, err := embed.AssetDir("universe")
	if err != nil {
		t.Fatal(err)
	}

	if len(names) == 0 {
		t.Fatal("no flux files were found in the \"universe\" stdlib directory")
	}
}

// Only flux files should be embedded.
func TestOnlyFluxFiles(t *testing.T) {
	for _, name := range embed.AssetNames() {
		if !strings.HasSuffix(name, ".flux") {
			t.Errorf("unexpected non-flux file: %s", name)
		}
	}
}
