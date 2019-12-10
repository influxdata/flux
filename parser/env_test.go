package parser_test

import (
	"os"
	"testing"

	"github.com/influxdata/flux/parser"
)

func TestFluxParserTypeEnvVar(t *testing.T) {
	src := `
package p

a = 10
`
	wantMeta := "parser-type=go"
	if os.Getenv("FLUX_PARSER_TYPE") == "rust" {
		wantMeta = "parser-type=rust"
	}
	pkg := parser.ParseSource(src)
	if want, got := wantMeta, pkg.Files[0].Metadata; want != got {
		t.Fatalf("wanted %q, got %q", want, got)
	}
}
